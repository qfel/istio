// Copyright 2019 Istio Authors
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in conformance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package conformance

import (
	"fmt"
	"net/http"
	"strings"
	"testing"

	"istio.io/istio/pkg/test"
	"istio.io/istio/pkg/test/conformance"
	"istio.io/istio/pkg/test/conformance/constraint"
	"istio.io/istio/pkg/test/framework"
	"istio.io/istio/pkg/test/framework/components/echo"
	"istio.io/istio/pkg/test/framework/components/echo/echoboot"
	"istio.io/istio/pkg/test/framework/components/environment"
	"istio.io/istio/pkg/test/framework/components/galley"
	"istio.io/istio/pkg/test/framework/components/namespace"
	"istio.io/istio/pkg/test/framework/components/pilot"
	"istio.io/istio/pkg/test/framework/label"
	"istio.io/istio/pkg/test/util/structpath"

	envoyAdmin "github.com/envoyproxy/go-control-plane/envoy/admin/v2alpha"
	"github.com/go-yaml/yaml"
)

func TestConformance(t *testing.T) {
	framework.Run(t, func(ctx framework.TestContext) {
		cases, err := loadCases()
		if err != nil {
			ctx.Fatalf("error loading test cases: %v", err)
		}

		gal := galley.NewOrFail(ctx, ctx, galley.Config{})
		p := pilot.NewOrFail(ctx, ctx, pilot.Config{Galley: gal})

		for _, ca := range cases {
			tst := ctx.NewSubTest(ca.Metadata.Name)

			for _, lname := range ca.Metadata.Labels {
				l, ok := label.Find(lname)
				if !ok {
					ctx.Fatalf("label not found: %v", lname)
				}
				tst = tst.Label(l)
			}

			if ca.Metadata.Isolated {
				tst.Run(runCaseFn(p, gal, ca))
			} else {
				tst.RunParallel(runCaseFn(p, gal, ca))
			}
		}
	})
}

func runCaseFn(p pilot.Instance, gal galley.Instance, ca *conformance.Test) func(framework.TestContext) {
	return func(ctx framework.TestContext) {
		match := true
	mainloop:
		for _, ename := range ca.Metadata.Environments {
			match = false
			for _, n := range environment.Names() {
				if n.String() == ename && n == ctx.Environment().EnvironmentName() {
					match = true
					break mainloop
				}
			}
		}

		if !match {
			ctx.Skipf("None of the expected environment(s) not found: %v", ca.Metadata.Environments)
		}

		if ca.Metadata.Skip {
			ctx.Skipf("Test is marked as skip")
		}

		// If there are any changes to the mesh config, then capture the original and restore.
		for _, s := range ca.Stages {
			if s.MeshConfig != nil {
				originalMeshCfg := gal.GetMeshConfigOrFail(ctx)
				defer gal.SetMeshConfigOrFail(ctx, originalMeshCfg)
				break
			}
		}

		ns := namespace.NewOrFail(ctx, ctx, "conf", true)

		if len(ca.Stages) == 1 {
			runStage(ctx, p, gal, ns, ca.Stages[0])
		} else {
			for i, s := range ca.Stages {
				ctx.NewSubTest(fmt.Sprintf("%d", i)).Run(func(ctx framework.TestContext) {
					runStage(ctx, p, gal, ns, s)
				})
			}
		}
	}
}

func runStage(ctx framework.TestContext, pil pilot.Instance, gal galley.Instance, ns namespace.Instance, s *conformance.Stage) {
	if s.MeshConfig != nil {
		gal.SetMeshConfigOrFail(ctx, *s.MeshConfig)
	}

	i := s.Input
	gal.ApplyConfigOrFail(ctx, ns, i)
	defer func() {
		gal.DeleteConfigOrFail(ctx, ns, i)
	}()

	if s.MCP != nil {
		validateMCPState(ctx, gal, ns, s)
	}
	if s.Traffic != nil {
		validateTraffic(ctx, pil, gal, ns, s)
	}

	// More and different types of validations can go here
}

func validateMCPState(ctx test.Failer, gal galley.Instance, ns namespace.Instance, s *conformance.Stage) {
	p := constraint.Params{
		Namespace: ns.Name(),
	}
	for _, coll := range s.MCP.Constraints {
		gal.WaitForSnapshotOrFail(ctx, coll.Name, func(actuals []*galley.SnapshotObject) error {
			for _, rangeCheck := range coll.Check {
				a := make([]interface{}, len(actuals))
				for i, item := range actuals {
					a[i] = item
					// Clear out for stable comparison.
					item.Metadata.CreateTime = nil
					item.Metadata.Version = ""
					if item.Metadata.Annotations != nil {
						delete(item.Metadata.Annotations, "kubectl.kubernetes.io/last-applied-configuration")
						if len(item.Metadata.Annotations) == 0 {
							item.Metadata.Annotations = nil
						}
					}
				}

				if err := rangeCheck.ValidateItems(a, p); err != nil {
					return err
				}
			}
			return nil
		})
	}

}

func DomainAcceptFunc(domains []string) func(*envoyAdmin.ConfigDump) (bool, error) {
	return func(cfg *envoyAdmin.ConfigDump) (bool, error) {
		validator := structpath.ForProto(cfg)
		// virtualHosts.domains
		const q = "{.configs[*].dynamicRouteConfigs[*].routeConfig.virtualHosts[*].domains[?(@ == %q)]}"
		for _, domain := range domains {
			// TODO(qfel): Figure out how to get rid of the loop.
			if err := validator.Exists(q, domain).Check(); err != nil {
				return false, err
			}
		}
		return true, nil
	}
}

func validateTraffic(ctx framework.TestContext, pil pilot.Instance, gal galley.Instance, ns namespace.Instance, stage *conformance.Stage) {
	echos := make([]echo.Instance, len(stage.Traffic.Services))
	b := echoboot.NewBuilderOrFail(ctx, ctx)
	for i, svc := range stage.Traffic.Services {
		b = b.With(&echos[i], echo.Config{
			Galley:    gal,
			Pilot:     pil,
			Service:   svc.Name,
			Namespace: ns,
			Ports:     svc.Ports,
		})
	}
	if err := b.Build(); err != nil {
		ctx.Fatal(err)
	}

	services := make(map[string]echo.Instance)
	for i, svc := range echos {
		services[stage.Traffic.Services[i].Name] = svc
		svc.WaitUntilCallableOrFail(ctx, echos...)
	}

	ready := make(map[string]bool)
	var vHosts []string

	var inputs []map[interface{}]interface{}
	for _, inputYAML := range strings.Split(stage.Input, "\n---\n") {
		var input map[interface{}]interface{}
		if err := yaml.Unmarshal([]byte(inputYAML), &input); err != nil {
			ctx.Fatal(err)
		}
		inputs = append(inputs, input)
	}
	for _, res := range inputs {
		if res["apiVersion"] != "networking.istio.io/v1alpha3" || res["kind"] != "VirtualService" {
			continue
		}
		spec := res["spec"].(map[interface{}]interface{})
		hosts := spec["hosts"].([]interface{})
		for _, h := range hosts {
			vHosts = append(vHosts, h.(string))
		}
	}

	for _, call := range stage.Traffic.Calls {
		caller := services[call.Caller]
		if !ready[call.Caller] {
			ctx.Logf("Waiting for sidecar(s) for %s to contain domains: %s", call.Caller, strings.Join(vHosts, ", "))
			for _, w := range caller.WorkloadsOrFail(ctx) {
				w.Sidecar().WaitForConfigOrFail(ctx, DomainAcceptFunc(vHosts))
			}
			ready[call.Caller] = true
		}

		resp := caller.CallOrFail(ctx, echo.CallOptions{
			Target:   services[call.Callee],
			PortName: "http",
			Host:     call.Host,
			Path:     call.Path,
			Count:    call.Count,
			Headers: http.Header{
				"Host": []string{call.Host},
			},
		})
		// TODO(qfel): Check constraints.
		for _, r := range resp {
			ctx.Logf(r.Body)
		}
	}
}
