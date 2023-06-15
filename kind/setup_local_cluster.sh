#!/bin/sh
set -e pipefail

reg_name='kind-registry'
reg_port='5001'

if [ "$(docker inspect -f '{{.State.Running}}' "${reg_name}" 2>/dev/null || true)" != 'true' ]; then
  printf '\n📀 Create registry container\n\n'
  docker run \
    -d --restart=always -p "127.0.0.1:${reg_port}:5000" --name "${reg_name}" \
    registry:2
fi

cluster_exists=$(kind get clusters | grep temporal-k8s || true)
if [ -z "$cluster_exists" ]; then
  printf '\n📀 Creating kind cluster called: temporal-k8s\n\n'
  kind create cluster --name temporal-k8s --config ./kind/kind-cluster-config.yaml
fi

if [ "$(docker inspect -f='{{json .NetworkSettings.Networks.kind}}' "${reg_name}")" = 'null' ]; then
  printf '\n📀 Connect the registry to the cluster network\n'
  docker network connect "kind" "${reg_name}"
fi

printf '\n📀 Map the local registry to cluster\n\n'
kubectl apply -f ./kind/config-map.yaml --wait=true

# Get the external IP address of the temporal-worker-app from kubectl

# kubectl get svc temporal-worker-app -o jsonpath='{.status.loadBalancer.ingress[0].ip}'