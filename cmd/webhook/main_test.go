//go:build conformance

package main_test

import (
	"os"
	"testing"

	acmetest "github.com/cert-manager/cert-manager/test/acme"
	"github.com/ionos-cloud/cert-manager-webhook-ionos-cloud/internal/resolver"
	"go.uber.org/zap"
)

var zone = os.Getenv("TEST_ZONE_NAME")

func TestBasicConformance(t *testing.T) {
	zone = "ionos-cloud-cdn.de"
	os.Setenv("KUBEBUILDER_ASSETS", "/home/zakaria/Desktop/myWork/ionos/workspace/cert-manager-webhook-ionos-cloud/bin/tools/k8s/1.36.2-linux-amd64")
	os.Setenv("TEST_ASSET_ETCD", "/home/zakaria/Desktop/myWork/ionos/workspace/cert-manager-webhook-ionos-cloud/bin/tools/etcd")
	os.Setenv("TEST_ASSET_KUBE_APISERVER", "/home/zakaria/Desktop/myWork/ionos/workspace/cert-manager-webhook-ionos-cloud/bin/tools/kube-apiserver")
	os.Setenv("TEST_ASSET_KUBECTL", "/home/zakaria/Desktop/myWork/ionos/workspace/cert-manager-webhook-ionos-cloud/bin/tools/kubectl")

	if zone == "" {
		t.Fatal("TEST_ZONE_NAME environment variable must be set before running the test")
	}
	logger, err := zap.NewDevelopment()
	if err != nil {
		t.Fatal(err.Error())
	}

	solver := resolver.NewResolver("basic-present-record", resolver.DefaultK8FactoryFactory,
		resolver.DefaultDNSAPIFactory, resolver.DefaultGenerateTokenFunc, logger)
	fixture := acmetest.NewFixture(solver,
		// cert-manager adds a dot a the end of the zone name
		acmetest.SetResolvedZone(zone+"."),
		acmetest.SetResolvedFQDN("_acme-challenge."+zone+"."),
		acmetest.SetManifestPath("./testdata"),
	)
	fixture.RunBasic(t)
}

func TestExtendedConformance(t *testing.T) {
	if zone == "" {
		t.Fatal("TEST_ZONE_NAME environment variable must be set before running the test")
	}
	logger, err := zap.NewDevelopment()
	if err != nil {
		t.Fatal(err.Error())
	}

	solver := resolver.NewResolver("extended-supports-multiple-same-domain", resolver.DefaultK8FactoryFactory,
		resolver.DefaultDNSAPIFactory, resolver.DefaultGenerateTokenFunc, logger)
	fixture := acmetest.NewFixture(solver,
		// cert-manager adds a dot a the end of the zone name
		acmetest.SetResolvedZone(zone+"."),
		acmetest.SetResolvedFQDN("_acme-challenge."+zone+"."),
		acmetest.SetManifestPath("./testdata"),
	)
	fixture.RunExtended(t)
}
