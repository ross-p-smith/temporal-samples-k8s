SHELL := /bin/bash

.PHONY: help
.DEFAULT_GOAL := help

help: ## 💬 This help message :)
	@grep -E '[a-zA-Z_-]+:.*?## .*$$' $(firstword $(MAKEFILE_LIST)) | awk 'BEGIN {FS = ":.*?## "}; {printf "\033[36m%-20s\033[0m %s\n", $$1, $$2}'

local-cluster: ## 🐳 Create a local kind cluster
	@echo -e "\e[34m$@\e[0m" || true
	@./kind/setup_local_cluster.sh

local-temporal: local-cluster push-local ## 🐳 Create a local kind cluster
	@echo -e "\e[34m$@\e[0m" || true
	@kubectl create namespace temporal-samples-k8s || true
	@kubectl apply -n temporal-samples-k8s -R -f ./temporal/worker/manifests
	@kubectl get all -n temporal-samples-k8s

expose-ui: ## 🌐 Expose Temporal Web UI
	@echo -e "\e[34m$@\e[0m" || true
	@kubectl port-forward -n temporal services/temporal-ui 8088:8088

expose-worker: ## 🌐 Expose Temporal Web UI
	@echo -e "\e[34m$@\e[0m" || true
	@kubectl port-forward -n temporal services/temporal-worker-app 8090:8090

build: ## 🔨 Build the worker
	@echo -e "\e[34m$@\e[0m" || true
	@docker build temporal/worker -t localhost:5001/temporal-worker-app:v3

push-local: build ## 🚚  Pushing docker image to local registry
	@echo -e "\e[34m$@\e[0m" || true
	@docker push localhost:5001/temporal-worker-app:v3

load:
	@echo -e "\e[34m$@\e[0m" || true
	@./load.sh

openssl-certs: ## 🔐 Generate the Certificates using openssl
	@echo -e "\e[34m$@\e[0m" || true
	@cd certs && ./generate-test-certs-openssl.sh

keda:
	@echo -e "\e[34m$@\e[0m" || true
	@helm repo add kedacore https://kedacore.github.io/charts
	@helm repo update
	@kubectl create namespace keda || true
	@helm install keda kedacore/keda --namespace keda

clean: ## 🧹 Clean up
	@echo -e "\e[34m$@\e[0m" || true
	@kubectl --namespace temporal-samples-k8s delete secret ca-cert || true
	@kubectl --namespace temporal-samples-k8s delete secret cluster-cert || true
	@kubectl --namespace temporal-samples-k8s delete secret client-cert || true
	@kubectl delete all --all -n temporal-samples-k8s --wait
	@kind delete cluster --name temporal-aks