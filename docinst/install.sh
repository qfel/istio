#!/bin/bash
set -ex
cd installer
export TAG=master-latest-daily
export HUB=gcr.io/istio-release

kubectl label namespace default istio-env=istio-control
kubectl apply -f crds/files
bin/iop istio-system citadel security/citadel
#bin/iop istio-cni istio-cni istio-cni
bin/iop istio-control istio-config istio-control/istio-config \
    --set configValidation=true
bin/iop istio-control istio-discovery istio-control/istio-discovery
bin/iop istio-control istio-autoinject istio-control/istio-autoinject
bin/iop istio-ingress istio-ingress gateways/istio-ingress
