package resolver

import (
	"context"
	"errors"
	"github.com/ionos-cloud/cert-manager-webhook-ionos-cloud/internal/clouddns"
	dnsclient "github.com/ionos-cloud/sdk-go-dns"
	"k8s.io/utils/ptr"
	"strings"

	"github.com/cert-manager/cert-manager/pkg/acme/webhook"
	"github.com/cert-manager/cert-manager/pkg/acme/webhook/apis/acme/v1alpha1"
	"go.uber.org/zap"
	"k8s.io/client-go/rest"
)

var typeTxtRecord = ptr.To(dnsclient.RecordType("TXT"))

func NewResolver(client clouddns.DNSAPI, logger *zap.Logger) webhook.Solver {
	return &ionosCloudDnsProviderResolver{
		ctx:    context.Background(),
		client: client,
		logger: logger,
	}
}

type ionosCloudDnsProviderResolver struct {
	ctx    context.Context
	client clouddns.DNSAPI
	logger *zap.Logger
}

// Name is used as the name for this DNS solver when referencing it on the ACME
// Issuer resource.
// This should be unique **within the group name**, i.e. you can have two
// solvers configured with the same Name() **so long as they do not co-exist
// within a single webhook deployment**.
// For example, `cloudflare` may be used as the name of a solver.
func (s *ionosCloudDnsProviderResolver) Name() string {
	return "ionos-cloud"
}

// Present is responsible for actually presenting the DNS record with the
// DNS provider.
// This method should tolerate being called multiple times with the same value.
// cert-manager itself will later perform a self check to ensure that the
// solver has correctly configured the DNS provider.
func (s *ionosCloudDnsProviderResolver) Present(ch *v1alpha1.ChallengeRequest) error {
	s.logger.Debug("Received dns challenge request", zap.String("uid", string(ch.UID)), zap.String("key", ch.Key),
		zap.String("dnsName", ch.DNSName), zap.String("resolvedZone", ch.ResolvedZone), zap.String("resolvedFQDN",
			ch.ResolvedFQDN))
	zoneId, err := s.findOrCreateZone(ch, false)
	if err != nil {
		return err
	}
	return s.findOrCreateRecord(ch, zoneId)
}

// CleanUp should delete the relevant TXT record from the DNS provider console.
// If multiple TXT records exist with the same record name (e.g.
// _acme-challenge.example.com) then **only** the record with the same `key`
// value provided on the ChallengeRequest should be cleaned up.
// This is in order to facilitate multiple DNS validations for the same domain
// concurrently.
func (s *ionosCloudDnsProviderResolver) CleanUp(ch *v1alpha1.ChallengeRequest) error {
	zoneId, err := s.findOrCreateZone(ch, true)
	if err != nil {
		return err
	}
	return s.deleteRecord(ch, zoneId)
}

// Initialize will be called when the webhook first starts.
// This method can be used to instantiate the webhook, i.e. initializing
// connections or warming up caches.
// Typically, the kubeClientConfig parameter is used to build a Kubernetes
// client that can be used to fetch resources from the Kubernetes API, e.g.
// Secret resources containing credentials used to authenticate with DNS
// provider accounts.
// The stopCh can be used to handle early termination of the webhook, in cases
// where a SIGTERM or similar signal is sent to the webhook process.
func (s *ionosCloudDnsProviderResolver) Initialize(kubeClientConfig *rest.Config, stopCh <-chan struct{}) error {
	s.logger.Info("IONOS Cloud resolver initialized")
	return nil
}

func (s *ionosCloudDnsProviderResolver) findOrCreateZone(ch *v1alpha1.ChallengeRequest, mustFind bool) (string, error) {
	// fetch zone
	zoneName := strings.TrimSuffix(ch.ResolvedZone, ".")
	s.logger.Debug("find zone...", zap.String("zoneName", zoneName))
	zoneList, err := s.client.GetZones(zoneName)
	if err != nil {
		s.logger.Error("Error fetching zone", zap.Error(err))
		return "", err
	}
	if zoneList.Items == nil || len(*zoneList.Items) == 0 {
		if mustFind {
			s.logger.Error("Error fetching zone, zone not found", zap.String("zoneName", zoneName))
			return "", errors.New("error fetching zone")
		}
		s.logger.Debug("zone not found, try to create...", zap.String("zoneName", zoneName))
		zone, err := s.client.CreateZone(zoneName)
		if err != nil {
			s.logger.Error("Error creating zone", zap.Error(err))
			return "", err
		}
		s.logger.Info("zone created", zap.String("zoneName", zoneName), zap.String("zoneId", *zone.Id))
		return *zone.Id, nil
	}
	if len(*zoneList.Items) > 1 {
		s.logger.Error("Error fetching zone, zone not unique", zap.Int("zoneCount", len(*zoneList.Items)))
		return "", errors.New("error fetching zone")
	}
	zone := (*zoneList.Items)[0]
	s.logger.Info("zone found", zap.String("zoneName", zoneName), zap.String("zoneId", *zone.Id))
	return *zone.Id, nil
}

func (s *ionosCloudDnsProviderResolver) findOrCreateRecord(ch *v1alpha1.ChallengeRequest, zoneId string) error {
	recordName := strings.TrimSuffix(ch.ResolvedFQDN, "."+ch.ResolvedZone)
	s.logger.Debug("find txt record...", zap.String("recordName", recordName), zap.String("fqdn", ch.ResolvedFQDN),
		zap.String("zoneId", zoneId))
	recordList, err := s.client.GetRecords(zoneId, recordName)
	if err != nil {
		s.logger.Error("Error fetching record", zap.Error(err))
		return err
	}
	if recordList.Items == nil || len(*recordList.Items) == 0 {
		s.logger.Debug("record not found, try to create record...", zap.String("recordName", recordName), zap.String("key", ch.Key),
			zap.String("zoneId", zoneId))
		record, err := s.client.CreateTXTRecord(zoneId, recordName, ch.Key)
		if err != nil {
			s.logger.Error("Error creating record", zap.Error(err))
			return err
		}
		s.logger.Info("record for dns challenge successfully created", zap.String("recordId", *record.Id),
			zap.String("recordName", recordName), zap.String("zoneId", zoneId))
		return nil
	}
	if len(*recordList.Items) > 1 {
		s.logger.Error("Error fetching record, record not unique", zap.Int("recordCount", len(*recordList.Items)),
			zap.String("recordName", recordName))
		return errors.New("error fetching record")
	}
	record := (*recordList.Items)[0]
	s.logger.Info("record found", zap.String("recordName", recordName), zap.String("recordId", *record.Id),
		zap.String("zoneId", zoneId))
	return nil
}

func (s *ionosCloudDnsProviderResolver) deleteRecord(ch *v1alpha1.ChallengeRequest, zoneId string) error {
	// fetch record
	recordName := strings.TrimSuffix(ch.ResolvedFQDN, ch.ResolvedZone)
	s.logger.Debug("try to find txt record...", zap.String("recordName", recordName), zap.String("fqdn", ch.ResolvedFQDN), zap.String("zoneId", zoneId))
	recordList, err := s.client.GetRecords(zoneId, recordName)
	if err != nil {
		s.logger.Error("Error fetching record", zap.Error(err))
		return err
	}
	if recordList.Items == nil || len(*recordList.Items) == 0 {
		s.logger.Info("record not found, nothing to clean up", zap.String("recordName", recordName))
		return nil
	}
	if len(*recordList.Items) > 1 {
		s.logger.Error("Error fetching record, record not unique", zap.Int("recordCount", len(*recordList.Items)),
			zap.String("recordName", recordName))
		return errors.New("error fetching record")
	}
	record := (*recordList.Items)[0]
	s.logger.Info("record found, try to delete...", zap.String("recordName", recordName), zap.String("recordId", *record.Id))
	err = s.client.DeleteRecord(zoneId, *record.Id)
	if err != nil {
		s.logger.Error("Error deleting record", zap.Error(err))
		return err
	}
	s.logger.Info("record successfully deleted", zap.String("recordId", *record.Id))
	return nil
}
