package resolver

import (
	"context"
	"fmt"
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
	zoneId, err := s.findZone(ch, true)
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
	zoneId, err := s.findZone(ch, false)
	if err != nil {
		return err
	}
	if zoneId == "" {
		s.logger.Info("zone not found, nothing to clean up", zap.String("zoneName", ch.ResolvedZone))
		return nil
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

func (s *ionosCloudDnsProviderResolver) findZone(ch *v1alpha1.ChallengeRequest, shouldFind bool) (string, error) {
	// fetch zone
	zoneName := zoneNameFromChallenge(ch)
	s.logger.Debug("find zone...", zap.String("zoneName", zoneName))
	zoneList, err := s.client.GetZones(zoneName)
	if err != nil {
		s.logger.Error("Error fetching zone", zap.Error(err))
		return "", err
	}
	if zoneList.Items == nil || len(*zoneList.Items) == 0 {
		s.logger.Info("zone not found", zap.String("zoneName", zoneName))
		if shouldFind {
			return "", fmt.Errorf("zone '%s' not found", zoneName)
		}
		return "", nil
	}
	zone := (*zoneList.Items)[0]
	s.logger.Info("zone found", zap.String("zoneName", zoneName), zap.String("zoneId", *zone.Id))
	return *zone.Id, nil
}

func (s *ionosCloudDnsProviderResolver) findOrCreateRecord(ch *v1alpha1.ChallengeRequest, zoneId string) error {
	recordName := recordNameFromChallenge(ch)
	s.logger.Debug("find txt record...", zap.String("recordName", recordName), zap.String("fqdn", ch.ResolvedFQDN),
		zap.String("zoneId", zoneId))
	recordList, err := s.client.GetRecords(zoneId, recordName)
	if err != nil {
		s.logger.Error("Error fetching record", zap.Error(err))
		return err
	}
	// check if record already exists
	for _, r := range *recordList.Items {
		content := r.GetProperties().GetContent()
		if content != nil && *content == ch.Key {
			s.logger.Info("record for dns challenge already exists", zap.String("recordId", *r.Id),
				zap.String("recordName", recordName), zap.String("zoneId", zoneId))
			return nil
		}
	}
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

func (s *ionosCloudDnsProviderResolver) deleteRecord(ch *v1alpha1.ChallengeRequest, zoneId string) error {
	recordName := recordNameFromChallenge(ch)
	s.logger.Debug("try to find txt record...", zap.String("recordName", recordName), zap.String("fqdn", ch.ResolvedFQDN), zap.String("zoneId", zoneId))
	recordList, err := s.client.GetRecords(zoneId, recordName)
	if err != nil {
		s.logger.Error("Error fetching record", zap.Error(err))
		return err
	}
	if recordList.Items == nil || len(*recordList.Items) == 0 {
		s.logger.Info("no record with that name found, nothing to clean up", zap.String("recordName", recordName),
			zap.String("zoneId", zoneId))
		return nil
	}
	var record *dnsclient.RecordRead
	for _, r := range *recordList.Items {
		content := r.GetProperties().GetContent()
		if content != nil && *content == ch.Key {
			record = &r
			break
		}
	}
	if record == nil {
		s.logger.Info("record with that name found, but key differs, nothing to clean up",
			zap.String("recordName", recordName), zap.String("zoneId", zoneId))
		return nil
	}
	s.logger.Info("record found, deleting...", zap.String("recordName", recordName), zap.String("recordId", *record.Id))
	err = s.client.DeleteRecord(zoneId, *record.Id)
	if err != nil {
		s.logger.Error("Error deleting record", zap.Error(err))
		return err
	}
	s.logger.Info("record successfully deleted", zap.String("recordId", *record.Id), zap.String("recordName", recordName),
		zap.String("zoneId", zoneId))
	return nil
}

func recordNameFromChallenge(ch *v1alpha1.ChallengeRequest) string {
	return strings.TrimSuffix(ch.ResolvedFQDN, "."+ch.ResolvedZone)
}

func zoneNameFromChallenge(ch *v1alpha1.ChallengeRequest) string {
	return strings.TrimSuffix(ch.ResolvedZone, ".")
}
