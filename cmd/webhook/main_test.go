//go:build conformance

package main_test

import (
	"os"
	"testing"

	acmetest "github.com/cert-manager/cert-manager/test/acme"
	"github.com/go-logr/logr"
	"github.com/ionos-cloud/cert-manager-webhook-ionos-cloud/internal/clouddns"
	"github.com/ionos-cloud/cert-manager-webhook-ionos-cloud/internal/resolver"
	ionoscloud "github.com/ionos-cloud/sdk-go-dns"
	"go.uber.org/zap"
	controller_runtime_log "sigs.k8s.io/controller-runtime/pkg/log"
)

var zone = os.Getenv("TEST_ZONE_NAME")

func TestRunsSuite(t *testing.T) {
	if zone == "" {
		t.Fatal("TEST_ZONE_NAME environment variable must be set before running the test")
	}
	logger, err := zap.NewDevelopment()
	if err != nil {
		t.Fatal(err.Error())
	}

	// this is to remove a log message warning in controller runtime
	controller_runtime_log.SetLogger(logr.New(controller_runtime_log.NullLogSink{}))

	config := ionoscloud.NewConfigurationFromEnv()
	dnsClient := ionoscloud.NewAPIClient(config)

	solver := resolver.NewResolver(clouddns.CreateDNSAPI(dnsClient), logger)
	fixture := acmetest.NewFixture(solver,
		// cert-manager adds a dot a the end of the zone name
		acmetest.SetResolvedZone(zone+"."),
		acmetest.SetResolvedFQDN("_acme-challenge."+zone+"."),
		acmetest.SetConfig("testdata"),
	)
	fixture.RunConformance(t)
}
