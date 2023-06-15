#!/bin/sh
set -e pipefail

# if ./certs/certs/ does not exist, stop
if [ ! -d "/workspaces/temporal-samples-k8s/certs/certs" ]; then
    echo "Please run 'make openssl-certs'."
    exit 1
fi

# This feels nicer to use, but you cannot change the name of the files
# and we have 3 certs and they would overide each other if they are all
# called cert.crt and cert.key
# kubectl --namespace temporal-samples-k8s create secret tls ca \
#     --cert=/workspaces/temporal-samples-k8s/certs/certs/ca.cert \
#     --key=/workspaces/temporal-samples-k8s/certs/certs/ca.key \
#     || true


kubectl --namespace temporal-samples-k8s create secret generic ca-cert \
    --from-file=/workspaces/temporal-samples-k8s/certs/certs/ca.cert \
    --from-file=/workspaces/temporal-samples-k8s/certs/certs/ca.key \
    || true

kubectl --namespace temporal-samples-k8s create secret generic cluster-cert \
    --from-file=/workspaces/temporal-samples-k8s/certs/certs/cluster.cert \
    --from-file=/workspaces/temporal-samples-k8s/certs/certs/cluster.key \
    || true

kubectl --namespace temporal-samples-k8s create secret generic client-cert \
    --from-file=/workspaces/temporal-samples-k8s/certs/certs/client.cert \
    --from-file=/workspaces/temporal-samples-k8s/certs/certs/client.key \
    || true