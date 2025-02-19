package main

import (
	"github.com/cert-manager/cert-manager/pkg/acme/webhook/cmd"
	"github.com/ionos-cloud/cert-manager-webhook-ionos-cloud/internal/resolver"
	ionoscloud "github.com/ionos-cloud/sdk-go-dns"
	"go.uber.org/zap"
	"os"
)

// GroupName is the K8s API group.
var GroupName = os.Getenv("GROUP_NAME")

func main() {
	if GroupName == "" {
		panic("GROUP_NAME must be specified")
	}

	logger, err := zap.NewProduction()
	if err != nil {
		panic(err)
	}
	dnsClient := ionoscloud.NewAPIClient(ionoscloud.NewConfigurationFromEnv())
	logger.Info("Starting webhook server")

	// This will register our custom DNS provider with the webhook serving
	// library, making it available as an API under the provided GroupName.
	// You can register multiple DNS provider implementations with a single
	// webhook, where the Name() method will be used to disambiguate between
	// the different implementations.
	cmd.RunWebhookServer(GroupName, resolver.NewResolver(dnsClient, logger))
}
