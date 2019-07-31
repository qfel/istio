// Copyright 2019 Istio Authors
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package common_test

import (
	"bytes"
	"context"
	"fmt"
	"io/ioutil"
	"testing"

	envoyAdmin "github.com/envoyproxy/go-control-plane/envoy/admin/v2alpha"
	"github.com/gogo/protobuf/jsonpb"

	"istio.io/istio/pkg/config/protocol"
	"istio.io/istio/pkg/test"
	"istio.io/istio/pkg/test/echo/client"
	"istio.io/istio/pkg/test/echo/proto"
	"istio.io/istio/pkg/test/framework/components/echo"
	"istio.io/istio/pkg/test/framework/components/echo/common"
	"istio.io/istio/pkg/test/framework/resource"
	"istio.io/istio/pkg/test/util/structpath"
)

func TestCheckOutboundConfig(t *testing.T) {
	configDump, err := ioutil.ReadFile("testdata/config_dump.json")
	if err != nil {
		t.Fatal(err)
	}

	cfg := &envoyAdmin.ConfigDump{}
	if err := jsonpb.Unmarshal(bytes.NewReader(configDump), cfg); err != nil {
		t.Fatal(err)
	}

	cfgs := []testConfig{
		{
			protocol:    protocol.HTTP,
			service:     "b",
			namespace:   "apps-1-99281",
			domain:      "cluster.local",
			servicePort: 80,
			address:     "10.43.241.185",
		},
		{
			protocol:    protocol.HTTP,
			service:     "b",
			namespace:   "apps-1-99281",
			domain:      "cluster.local",
			servicePort: 8080,
			address:     "10.43.241.185",
		},
		{
			protocol:    protocol.TCP,
			service:     "b",
			namespace:   "apps-1-99281",
			domain:      "cluster.local",
			servicePort: 90,
			address:     "10.43.241.185",
		},
		{
			protocol:    protocol.HTTPS,
			service:     "b",
			namespace:   "apps-1-99281",
			domain:      "cluster.local",
			servicePort: 9090,
			address:     "10.43.241.185",
		},
		{
			protocol:    protocol.HTTP2,
			service:     "b",
			namespace:   "apps-1-99281",
			domain:      "cluster.local",
			servicePort: 70,
			address:     "10.43.241.185",
		},
		{
			protocol:    protocol.GRPC,
			service:     "b",
			namespace:   "apps-1-99281",
			domain:      "cluster.local",
			servicePort: 7070,
			address:     "10.43.241.185",
		},
	}

	validator := structpath.ForProto(cfg)

	for _, cfg := range cfgs {
		t.Run(fmt.Sprintf("%s_%d[%s]", cfg.service, cfg.servicePort, cfg.protocol), func(t *testing.T) {
			if err := common.CheckOutboundConfig(&cfg, cfg.Config().Ports[0], validator); err != nil {
				t.Fatal(err)
			}
		})
	}
}

var _ echo.Instance = &testConfig{}
var _ echo.Workload = &testConfig{}

type testConfig struct {
	protocol    protocol.Instance
	servicePort int
	address     string
	service     string
	domain      string
	namespace   string
}

func (e *testConfig) Owner() echo.Instance {
	return e
}

func (e *testConfig) Port() echo.Port {
	return echo.Port{
		ServicePort: e.servicePort,
		Protocol:    e.protocol,
	}
}

func (e *testConfig) Address() string {
	return e.address
}

func (e *testConfig) Config() echo.Config {
	return echo.Config{
		Service: e.service,
		Namespace: &fakeNamespace{
			name: e.namespace,
		},
		Domain: e.domain,
		Ports: []echo.Port{
			{
				ServicePort: e.servicePort,
				Protocol:    e.protocol,
			},
		},
	}
}

func (e *testConfig) Workloads() ([]echo.Workload, error) {
	return []echo.Workload{e}, nil
}

func (*testConfig) ID() resource.ID {
	panic("not implemented")
}

func (*testConfig) WorkloadsOrFail(t test.Failer) []echo.Workload {
	panic("not implemented")
}

func (*testConfig) WaitUntilCallable(_ ...echo.Instance) error {
	panic("not implemented")
}

func (*testConfig) WaitUntilCallableOrFail(_ test.Failer, _ ...echo.Instance) {
	panic("not implemented")
}

func (*testConfig) Call(_ echo.CallOptions) (client.ParsedResponses, error) {
	panic("not implemented")
}

func (*testConfig) CallOrFail(_ test.Failer, _ echo.CallOptions) client.ParsedResponses {
	panic("not implemented")
}

func (*testConfig) Sidecar() echo.Sidecar {
	panic("not implemented")
}

func (*testConfig) ForwardEcho(context.Context, *proto.ForwardEchoRequest) (client.ParsedResponses, error) {
	panic("not implemented")
}

type fakeNamespace struct {
	name string
}

func (n *fakeNamespace) Name() string {
	return n.name
}

func (n *fakeNamespace) ID() resource.ID {
	panic("not implemented")
}

const js = `{
             "configs": [
              {
               "@type": "type.googleapis.com/envoy.admin.v2alpha.BootstrapConfigDump",
               "bootstrap": {
                "node": {
                 "id": "sidecar~10.48.0.143~echo-v1-97565478c-dnsc8.conf-1-14216~conf-1-14216.svc.cluster.local",
                 "cluster": "echo.conf-1-14216",
                 "metadata": {
                  "ISTIO_PROXY_SHA": "istio-proxy:cb503fee7392ff2f6a7b19846c3343122dea72dc",
                  "app": "echo",
                  "INSTANCE_IPS": "10.48.0.143",
                  "pod-template-hash": "97565478c",
                  "TRAFFICDIRECTOR_INTERCEPTION_PORT": "15001",
                  "istio.io/metadata": {
                   "namespace": "conf-1-14216",
                   "platform_metadata": {
                    "gcp_cluster_location": "us-west1-pj1",
                    "gcp_cluster_name": "k8s-staging-west",
                    "gcp_project": "qfel-dev"
                   },
                   "labels": {
                    "pod-template-hash": "97565478c",
                    "app": "echo",
                    "version": "v1"
                   },
                   "ip": "10.48.0.143",
                   "name": "echo-v1-97565478c-dnsc8"
                  },
                  "INTERCEPTION_MODE": "REDIRECT",
                  "CONFIG_NAMESPACE": "conf-1-14216",
                  "version": "v1",
                  "ISTIO_VERSION": "master-20190730-09-16",
                  "POD_NAME": "echo-v1-97565478c-dnsc8",
                  "foo": "bar",
                  "istio": "sidecar",
                  "ISTIO_PROXY_VERSION": "1.1.3"
                 },
                 "locality": {
                  "region": "us-west1",
                  "zone": "us-west1-pj1"
                 },
                 "build_version": "cb503fee7392ff2f6a7b19846c3343122dea72dc/1.11.0-dev/Clean/RELEASE/BoringSSL"
                },
                "dynamic_resources": {
                 "lds_config": {
                  "ads": {}
                 },
                 "cds_config": {
                  "ads": {}
                 },
                 "ads_config": {
                  "api_type": "GRPC",
                  "grpc_services": [
                   {
                    "google_grpc": {
                     "target_uri": "staging-trafficdirector.sandbox.googleapis.com:443",
                     "channel_credentials": {
                      "ssl_credentials": {
                       "root_certs": {
                        "filename": "/etc/ssl/certs/ca-certificates.crt"
                       }
                      }
                     },
                     "call_credentials": [
                      {
                       "google_compute_engine": {}
                      }
                     ],
                     "stat_prefix": "googlegrpcxds"
                    }
                   }
                  ]
                 }
                },
                "cluster_manager": {
                 "load_stats_config": {
                  "api_type": "GRPC",
                  "grpc_services": [
                   {
                    "google_grpc": {
                     "target_uri": "staging-trafficdirector.sandbox.googleapis.com:443",
                     "channel_credentials": {
                      "ssl_credentials": {
                       "root_certs": {
                        "filename": "/etc/ssl/certs/ca-certificates.crt"
                       }
                      }
                     },
                     "call_credentials": [
                      {
                       "google_compute_engine": {}
                      }
                     ],
                     "stat_prefix": "googlegrpcxds"
                    }
                   }
                  ]
                 }
                },
                "admin": {
                 "access_log_path": "/dev/null",
                 "address": {
                  "socket_address": {
                   "address": "127.0.0.1",
                   "port_value": 15000
                  }
                 }
                }
               },
               "last_updated": "2019-08-01T19:03:14.760Z"
              },
              {
               "@type": "type.googleapis.com/envoy.admin.v2alpha.ClustersConfigDump",
               "version_info": "1564686465324726666",
               "dynamic_warming_clusters": [
                {
                 "version_info": "1564686465324726666",
                 "cluster": {
                  "name": "cloud-internal-istio-staging:cloud_mp_987637929440_2779893103235860842",
                  "type": "EDS",
                  "eds_cluster_config": {
                   "eds_config": {
                    "ads": {}
                   }
                  },
                  "connect_timeout": "30s",
                  "circuit_breakers": {
                   "thresholds": [
                    {
                     "max_connections": 65536
                    }
                   ]
                  },
                  "http_protocol_options": {},
                  "outlier_detection": {
                   "interval": "1s",
                   "max_ejection_percent": 50,
                   "enforcing_consecutive_5xx": 0,
                   "consecutive_gateway_failure": 3,
                   "enforcing_consecutive_gateway_failure": 100
                  },
                  "metadata": {
                   "filter_metadata": {
                    "com.google.trafficdirector": {
                     "backend_service_name": "csm-echo1-conf-1-14216-svc-cluster-local-8080"
                    }
                   }
                  },
                  "common_lb_config": {
                   "locality_weighted_lb_config": {}
                  }
                 },
                 "last_updated": "2019-08-01T19:08:11.377Z"
                },
                {
                 "version_info": "1564686465324726666",
                 "cluster": {
                  "name": "cloud-internal-istio-staging:cloud_mp_987637929440_3485620899291236715",
                  "type": "EDS",
                  "eds_cluster_config": {
                   "eds_config": {
                    "ads": {}
                   }
                  },
                  "connect_timeout": "30s",
                  "circuit_breakers": {
                   "thresholds": [
                    {
                     "max_connections": 65536
                    }
                   ]
                  },
                  "http_protocol_options": {},
                  "outlier_detection": {
                   "interval": "1s",
                   "max_ejection_percent": 50,
                   "enforcing_consecutive_5xx": 0,
                   "consecutive_gateway_failure": 3,
                   "enforcing_consecutive_gateway_failure": 100
                  },
                  "metadata": {
                   "filter_metadata": {
                    "com.google.trafficdirector": {
                     "backend_service_name": "csm-echo-conf-1-14216-svc-cluster-local-8080"
                    }
                   }
                  },
                  "common_lb_config": {
                   "locality_weighted_lb_config": {}
                  }
                 },
                 "last_updated": "2019-08-01T19:08:11.379Z"
                },
                {
                 "version_info": "1564686465324726666",
                 "cluster": {
                  "name": "cloud-internal-istio-staging:cloud_mp_987637929440_4603802483614979971",
                  "type": "EDS",
                  "eds_cluster_config": {
                   "eds_config": {
                    "ads": {}
                   }
                  },
                  "connect_timeout": "30s",
                  "circuit_breakers": {
                   "thresholds": [
                    {
                     "max_connections": 65536
                    }
                   ]
                  },
                  "http_protocol_options": {},
                  "outlier_detection": {
                   "interval": "1s",
                   "max_ejection_percent": 50,
                   "enforcing_consecutive_5xx": 0,
                   "consecutive_gateway_failure": 3,
                   "enforcing_consecutive_gateway_failure": 100
                  },
                  "metadata": {
                   "filter_metadata": {
                    "com.google.trafficdirector": {
                     "backend_service_name": "serve404"
                    }
                   }
                  },
                  "common_lb_config": {
                   "locality_weighted_lb_config": {}
                  }
                 },
                 "last_updated": "2019-08-01T19:08:11.377Z"
                },
                {
                 "version_info": "1564686465324726666",
                 "cluster": {
                  "name": "cloud-internal-istio-staging:cloud_mp_987637929440_6545916046949124420",
                  "type": "EDS",
                  "eds_cluster_config": {
                   "eds_config": {
                    "ads": {}
                   }
                  },
                  "connect_timeout": "30s",
                  "circuit_breakers": {
                   "thresholds": [
                    {
                     "max_connections": 65536
                    }
                   ]
                  },
                  "http_protocol_options": {},
                  "outlier_detection": {
                   "interval": "1s",
                   "max_ejection_percent": 50,
                   "enforcing_consecutive_5xx": 0,
                   "consecutive_gateway_failure": 3,
                   "enforcing_consecutive_gateway_failure": 100
                  },
                  "metadata": {
                   "filter_metadata": {
                    "com.google.trafficdirector": {
                     "backend_service_name": "csm-echo2-conf-1-14216-svc-cluster-local-8080"
                    }
                   }
                  },
                  "common_lb_config": {
                   "locality_weighted_lb_config": {}
                  }
                 },
                 "last_updated": "2019-08-01T19:08:11.378Z"
                }
               ]
              },
              {
               "@type": "type.googleapis.com/envoy.admin.v2alpha.ListenersConfigDump",
               "version_info": "1564686465324726666",
               "dynamic_warming_listeners": [
                {
                 "version_info": "1564686465324726666",
                 "listener": {
                  "name": "TRAFFICDIRECTOR_INTERCEPTION_LISTENER",
                  "address": {
                   "socket_address": {
                    "address": "0.0.0.0",
                    "port_value": 15001
                   }
                  },
                  "filter_chains": [
                   {
                    "filter_chain_match": {
                     "prefix_ranges": [
                      {
                       "address_prefix": "10.51.240.86",
                       "prefix_len": 32
                      }
                     ],
                     "destination_port": 8080
                    },
                    "filters": [
                     {
                      "name": "envoy.http_connection_manager",
                      "typed_config": {
                       "@type": "type.googleapis.com/envoy.config.filter.network.http_connection_manager.v2.HttpConnectionManager",
                       "stat_prefix": "trafficdirector",
                       "rds": {
                        "config_source": {
                         "ads": {}
                        },
                        "route_config_name": "URL_MAP/987637929440.csm-mesh-url-map"
                       },
                       "http_filters": [
                        {
                         "name": "envoy.fault"
                        },
                        {
                         "name": "envoy.cors"
                        },
                        {
                         "name": "envoy.router",
                         "typed_config": {
                          "@type": "type.googleapis.com/envoy.config.filter.http.router.v2.Router",
                          "suppress_envoy_headers": true
                         }
                        }
                       ],
                       "use_remote_address": true,
                       "generate_request_id": false
                      }
                     }
                    ]
                   },
                   {
                    "filter_chain_match": {
                     "prefix_ranges": [
                      {
                       "address_prefix": "10.51.252.246",
                       "prefix_len": 32
                      }
                     ],
                     "destination_port": 8080
                    },
                    "filters": [
                     {
                      "name": "envoy.http_connection_manager",
                      "typed_config": {
                       "@type": "type.googleapis.com/envoy.config.filter.network.http_connection_manager.v2.HttpConnectionManager",
                       "stat_prefix": "trafficdirector",
                       "rds": {
                        "config_source": {
                         "ads": {}
                        },
                        "route_config_name": "URL_MAP/987637929440.csm-mesh-url-map"
                       },
                       "http_filters": [
                        {
                         "name": "envoy.fault"
                        },
                        {
                         "name": "envoy.cors"
                        },
                        {
                         "name": "envoy.router",
                         "typed_config": {
                          "@type": "type.googleapis.com/envoy.config.filter.http.router.v2.Router",
                          "suppress_envoy_headers": true
                         }
                        }
                       ],
                       "use_remote_address": true,
                       "generate_request_id": false
                      }
                     }
                    ]
                   },
                   {
                    "filter_chain_match": {
                     "prefix_ranges": [
                      {
                       "address_prefix": "10.51.242.216",
                       "prefix_len": 32
                      }
                     ],
                     "destination_port": 8080
                    },
                    "filters": [
                     {
                      "name": "envoy.http_connection_manager",
                      "typed_config": {
                       "@type": "type.googleapis.com/envoy.config.filter.network.http_connection_manager.v2.HttpConnectionManager",
                       "stat_prefix": "trafficdirector",
                       "rds": {
                        "config_source": {
                         "ads": {}
                        },
                        "route_config_name": "URL_MAP/987637929440.csm-mesh-url-map"
                       },
                       "http_filters": [
                        {
                         "name": "envoy.fault"
                        },
                        {
                         "name": "envoy.cors"
                        },
                        {
                         "name": "envoy.router",
                         "typed_config": {
                          "@type": "type.googleapis.com/envoy.config.filter.http.router.v2.Router",
                          "suppress_envoy_headers": true
                         }
                        }
                       ],
                       "use_remote_address": true,
                       "generate_request_id": false
                      }
                     }
                    ]
                   },
                   {
                    "filter_chain_match": {
                     "prefix_ranges": [
                      {
                       "address_prefix": "10.51.241.7",
                       "prefix_len": 32
                      }
                     ],
                     "destination_port": 8080
                    },
                    "filters": [
                     {
                      "name": "envoy.http_connection_manager",
                      "typed_config": {
                       "@type": "type.googleapis.com/envoy.config.filter.network.http_connection_manager.v2.HttpConnectionManager",
                       "stat_prefix": "trafficdirector",
                       "rds": {
                        "config_source": {
                         "ads": {}
                        },
                        "route_config_name": "URL_MAP/987637929440.csm-mesh-url-map"
                       },
                       "http_filters": [
                        {
                         "name": "envoy.fault"
                        },
                        {
                         "name": "envoy.cors"
                        },
                        {
                         "name": "envoy.router",
                         "typed_config": {
                          "@type": "type.googleapis.com/envoy.config.filter.http.router.v2.Router",
                          "suppress_envoy_headers": true
                         }
                        }
                       ],
                       "use_remote_address": true,
                       "generate_request_id": false
                      }
                     }
                    ]
                   }
                  ],
                  "listener_filters": [
                   {
                    "name": "envoy.listener.original_dst"
                   },
                   {
                    "name": "envoy.listener.tls_inspector"
                   }
                  ]
                 },
                 "last_updated": "2019-08-01T19:08:11.381Z"
                }
               ]
              },
              {
               "@type": "type.googleapis.com/envoy.admin.v2alpha.ScopedRoutesConfigDump"
              },
              {
               "@type": "type.googleapis.com/envoy.admin.v2alpha.RoutesConfigDump"
              }
             ]
            }`

func TestXXX(t *testing.T) {
	msg := &envoyAdmin.ConfigDump{}
	if err := common.UnmarshalScrubAny([]byte(js), msg); err != nil {
		t.Fatal(err)
	}
}
