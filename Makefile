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
	@GO111MODULE=on go test -v ./... -coverprofile coverage.out

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
	$(CONTROLLER_GEN) rbac:roleName=manager-role crd webhook paths="./..." output:crd:artifacts:config=charts/capsule-argo-addon/crds

# Generate code
generate: controller-gen
	$(CONTROLLER_GEN) object:headerFile="hack/boilerplate.go.txt" paths="./..."

apidocs: TARGET_DIR      := $(shell mktemp -d)
apidocs: apidocs-gen generate
	$(APIDOCS_GEN) crdoc --resources charts/capsule-argo-addon/crds --output docs/reference.md --template ./hack/templates/crds.tmpl


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

helm-docs: HELMDOCS_VERSION := v1.11.0
helm-docs: docker
	@docker run -v "$(SRC_ROOT):/helm-docs" jnorwood/helm-docs:$(HELMDOCS_VERSION) --chart-search-root /helm-docs

helm-lint: CT_VERSION := v3.11.0
helm-lint: docker
	@docker run -v "$(SRC_ROOT):/workdir" --entrypoint /bin/sh quay.io/helmpack/chart-testing:$(CT_VERSION) -c "cd /workdir; ct lint --config .github/configs/ct.yaml  --lint-conf .github/configs/lintconf.yaml  --all --debug"

helm-test: kind ct
	@$(KIND) create cluster --wait=60s --name helm-capsule-argo-addon
	@$(MAKE) e2e-install-distro
	@$(MAKE) helm-test-exec
	@$(KIND) delete cluster --name helm-capsule-argo-addon

helm-test-exec: ct ko-build-all
	@$(KIND) load docker-image --name helm-capsule-argo-addon $(FULL_IMG):latest
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
e2e-install-addon: e2e-load-image
	helm upgrade \
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
		capsule-argo-addon \
		./charts/capsule-argo-addon

e2e-install-distro:
	@flux install
	@kubectl kustomize e2e/objects/distro/ | kubectl apply -f -
	@$(MAKE) wait-for-helmreleases

.PHONY: e2e-load-image
e2e-load-image: ko-build-all
	kind load docker-image --name $(K3S_CLUSTER) $(FULL_IMG):$(VERSION)

dev-kubeconf-user: 
	@mkdir -p hack/kubeconfs || true
	@cd hack/kubeconfs \
	    && kubectl get secret capsule-argocd-addon-proxy -n capsule-argocd-addon -o jsonpath='{.data.ca\.crt}' | base64 -d > root-ca.pem \
		&& rm -f alice.kubeconfig \
		&& curl -s https://raw.githubusercontent.com/projectcapsule/capsule/main/hack/create-user.sh | bash -s -- alice projectcapsule.dev \
		&& mv alice-*.kubeconfig alice.kubeconfig \
		&& KUBECONFIG=alice.kubeconfig kubectl config set clusters.kind-$(K3S_CLUSTER).server https://127.0.0.1:9001 \
		&& KUBECONFIG=alice.kubeconfig kubectl config set clusters.kind-$(K3S_CLUSTER).certificate-authority-data $$(cat root-ca.pem | base64 |tr -d '\n') 

wait-for-helmreleases:
	@ echo "Waiting for all HelmReleases to have observedGeneration >= 0..."
	@while [ "$$(kubectl get helmrelease -A -o jsonpath='{range .items[?(@.status.observedGeneration<0)]}{.metadata.namespace}{" "}{.metadata.name}{"\n"}{end}' | wc -l)" -ne 0 ]; do \
	  sleep 5; \
	done

####################
# -- Tools
####################
CONTROLLER_GEN         := $(shell pwd)/bin/controller-gen
CONTROLLER_GEN_VERSION := v0.16.3
controller-gen: ## Download controller-gen locally if necessary.
	$(call go-install-tool,$(CONTROLLER_GEN),sigs.k8s.io/controller-tools/cmd/controller-gen@$(CONTROLLER_GEN_VERSION))

GINKGO         := $(shell pwd)/bin/ginkgo
ginkgo: ## Download ginkgo locally if necessary.
	$(call go-install-tool,$(GINKGO),github.com/onsi/ginkgo/v2/ginkgo)

CT         := $(shell pwd)/bin/ct
CT_VERSION := v3.11.0
ct: ## Download ct locally if necessary.
	$(call go-install-tool,$(CT),github.com/helm/chart-testing/v3/ct@$(CT_VERSION))

KIND         := $(shell pwd)/bin/kind
KIND_VERSION := v0.17.0
kind: ## Download kind locally if necessary.
	$(call go-install-tool,$(KIND),sigs.k8s.io/kind/cmd/kind@$(KIND_VERSION))

KUSTOMIZE         := $(shell pwd)/bin/kustomize
KUSTOMIZE_VERSION := 3.8.7
kustomize: ## Download kustomize locally if necessary.
	$(call install-kustomize,$(KUSTOMIZE),$(KUSTOMIZE_VERSION))

KO = $(shell pwd)/bin/ko
KO_VERSION = v0.14.1
ko:
	$(call go-install-tool,$(KO),github.com/google/ko@$(KO_VERSION))


GOLANGCI_LINT = $(shell pwd)/bin/golangci-lint
GOLANGCI_LINT_VERSION = v1.56.2
golangci-lint: ## Download golangci-lint locally if necessary.
	$(call go-install-tool,$(GOLANGCI_LINT),github.com/golangci/golangci-lint/cmd/golangci-lint@$(GOLANGCI_LINT_VERSION))

APIDOCS_GEN         := $(shell pwd)/bin/crdoc
APIDOCS_GEN_VERSION := latest
apidocs-gen: ## Download crdoc locally if necessary.
	$(call go-install-tool,$(APIDOCS_GEN),fybrik.io/crdoc@$(APIDOCS_GEN_VERSION))

# go-install-tool will 'go install' any package $2 and install it to $1.
PROJECT_DIR := $(shell dirname $(abspath $(lastword $(MAKEFILE_LIST))))
define go-install-tool
@[ -f $(1) ] || { \
set -e ;\
GOBIN=$(PROJECT_DIR)/bin go install $(2) ;\
}
endef