//go:build unit

package resolver

import (
	"fmt"
	"strings"
	"testing"

	"github.com/cert-manager/cert-manager/pkg/acme/webhook/apis/acme/v1alpha1"
	"github.com/ionos-cloud/cert-manager-webhook-ionos-cloud/internal/clouddns/mocks"
	dnsclient "github.com/ionos-cloud/sdk-go-dns"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
	"go.uber.org/zap"
	"k8s.io/utils/ptr"
)

type ResolverTestSuite struct {
	suite.Suite
	dnsAPIMock *mocks.DNSAPI
	logger     *zap.Logger
}

func TestSuite(t *testing.T) {
	suite.Run(t, new(ResolverTestSuite))
}

func (s *ResolverTestSuite) SetupSuite() {
	logger, err := zap.NewDevelopment()
	require.NoError(s.T(), err)
	s.logger = logger
}

func (s *ResolverTestSuite) setupMocks() {
	s.dnsAPIMock = mocks.NewDNSAPI(s.T())
	s.logger.Debug("apiClient with mocks is created")
}

func (s *ResolverTestSuite) TestPresent() {
	testCases := []struct {
		name                  string
		givenZones            []dnsclient.ZoneRead
		givenRecords          []dnsclient.RecordRead
		whenChallenge         *v1alpha1.ChallengeRequest
		whenZonesReadError    error
		whenRecordsReadError  error
		whenRecordCreateError error
		thenError             string
		thenRecordCreateKey   string
	}{
		{
			name:       "no zones",
			givenZones: []dnsclient.ZoneRead{},
			whenChallenge: &v1alpha1.ChallengeRequest{
				UID:          "test-UID",
				Key:          "test-key",
				DNSName:      "*.test.com",
				ResolvedZone: "test.com",
				ResolvedFQDN: "_acme-challenge.test.com.",
			},
			thenError: "zone 'test.com' not found",
		},
		{
			name: "zone already exists",
			givenZones: []dnsclient.ZoneRead{
				{
					Id: ptr.To("test-zone-id"),
					Properties: &dnsclient.Zone{
						ZoneName: ptr.To("test.com"),
					},
					Type: ptr.To("NATIVE"),
				},
			},
			givenRecords: []dnsclient.RecordRead{},
			whenChallenge: &v1alpha1.ChallengeRequest{
				UID:          "test-UID",
				Key:          "test-key",
				DNSName:      "*.test.com",
				ResolvedZone: "test.com",
				ResolvedFQDN: "_acme-challenge.test.com.",
			},
			thenRecordCreateKey: "test-key",
		},
		{
			name: "record with the same name and key already exists",
			givenZones: []dnsclient.ZoneRead{
				{
					Id: ptr.To("test-zone-id"),
					Properties: &dnsclient.Zone{
						ZoneName: ptr.To("test.com"),
					},
					Type: ptr.To("NATIVE"),
				},
			},
			givenRecords: []dnsclient.RecordRead{
				{
					Id: ptr.To("test-record-id"),
					Properties: &dnsclient.Record{
						Name:    ptr.To("_acme-challenge"),
						Type:    typeTxtRecord,
						Content: ptr.To("test-key"),
					},
				},
			},
			whenChallenge: &v1alpha1.ChallengeRequest{
				UID:          "test-UID",
				Key:          "test-key",
				DNSName:      "*.test.com",
				ResolvedZone: "test.com",
				ResolvedFQDN: "_acme-challenge.test.com.",
			},
			thenRecordCreateKey: "", // no record should be created
		},
		{
			name: "record with the same name but different key already exists",
			givenZones: []dnsclient.ZoneRead{
				{
					Id: ptr.To("test-zone-id"),
					Properties: &dnsclient.Zone{
						ZoneName: ptr.To("test.com"),
					},
					Type: ptr.To("NATIVE"),
				},
			},
			givenRecords: []dnsclient.RecordRead{
				{
					Id: ptr.To("test-record-id"),
					Properties: &dnsclient.Record{
						Name:    ptr.To("_acme-challenge"),
						Type:    typeTxtRecord,
						Content: ptr.To("different-key"),
					},
				},
			},
			whenChallenge: &v1alpha1.ChallengeRequest{
				UID:          "test-UID",
				Key:          "test-key",
				DNSName:      "*.test.com",
				ResolvedZone: "test.com",
				ResolvedFQDN: "_acme-challenge.test.com.",
			},
			thenRecordCreateKey: "test-key",
		},
		{
			name:               "error fetching zones",
			givenZones:         []dnsclient.ZoneRead{},
			whenZonesReadError: fmt.Errorf("error fetching zones"),
			whenChallenge: &v1alpha1.ChallengeRequest{
				UID:          "test-UID",
				Key:          "test-key",
				DNSName:      "*.test.com",
				ResolvedZone: "test.com",
				ResolvedFQDN: "_acme-challenge.test.com.",
			},
			thenError: "error fetching zones",
		},
		{
			name: "error fetching records",
			givenZones: []dnsclient.ZoneRead{
				{
					Id: ptr.To("test-zone-id"),
					Properties: &dnsclient.Zone{
						ZoneName: ptr.To("test.com"),
					},
					Type: ptr.To("NATIVE"),
				},
			},
			givenRecords:         []dnsclient.RecordRead{},
			whenRecordsReadError: fmt.Errorf("error fetching records"),
			whenChallenge: &v1alpha1.ChallengeRequest{
				UID:          "test-UID",
				Key:          "test-key",
				DNSName:      "*.test.com",
				ResolvedZone: "test.com",
				ResolvedFQDN: "_acme-challenge.test.com.",
			},
			thenError: "error fetching records",
		},
		{
			name: "error creating record",
			givenZones: []dnsclient.ZoneRead{
				{
					Id: ptr.To("test-zone-id"),
					Properties: &dnsclient.Zone{
						ZoneName: ptr.To("test.com"),
					},
					Type: ptr.To("NATIVE"),
				},
			},
			givenRecords:          []dnsclient.RecordRead{},
			whenRecordCreateError: fmt.Errorf("error creating record"),
			whenChallenge: &v1alpha1.ChallengeRequest{
				UID:          "test-UID",
				Key:          "test-key",
				DNSName:      "*.test.com",
				ResolvedZone: "test.com",
				ResolvedFQDN: "_acme-challenge.test.com.",
			},
			thenRecordCreateKey: "test-key",
			thenError:           "error creating record",
		},
	}
	for _, tc := range testCases {
		s.Run(tc.name, func() {
			s.setupMocks()
			zoneName := strings.TrimSuffix(tc.whenChallenge.ResolvedZone, ".")
			if tc.givenZones != nil {
				zoneReadList := dnsclient.ZoneReadList{
					Items: &tc.givenZones,
				}
				s.dnsAPIMock.EXPECT().GetZones(zoneName).Return(zoneReadList, tc.whenZonesReadError)
			}
			if tc.givenRecords != nil {
				recordName := strings.TrimSuffix(tc.whenChallenge.ResolvedFQDN, "."+tc.whenChallenge.ResolvedZone+".")
				s.dnsAPIMock.EXPECT().GetRecords("test-zone-id", recordName).Return(dnsclient.RecordReadList{
					Items: &tc.givenRecords,
				}, tc.whenRecordsReadError)
			}
			if tc.thenRecordCreateKey != "" {
				s.dnsAPIMock.EXPECT().CreateTXTRecord("test-zone-id", "_acme-challenge", tc.thenRecordCreateKey).
					Return(dnsclient.RecordRead{
						Id: ptr.To("test-record-id"),
					}, tc.whenRecordCreateError)
			}
			resolver := NewResolver(s.dnsAPIMock, s.logger)
			err := resolver.Present(tc.whenChallenge)
			if tc.thenError != "" {
				require.Error(s.T(), err)
				require.Equal(s.T(), tc.thenError, err.Error())
			} else {
				require.NoError(s.T(), err)
			}
		})
	}
}

func (s *ResolverTestSuite) TestCleanUp() {
	testCases := []struct {
		name                  string
		givenZones            []dnsclient.ZoneRead
		givenRecords          []dnsclient.RecordRead
		whenChallenge         *v1alpha1.ChallengeRequest
		whenZonesReadError    error
		whenRecordsReadError  error
		whenRecordDeleteError error
		thenDeleteRecordId    string
		thenError             string
	}{
		{
			name:       "no zones",
			givenZones: []dnsclient.ZoneRead{},
			whenChallenge: &v1alpha1.ChallengeRequest{
				UID:          "test-UID",
				Key:          "test-key",
				DNSName:      "*.test.com",
				ResolvedZone: "test.com",
				ResolvedFQDN: "_acme-challenge.test.com.",
			},
			thenError:          "", // no error
			thenDeleteRecordId: "", // no record to delete
		},
		{
			name: "zone exists, but no record",
			givenZones: []dnsclient.ZoneRead{
				{
					Id: ptr.To("test-zone-id"),
					Properties: &dnsclient.Zone{
						ZoneName: ptr.To("test.com"),
					},
					Type: ptr.To("NATIVE"),
				},
			},
			givenRecords: []dnsclient.RecordRead{},
			whenChallenge: &v1alpha1.ChallengeRequest{
				UID:          "test-UID",
				Key:          "test-key",
				DNSName:      "*.test.com",
				ResolvedZone: "test.com",
				ResolvedFQDN: "_acme-challenge.test.com.",
			},
			thenError:          "", // no error
			thenDeleteRecordId: "", // no record to delete
		},
		{
			name: "zone and record with same name exists, but has a different key",
			givenZones: []dnsclient.ZoneRead{
				{
					Id: ptr.To("test-zone-id"),
					Properties: &dnsclient.Zone{
						ZoneName: ptr.To("test.com"),
					},
					Type: ptr.To("NATIVE"),
				},
			},
			givenRecords: []dnsclient.RecordRead{
				{
					Id: ptr.To("test-record-id"),
					Properties: &dnsclient.Record{
						Name:    ptr.To("_acme-challenge"),
						Type:    typeTxtRecord,
						Content: ptr.To("different-key"),
					},
				},
			},
			whenChallenge: &v1alpha1.ChallengeRequest{
				UID:          "test-UID",
				Key:          "test-key",
				DNSName:      "*.test.com",
				ResolvedZone: "test.com",
				ResolvedFQDN: "_acme-challenge.test.com.",
			},
			thenError:          "", // no error
			thenDeleteRecordId: "", // no record to delete
		},
		{
			name:               "zone read error",
			givenZones:         []dnsclient.ZoneRead{},
			whenZonesReadError: fmt.Errorf("error fetching zones"),
			whenChallenge: &v1alpha1.ChallengeRequest{
				UID:          "test-UID",
				Key:          "test-key",
				DNSName:      "*.test.com",
				ResolvedZone: "test.com",
				ResolvedFQDN: "_acme-challenge.test.com.",
			},
			thenError: "error fetching zones",
		},
		{
			name: "record read error",
			givenZones: []dnsclient.ZoneRead{
				{
					Id: ptr.To("test-zone-id"),
					Properties: &dnsclient.Zone{
						ZoneName: ptr.To("test.com"),
					},
					Type: ptr.To("NATIVE"),
				},
			},
			givenRecords:         []dnsclient.RecordRead{},
			whenRecordsReadError: fmt.Errorf("error fetching records"),
			whenChallenge: &v1alpha1.ChallengeRequest{
				UID:          "test-UID",
				Key:          "test-key",
				DNSName:      "*.test.com",
				ResolvedZone: "test.com",
				ResolvedFQDN: "_acme-challenge.test.com.",
			},
			thenError: "error fetching records",
		},
		{
			name: "record delete error",
			givenZones: []dnsclient.ZoneRead{
				{
					Id: ptr.To("test-zone-id"),
					Properties: &dnsclient.Zone{
						ZoneName: ptr.To("test.com"),
					},
					Type: ptr.To("NATIVE"),
				},
			},
			givenRecords: []dnsclient.RecordRead{
				{
					Id: ptr.To("test-record-id"),
					Properties: &dnsclient.Record{
						Name:    ptr.To("_acme-challenge"),
						Type:    typeTxtRecord,
						Content: ptr.To("test-key"),
					},
				},
			},
			whenChallenge: &v1alpha1.ChallengeRequest{
				UID:          "test-UID",
				Key:          "test-key",
				DNSName:      "*.test.com",
				ResolvedZone: "test.com",
				ResolvedFQDN: "_acme-challenge.test.com.",
			},
			whenRecordDeleteError: fmt.Errorf("error deleting record"),
			thenError:             "error deleting record",
			thenDeleteRecordId:    "test-record-id",
		},
		{
			name: "record with key exists",
			givenZones: []dnsclient.ZoneRead{
				{
					Id: ptr.To("test-zone-id"),
					Properties: &dnsclient.Zone{
						ZoneName: ptr.To("test.com"),
					},
					Type: ptr.To("NATIVE"),
				},
			},
			givenRecords: []dnsclient.RecordRead{
				{
					Id: ptr.To("test-record-id"),
					Properties: &dnsclient.Record{
						Name:    ptr.To("_acme-challenge"),
						Type:    typeTxtRecord,
						Content: ptr.To("test-key"),
					},
				},
			},
			whenChallenge: &v1alpha1.ChallengeRequest{
				UID:          "test-UID",
				Key:          "test-key",
				DNSName:      "*.test.com",
				ResolvedZone: "test.com",
				ResolvedFQDN: "_acme-challenge.test.com.",
			},
			thenDeleteRecordId: "test-record-id",
		},
	}
	for _, tc := range testCases {
		s.Run(tc.name, func() {
			s.setupMocks()
			zoneName := strings.TrimSuffix(tc.whenChallenge.ResolvedZone, ".")
			zoneReadList := dnsclient.ZoneReadList{
				Items: &tc.givenZones,
			}
			s.dnsAPIMock.EXPECT().GetZones(zoneName).Return(zoneReadList, tc.whenZonesReadError)
			if len(tc.givenZones) > 0 {
				zoneId := *tc.givenZones[0].GetId()
				if tc.givenRecords != nil {
					recordName := strings.TrimSuffix(tc.whenChallenge.ResolvedFQDN, "."+tc.whenChallenge.ResolvedZone+".")
					s.dnsAPIMock.EXPECT().GetRecords(zoneId, recordName).Return(dnsclient.RecordReadList{
						Items: &tc.givenRecords,
					}, tc.whenRecordsReadError)
				}
				if tc.thenDeleteRecordId != "" {
					s.dnsAPIMock.EXPECT().DeleteRecord(zoneId, tc.thenDeleteRecordId).Return(tc.whenRecordDeleteError)
				}
			}
			resolver := NewResolver(s.dnsAPIMock, s.logger)
			err := resolver.CleanUp(tc.whenChallenge)
			if tc.thenError != "" {
				require.Error(s.T(), err)
				require.Equal(s.T(), tc.thenError, err.Error())
			} else {
				require.NoError(s.T(), err)
			}
		})
	}
}
