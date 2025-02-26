package main

import (
	"context"
	"github.com/ionos-cloud/cert-manager-webhook-ionos-cloud/internal/dnsclient"
	"os"
	"testing"
)

func TestGetZones(t *testing.T) {
	config := dnsclient.NewConfiguration()
	token := os.Getenv("IONOS_TOKEN")
	config.DefaultHeader["Authorization"] = "Bearer " + token
	config.Debug = true
	client := dnsclient.NewAPIClient(config)
	ctx := context.Background()
	zoneList, resp, err := client.ZonesAPI.ZonesGet(ctx).FilterZoneName("alexkrieg.com").Execute()
	t.Logf("zoneList: %v resp: %v err: %v", zoneList, resp, err)
}
