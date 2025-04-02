package main

import (
	"os"

	"github.com/cert-manager/cert-manager/pkg/acme/webhook/cmd"
	"github.com/ionos-cloud/cert-manager-webhook-ionos-cloud/internal/resolver"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"

	"go.uber.org/zap"
)

// groupName is the K8s API group.
var (
	groupName = os.Getenv("GROUP_NAME")
	namespace = os.Getenv("NAMESPACE")
)

func main() {
	if groupName == "" {
		panic("GROUP_NAME must be specified")
	}

	if namespace == "" {
		panic("NAMESPACE must be specified")
	}

	logger, err := zap.NewProduction()
	if err != nil {
		panic(err)
	}

	config, err := rest.InClusterConfig()
	if err != nil {
		panic(err)
	}

	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		panic(err)
	}

	logger.Info("Starting webhook server")

	// This will register our custom DNS provider with the webhook serving
	// library, making it available as an API under the provided GroupName.
	// You can register multiple DNS provider implementations with a single
	// webhook, where the Name() method will be used to disambiguate between
	// the different implementations.
	cmd.RunWebhookServer(groupName, resolver.NewResolver(clientset, "", resolver.DefaultDNSAPIFactory, logger))
}
