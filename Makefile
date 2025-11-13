
ENVTEST_K8S_VERSION = 1.34

VERSION ?= 0.0.1
IMAGE_TAG_BASE ?= ionos-cloud/cert-manager-webhook-ionos-cloud
IMG ?= $(IMAGE_TAG_BASE):$(VERSION)

BUILD_VERSION ?= $(shell git branch --show-current)
BUILD_COMMIT ?= $(shell git rev-parse --short HEAD)
BUILD_TIMESTAMP ?= $(shell date -u '+%Y-%m-%d %H:%M:%S')
GO_TOOL := go tool

PWD = $(shell pwd)
export PATH := $(PWD)/bin:$(PATH)

##@ General

.PHONY: help
help: ## Display this help.
	@awk 'BEGIN {FS = ":.*##"; printf "\nUsage:\n  make \033[36m<target>\033[0m\n"} /^[a-zA-Z_0-9-]+:.*?##/ { printf "  \033[36m%-15s\033[0m %s\n", $$1, $$2 } /^##@/ { printf "\n\033[1m%s\033[0m\n", substr($$0, 5) } ' $(MAKEFILE_LIST)


out:
	@mkdir -pv "$(@)"

.PHONY: download
download: ## Download dependencies
	go mod download

.PHONY: clean
clean:
	rm -rf ./bin
	rm -rf ./out
	rm -rf ./internal/clouddns/mocks


##@ Development cycle
.PHONY: generate-mocks
generate-mocks: ## Generate mocks
	$(GO_TOOL) github.com/vektra/mockery/v2

.PHONY: build
build: ## Build the binary
	CGO_ENABLED=0 go build -ldflags "-s -w" -o ./bin/cert-manager-webhook-ionos-cloud -v cmd/webhook/main.go

GO_TEST = go tool gotest.tools/gotestsum --format pkgname
.PHONY: unit-test
unit-test: out ## Run unit tests with coverage and generate json report
	$(GO_TEST) --junitfile out/report.xml -- -race ./... -count=1 -short -tags=unit -cover -coverprofile=out/cover.out

.PHONY: html-coverage
html-coverage: out/report.xml ## Generate html coverage report
	go tool cover -html=out/cover.out

.PHONY: run
run: ## Run the application
	go run cmd/webhook/main.go

.PHONY: fmt
fmt: ## Run go fmt against code.
	$(GO_TOOL) mvdan.cc/gofumpt -w .

##@ static analysis

.PHONY: lint
lint: 
	$(GO_TOOL) github.com/golangci/golangci-lint/cmd/golangci-lint run -v

.PHONY: lint-with-fix
lint-with-fix: ## Run golangci-lint against code with fix.
	$(GO_TOOL) github.com/golangci/golangci-lint/cmd/golangci-lint run --build-tags unit --fix

.PHONY: vet
vet: ## Run go vet against code.
	go vet -tags unit ./...

.PHONY: static-analysis
static-analysis: lint vet ## Run static analysis against code.

##@ helm

helm-docs: ## Generate helm documentation
	$(GO_TOOL) github.com/norwoodj/helm-docs/cmd/helm-docs

reports:
	@mkdir -pv "$(@)/licenses"

.PHONY: release-check
release-check: ## Check if the release will work
	GITHUB_SERVER_URL=github.com GITHUB_REPOSITORY=ionos-cloud/cert-manager-webhook-ionos-cloud \
	REGISTRY=$(REGISTRY) \
	IMAGE_NAME=$(IMAGE_NAME) \
	$(GO_TOOL) github.com/goreleaser/goreleaser/v2 release --snapshot --clean --skip=publish

##@ licenses

manualLicenses := $(shell cat .licenses/licenses-manual-list.csv | cut -d "," -f 1 | tr '\n' ',')
ignoredLicenses := $(shell cat .licenses/licenses-ignore-list.txt | tr '\n' ',')

check-licenses: ## Check the licenses
	$(GO_TOOL) github.com/google/go-licenses/v2 check --include_tests --ignore $(manualLicenses) --ignore $(ignoredLicenses) ./...


##@ conformance tests
ENVTEST = $(shell pwd)/bin/setup-envtest
.PHONY: envtest
envtest: ## Download envtest-setup locally if necessary.
	$(call go-get-tool,$(ENVTEST),sigs.k8s.io/controller-runtime/tools/setup-envtest@latest)



# this step should be removed, the binaries downloaded are not needed in tests
# their are just here because of some old code here: 
# https://github.com/cert-manager/cert-manager/blob/master/test/apiserver/envs.go#L31
get-dependencies:
	mkdir -p bin/tools
	wget -P bin/tools https://cloud-dns-experimental.s3-eu-central-2.ionoscloud.com/test-binaries/etcd
	wget -P bin/tools https://cloud-dns-experimental.s3-eu-central-2.ionoscloud.com/test-binaries/kube-apiserver
	wget -P bin/tools https://cloud-dns-experimental.s3-eu-central-2.ionoscloud.com/test-binaries/kubectl
	chmod 755 bin/tools/etcd
	chmod 755 bin/tools/kube-apiserver
	chmod 755 bin/tools/kubectl


conformance-test-standalone: ## runs conformance tests without setup, if running locally no need to repeat the setup steps for every run
	kubectl create secret generic test-ionos-cloud-credentials \
	--from-literal=token=$$IONOS_TOKEN --dry-run=client -o yaml > cmd/webhook/testdata/ionos-credentials.test.yaml
	KUBEBUILDER_ASSETS="$(shell $(ENVTEST) use $(ENVTEST_K8S_VERSION) -p path)" \
	TEST_ASSET_ETCD="$(PWD)/bin/tools/etcd" \
	TEST_ASSET_KUBE_APISERVER="$(PWD)/bin/tools/kube-apiserver" \
	TEST_ASSET_KUBECTL="$(PWD)/bin/tools/kubectl" \
	go test -tags=conformance -v cmd/webhook/main_test.go


conformance-test: envtest get-dependencies conformance-test-standalone ## runs conformance tests

# go-get-tool will 'go get' any package $2 and install it to $1.
# source: https://book.kubebuilder.io/cronjob-tutorial/basic-project#build-infrastructure
PROJECT_DIR := $(shell dirname $(abspath $(lastword $(MAKEFILE_LIST))))
define go-get-tool
@[ -f $(1) ] || { \
set -e ;\
TMP_DIR=$$(mktemp -d) ;\
cd $$TMP_DIR ;\
go mod init tmp ;\
echo "Downloading $(2)" ;\
GOBIN=$(PROJECT_DIR)/bin go install $(2) ;\
rm -rf $$TMP_DIR ;\
}
endef