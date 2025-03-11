package main_test

import (
	"os"
	"testing"

	acmetest "github.com/cert-manager/cert-manager/test/acme"
	"github.com/ionos-cloud/cert-manager-webhook-ionos-cloud/internal/clouddns"
	"github.com/ionos-cloud/cert-manager-webhook-ionos-cloud/internal/resolver"
	ionoscloud "github.com/ionos-cloud/sdk-go-dns"
	"go.uber.org/zap"
)

var (
	zone = os.Getenv("TEST_ZONE_NAME")
)

func TestRunsSuite(t *testing.T) {
	// The manifest path should contain a file named config.json that is a
	// snippet of valid configuration that should be included on the
	// ChallengeRequest passed as part of the test cases.

	logger, err := zap.NewDevelopment()
	if err != nil {
		t.Fatal(err.Error())
	}
	config := ionoscloud.NewConfigurationFromEnv()
	dnsClient := ionoscloud.NewAPIClient(config)

	solver := resolver.NewResolver(clouddns.CreateDNSAPI(dnsClient), logger)
	fixture := acmetest.NewFixture(solver,
		acmetest.SetResolvedZone(zone),
		acmetest.SetManifestPath("testdata"),
		acmetest.SetDNSServer("127.0.0.1:59351"),
		acmetest.SetUseAuthoritative(false),
	)
	//need to uncomment and  RunConformance delete runBasic and runExtended once https://github.com/cert-manager/cert-manager/pull/4835 is merged
	fixture.RunConformance(t)
	//fixture.RunBasic(t)
	//fixture.RunExtended(t)

}
