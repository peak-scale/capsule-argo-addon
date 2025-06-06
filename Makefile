# Version
GIT_HEAD_COMMIT ?= $(shell git rev-parse --short HEAD)
VERSION         ?= $(or $(shell git describe --abbrev=0 --tags --match "v*" 2>/dev/null),$(GIT_HEAD_COMMIT))
GOOS            ?= $(shell go env GOOS)
GOARCH          ?= $(shell go env GOARCH)

# Defaults
REGISTRY        ?= ghcr.io
REPOSITORY      ?= peak-scale/capsule-argo-addon
GIT_TAG_COMMIT  ?= $(shell git rev-parse --short $(VERSION))
GIT_MODIFIED_1  ?= $(shell git diff $(GIT_HEAD_COMMIT) $(GIT_TAG_COMMIT) --quiet && echo "" || echo ".dev")
GIT_MODIFIED_2  ?= $(shell git diff --quiet && echo "" || echo ".dirty")
GIT_MODIFIED    ?= $(shell echo "$(GIT_MODIFIED_1)$(GIT_MODIFIED_2)")
GIT_REPO        ?= $(shell git config --get remote.origin.url)
BUILD_DATE      ?= $(shell git log -1 --format="%at" | xargs -I{} sh -c 'if [ "$(shell uname)" = "Darwin" ]; then date -r {} +%Y-%m-%dT%H:%M:%S; else date -d @{} +%Y-%m-%dT%H:%M:%S; fi')
IMG_BASE        ?= $(REPOSITORY)
IMG             ?= $(IMG_BASE):$(VERSION)
FULL_IMG        ?= $(REGISTRY)/$(IMG_BASE)

## Tool Binaries
KUBECTL ?= kubectl

# Get the currently used golang install path (in GOPATH/bin, unless GOBIN is set)
ifeq (,$(shell go env GOBIN))
GOBIN=$(shell go env GOPATH)/bin
else
GOBIN=$(shell go env GOBIN)
endif

####################
# -- Golang
####################

.PHONY: golint
golint: golangci-lint
	$(GOLANGCI_LINT) run -c .golangci.yml

all: manager

# Run tests
.PHONY: test
test: test-clean generate manifests test-clean
	@GO111MODULE=on go test -v $(shell go list ./... | grep -v "e2e") -coverprofile coverage.out

.PHONY: test-clean
test-clean: ## Clean tests cache
	@go clean -testcache

# Build manager binary
manager: generate golint
	go build -o bin/manager

# Run against the configured Kubernetes cluster in ~/.kube/config
run: generate manifests
	go run .

# Generate manifests e.g. CRD, RBAC etc.
manifests: controller-gen apidocs
	@$(CONTROLLER_GEN) crd paths="./..." output:crd:artifacts:config=charts/capsule-argo-addon/crds

# Generate code
generate: controller-gen
	@$(CONTROLLER_GEN) object:headerFile="hack/boilerplate.go.txt" paths="./..."

apidocs: TARGET_DIR      := $(shell mktemp -d)
apidocs: apidocs-gen generate
	@$(APIDOCS_GEN) crdoc --resources charts/capsule-argo-addon/crds --output docs/reference.md --template ./hack/templates/crds.tmpl

####################
# -- Docker
####################

KO_PLATFORM     ?= linux/$(GOARCH)
KOCACHE         ?= /tmp/ko-cache
KO_REGISTRY     := ko.local
KO_TAGS         ?= "latest"
ifdef VERSION
KO_TAGS         := $(KO_TAGS),$(VERSION)
endif

LD_FLAGS        := "-X main.Version=$(VERSION) \
					-X main.GitCommit=$(GIT_HEAD_COMMIT) \
					-X main.GitTag=$(VERSION) \
					-X main.GitTreeState=$(GIT_MODIFIED) \
					-X main.BuildDate=$(BUILD_DATE) \
					-X main.GitRepo=$(GIT_REPO)"

# Docker Image Build
# ------------------

.PHONY: ko-build-controller
ko-build-controller: ko
	@echo Building Controller $(FULL_IMG) - $(KO_TAGS) >&2
	@LD_FLAGS=$(LD_FLAGS) KOCACHE=$(KOCACHE) KO_DOCKER_REPO=$(FULL_IMG) \
		$(KO) build ./cmd/ --bare --tags=$(KO_TAGS) --push=false --local --platform=$(KO_PLATFORM)

.PHONY: ko-build-all
ko-build-all: ko-build-controller

# Docker Image Publish
# ------------------

REGISTRY_PASSWORD   ?= dummy
REGISTRY_USERNAME   ?= dummy

.PHONY: ko-login
ko-login: ko
	@$(KO) login $(REGISTRY) --username $(REGISTRY_USERNAME) --password $(REGISTRY_PASSWORD)

.PHONY: ko-publish-controller
ko-publish-controller: ko-login
	@echo Publishing Controller $(FULL_IMG) - $(KO_TAGS) >&2
	@LD_FLAGS=$(LD_FLAGS) KOCACHE=$(KOCACHE) KO_DOCKER_REPO=$(FULL_IMG) \
		$(KO) build ./cmd/ --bare --tags=$(KO_TAGS) --push=true

.PHONY: ko-publish-all
ko-publish-all: ko-publish-controller

####################
# -- Helm
####################

# Helm
SRC_ROOT = $(shell git rev-parse --show-toplevel)

helm-controller-version:
	$(eval VERSION := $(shell grep 'appVersion:' charts/capsule-argo-addon/Chart.yaml | awk '{print $$2}'))
	$(eval KO_TAGS := $(shell grep 'appVersion:' charts/capsule-argo-addon/Chart.yaml | awk '{print $$2}'))


helm-docs: helm-doc
	$(HELM_DOCS) --chart-search-root ./charts

helm-lint: ct
	@$(CT) lint --config .github/configs/ct.yaml --validate-yaml=false --all --debug

helm-schema: helm helm-plugin-schema
	cd charts/capsule-argo-addon && $(HELM) schema -output values.schema.json

helm-test: kind ct
	@$(KIND) create cluster --wait=60s --name helm-capsule-argo-addon
	@$(MAKE) e2e-install-distro
	@$(MAKE) helm-test-exec
	@$(KIND) delete cluster --name helm-capsule-argo-addon

helm-test-exec: ct helm-controller-version ko-build-all
	@$(KIND) load docker-image --name helm-capsule-argo-addon $(FULL_IMG):$(VERSION)
	@$(CT) install --config $(SRC_ROOT)/.github/configs/ct.yaml --all --debug

docker:
	@hash docker 2>/dev/null || {\
		echo "You need docker" &&\
		exit 1;\
	}

####################
# -- Install E2E Tools
####################
K3S_CLUSTER ?= "capsule-argo-addon"

e2e: e2e-build e2e-exec e2e-destroy

e2e-build: kind
	$(KIND) create cluster --wait=60s --name $(K3S_CLUSTER) --config ./e2e/kind.yaml --image=kindest/node:$${KIND_K8S_VERSION:-v1.30.0}
	$(MAKE) e2e-install

e2e-exec: ginkgo
	$(GINKGO) -r -vv ./e2e

e2e-destroy: kind
	$(KIND) delete cluster --name $(K3S_CLUSTER)

e2e-install: e2e-install-distro e2e-install-addon

.PHONY: e2e-install
e2e-install-addon: helm e2e-load-image
	$(HELM) upgrade \
	    --dependency-update \
		--debug \
		--install \
		--namespace capsule-argo-addon \
		--create-namespace \
		--set 'image.pullPolicy=Never' \
		--set "image.tag=$(VERSION)" \
		--set certManager.certificate.dnsNames={localhost} \
		--set proxy.enabled=true \
		--set proxy.crds.install=true \
        --set certManager.certificate.dnsNames={localhost} \
		--set webhooks.enabled=true \
		--set args.logLevel=10 \
		capsule-argo-addon \
		./charts/capsule-argo-addon

e2e-install-distro:
	@$(KUBECTL) kustomize e2e/objects/flux/ | kubectl apply -f -
	@$(KUBECTL) kustomize e2e/objects/distro/ | kubectl apply -f -
	@$(MAKE) wait-for-helmreleases

.PHONY: e2e-load-image
e2e-load-image: ko-build-all
	kind load docker-image --name $(K3S_CLUSTER) $(FULL_IMG):$(VERSION)

dev-kubeconf-user:
	@mkdir -p hack/kubeconfs || true
	@cd hack/kubeconfs \
	    && $(KUBECTL) get secret capsule-argocd-addon-proxy -n capsule-argocd-addon -o jsonpath='{.data.ca\.crt}' | base64 -d > root-ca.pem \
		&& rm -f alice.kubeconfig \
		&& curl -s https://raw.githubusercontent.com/projectcapsule/capsule/main/hack/create-user.sh | bash -s -- alice projectcapsule.dev \
		&& mv alice-*.kubeconfig alice.kubeconfig \
		&& KUBECONFIG=alice.kubeconfig $(KUBECTL) config set clusters.kind-$(K3S_CLUSTER).server https://127.0.0.1:9001 \
		&& KUBECONFIG=alice.kubeconfig $(KUBECTL) config set clusters.kind-$(K3S_CLUSTER).certificate-authority-data $$(cat root-ca.pem | base64 |tr -d '\n')

wait-for-helmreleases:
	@ echo "Waiting for all HelmReleases to have observedGeneration >= 0..."
	@while [ "$$($(KUBECTL) get helmrelease -A -o jsonpath='{range .items[?(@.status.observedGeneration<0)]}{.metadata.namespace}{" "}{.metadata.name}{"\n"}{end}' | wc -l)" -ne 0 ]; do \
	  sleep 5; \
	done


# Setup development env
# Usage:
# 	LAPTOP_HOST_IP=<YOUR_LAPTOP_IP> make dev-setup
# For example:
#	LAPTOP_HOST_IP=192.168.10.101 make dev-setup
define TLS_CNF
[ req ]
default_bits       = 4096
distinguished_name = req_distinguished_name
req_extensions     = req_ext
[ req_distinguished_name ]
countryName                = SG
stateOrProvinceName        = SG
localityName               = SG
organizationName           = CAPSULE
commonName                 = CAPSULE
[ req_ext ]
subjectAltName = @alt_names
[alt_names]
IP.1   = $(LAPTOP_HOST_IP)
endef
export TLS_CNF
dev-setup: helm
	mkdir -p /tmp/k8s-webhook-server/serving-certs
	echo "$${TLS_CNF}" > _tls.cnf
	openssl req -newkey rsa:4096 -days 3650 -nodes -x509 \
		-subj "/C=SG/ST=SG/L=SG/O=CAPSULE/CN=CAPSULE" \
		-extensions req_ext \
		-config _tls.cnf \
		-keyout /tmp/k8s-webhook-server/serving-certs/tls.key \
		-out /tmp/k8s-webhook-server/serving-certs/tls.crt
	$(KUBECTL) create secret tls capsule-tls -n capsule-system \
		--cert=/tmp/k8s-webhook-server/serving-certs/tls.crt\
		--key=/tmp/k8s-webhook-server/serving-certs/tls.key || true
	rm -f _tls.cnf
	export WEBHOOK_URL="https://$${LAPTOP_HOST_IP}:9443"; \
	export CA_BUNDLE=`openssl base64 -in /tmp/k8s-webhook-server/serving-certs/tls.crt | tr -d '\n'`; \
	$(HELM) upgrade \
	    --dependency-update \
		--debug \
		--install \
		--namespace capsule-argo-addon \
		--create-namespace \
		--set 'crds.install=true' \
		--set webhooks.enabled=true \
		--set "webhooks.service.url=$${WEBHOOK_URL}" \
		--set "webhooks.service.caBundle=$${CA_BUNDLE}" \
		capsule-argo-addon \
		./charts/capsule-argo-addon
	$(KUBECTL) -n capsule-argo-addon scale deployment --all --replicas=0 || true


##@ Build Dependencies

## Location to install dependencies to
LOCALBIN ?= $(shell pwd)/bin
$(LOCALBIN):
	mkdir -p $(LOCALBIN)

####################
# -- Helm Plugins
####################

HELM_SCHEMA_VERSION   := ""
helm-plugin-schema: helm
	@$(HELM) plugin install https://github.com/losisin/helm-values-schema-json.git --version $(HELM_SCHEMA_VERSION) || true

HELM_DOCS         := $(LOCALBIN)/helm-docs
HELM_DOCS_VERSION := v1.14.1
HELM_DOCS_LOOKUP  := norwoodj/helm-docs
helm-doc:
	@test -s $(HELM_DOCS) || \
	$(call go-install-tool,$(HELM_DOCS),github.com/$(HELM_DOCS_LOOKUP)/cmd/helm-docs@$(HELM_DOCS_VERSION))

####################
# -- Tools
####################
CONTROLLER_GEN         := $(LOCALBIN)/controller-gen
CONTROLLER_GEN_VERSION ?= v0.17.1
CONTROLLER_GEN_LOOKUP  := kubernetes-sigs/controller-tools
controller-gen:
	@test -s $(CONTROLLER_GEN) && $(CONTROLLER_GEN) --version | grep -q $(CONTROLLER_GEN_VERSION) || \
	$(call go-install-tool,$(CONTROLLER_GEN),sigs.k8s.io/controller-tools/cmd/controller-gen@$(CONTROLLER_GEN_VERSION))

GINKGO := $(LOCALBIN)/ginkgo
ginkgo:
	$(call go-install-tool,$(GINKGO),github.com/onsi/ginkgo/v2/ginkgo)

CT         := $(LOCALBIN)/ct
CT_VERSION := v3.13.0
CT_LOOKUP  := helm/chart-testing
ct:
	@test -s $(CT) && $(CT) version | grep -q $(CT_VERSION) || \
	$(call go-install-tool,$(CT),github.com/$(CT_LOOKUP)/v3/ct@$(CT_VERSION))

KIND         := $(LOCALBIN)/kind
KIND_VERSION := v0.29.0
KIND_LOOKUP  := kubernetes-sigs/kind
kind:
	@test -s $(KIND) && $(KIND) --version | grep -q $(KIND_VERSION) || \
	$(call go-install-tool,$(KIND),sigs.k8s.io/kind/cmd/kind@$(KIND_VERSION))

HELM         := $(LOCALBIN)/helm
HELM_VERSION := v3.17.2
HELM_LOOKUP  := helm/helm
helm:
	@test -s $(HELM) && $(HELM) version | grep -q $(HELM_VERSION) || \
	$(call go-install-tool,$(HELM),helm.sh/helm/v3/cmd/helm@$(HELM_VERSION))

KO           := $(LOCALBIN)/ko
KO_VERSION   := v0.18.0
KO_LOOKUP    := google/ko
ko:
	@test -s $(KO) && $(KO) -h | grep -q $(KO_VERSION) || \
	$(call go-install-tool,$(KO),github.com/$(KO_LOOKUP)@$(KO_VERSION))

GOLANGCI_LINT          := $(LOCALBIN)/golangci-lint
GOLANGCI_LINT_VERSION  := v1.64.8
GOLANGCI_LINT_LOOKUP   := golangci/golangci-lint
golangci-lint: ## Download golangci-lint locally if necessary.
	@test -s $(GOLANGCI_LINT) && $(GOLANGCI_LINT) -h | grep -q $(GOLANGCI_LINT_VERSION) || \
	$(call go-install-tool,$(GOLANGCI_LINT),github.com/$(GOLANGCI_LINT_LOOKUP)/cmd/golangci-lint@$(GOLANGCI_LINT_VERSION))

APIDOCS_GEN         := $(LOCALBIN)/crdoc
APIDOCS_GEN_VERSION := v0.6.4
APIDOCS_GEN_LOOKUP  := fybrik/crdoc
apidocs-gen: ## Download crdoc locally if necessary.
	@test -s $(APIDOCS_GEN) && $(APIDOCS_GEN) --version | grep -q $(APIDOCS_GEN_VERSION) || \
	$(call go-install-tool,$(APIDOCS_GEN),fybrik.io/crdoc@$(APIDOCS_GEN_VERSION))

# go-install-tool will 'go install' any package $2 and install it to $1.
PROJECT_DIR := $(shell dirname $(abspath $(lastword $(MAKEFILE_LIST))))
define go-install-tool
[ -f $(1) ] || { \
    set -e ;\
    GOBIN=$(LOCALBIN) go install $(2) ;\
}
endef
