#!/bin/sh
set -e pipefail

external_ip=$(kubectl get svc -n temporal-samples-k8s temporal-worker-app --output jsonpath='{.status.loadBalancer.ingress[0].ip}')

seq 1000 |  parallel -n0 -j4 "curl http://${external_ip}:8090/delay"

