MOCKERY_VERSION = 2.52.2
GOLANGCI_VERSION = 1.62.0
HELM_DOCS_VERSION = 1.14.2
GO_LICENSES_VERSION = 1.6.0
LICENCES_IGNORE_LIST = $(shell cat licenses/licenses-ignore-list.txt)
ENVTEST_K8S_VERSION = 1.30
CERT_MANAGER_VERSION = v1.17.1

VERSION ?= 0.0.1
IMAGE_TAG_BASE ?= ionos-cloud/cert-manager-webhook-ionos-cloud
IMG ?= $(IMAGE_TAG_BASE):$(VERSION)

BUILD_VERSION ?= $(shell git branch --show-current)
BUILD_COMMIT ?= $(shell git rev-parse --short HEAD)
BUILD_TIMESTAMP ?= $(shell date -u '+%Y-%m-%d %H:%M:%S')

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

MOCKERY = bin/mockery-$(GOLANGCI_VERSION)
$(MOCKERY):
	GOBIN=$(PWD)/bin go install github.com/vektra/mockery/v2@v$(MOCKERY_VERSION)

.PHONY: generate-mocks
generate-mocks: $(MOCKERY) ## Generate mocks
	bin/mockery --name DNSAPI --output internal/clouddns/mocks --recursive

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

##@ static analysis

GOLANGCI_LINT = bin/golangci-lint-$(GOLANGCI_VERSION)
$(GOLANGCI_LINT):
	curl -sSfL https://raw.githubusercontent.com/golangci/golangci-lint/master/install.sh | bash -s -- -b bin v$(GOLANGCI_VERSION)
	@mv bin/golangci-lint "$(@)"

.PHONY: lint
lint: $(GOLANGCI_LINT) download ## Run linter
	$(GOLANGCI_LINT) run -v

##@ helm

HELM_DOCS = bin/helm-docs
$(HELM_DOCS):
	GOBIN=$(PWD)/bin go install github.com/norwoodj/helm-docs/cmd/helm-docs@v$(HELM_DOCS_VERSION)

helm-docs: $(HELM_DOCS) ## Generate helm documentation
	$(HELM_DOCS)


reports:
	@mkdir -pv "$(@)/licenses"

##@ release

GO_RELEASER = bin/goreleaser
$(GO_RELEASER):
	GOBIN=$(PWD)/bin go install github.com/goreleaser/goreleaser@latest


.PHONY: release-check
release-check: $(GO_RELEASER) ## Check if the release will work
	GITHUB_SERVER_URL=github.com GITHUB_REPOSITORY=ionos-cloud/cert-manager-webhook-ionos-cloud REGISTRY=$(REGISTRY) IMAGE_NAME=$(IMAGE_NAME) $(GO_RELEASER) release --snapshot --clean --skip-publish

##@ licenses

GO_LICENSES = bin/go-licenses
$(GO_LICENSES):
	GOBIN=$(PWD)/bin go install github.com/google/go-licenses@v$(GO_LICENSES_VERSION)

manualLicenses := $(shell cat .licenses/licenses-manual-list.csv | cut -d "," -f 1 | tr '\n' ',')
ignoredLicenses := $(shell cat .licenses/licenses-ignore-list.txt | tr '\n' ',')

check-licenses: $(GO_LICENSES)  ## Check the licenses
	$(GO_LICENSES) check --include_tests --ignore $(manualLicenses) --ignore $(ignoredLicenses) ./...


## conformance tests set up
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

# if running locally no need to repeat the setup steps for every run
conformace-test-standalone:
	KUBEBUILDER_ASSETS="$(shell $(ENVTEST) use $(ENVTEST_K8S_VERSION) -p path)" \
	TEST_ASSET_ETCD="$(PWD)/bin/tools/etcd" \
	TEST_ASSET_KUBE_APISERVER="$(PWD)/bin/tools/kube-apiserver" \
	TEST_ASSET_KUBECTL="$(PWD)/bin/tools/kubectl" \
	go test -tags=conformance -v cmd/webhook/main_test.go

conformance-test: envtest get-dependencies conformace-test-standalone

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