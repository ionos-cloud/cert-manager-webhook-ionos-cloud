package resolver

import (
	"fmt"
	"github.com/cert-manager/cert-manager/pkg/acme/webhook/apis/acme/v1alpha1"
	"github.com/ionos-cloud/cert-manager-webhook-ionos-cloud/internal/clouddns/mocks"
	dnsclient "github.com/ionos-cloud/sdk-go-dns"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
	"go.uber.org/zap"
	"k8s.io/utils/ptr"
	"strings"
	"testing"
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
		whenZoneCreateError   error
		whenRecordsReadError  error
		whenRecordCreateError error
		thenError             string
		thenZoneCreate        bool
		thenRecordCreateKey   string
	}{
		{
			name:         "no zones",
			givenZones:   []dnsclient.ZoneRead{},
			givenRecords: []dnsclient.RecordRead{},
			whenChallenge: &v1alpha1.ChallengeRequest{
				UID:          "test-UID",
				Key:          "test-key",
				DNSName:      "*.test.com",
				ResolvedZone: "test.com.",
				ResolvedFQDN: "_acme-challenge.test.com.",
			},
			thenZoneCreate:      true,
			thenRecordCreateKey: "test-key",
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
				ResolvedZone: "test.com.",
				ResolvedFQDN: "_acme-challenge.test.com.",
			},
			thenZoneCreate:      false,
			thenRecordCreateKey: "test-key",
		},
		{
			name: "record with the same name already exists",
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
				ResolvedZone: "test.com.",
				ResolvedFQDN: "_acme-challenge.test.com.",
			},
			thenZoneCreate:      false,
			thenRecordCreateKey: "", // no record should be created
		},
		{
			name:               "error fetching zones",
			givenZones:         []dnsclient.ZoneRead{},
			whenZonesReadError: fmt.Errorf("error fetching zones"),
			whenChallenge: &v1alpha1.ChallengeRequest{
				UID:          "test-UID",
				Key:          "test-key",
				DNSName:      "*.test.com",
				ResolvedZone: "test.com.",
				ResolvedFQDN: "_acme-challenge.test.com.",
			},
			thenError: "error fetching zones",
		},
		{
			name:                "error creating zone",
			givenZones:          []dnsclient.ZoneRead{},
			whenZoneCreateError: fmt.Errorf("error creating zone"),
			whenChallenge: &v1alpha1.ChallengeRequest{
				UID:          "test-UID",
				Key:          "test-key",
				DNSName:      "*.test.com",
				ResolvedZone: "test.com.",
			},
			thenZoneCreate: true,
			thenError:      "error creating zone",
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
				ResolvedZone: "test.com.",
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
				ResolvedZone: "test.com.",
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
			if tc.thenZoneCreate {
				s.dnsAPIMock.EXPECT().CreateZone(zoneName).Return(dnsclient.ZoneRead{
					Id: ptr.To("test-zone-id"),
				}, tc.whenZoneCreateError)
			}
			if tc.givenRecords != nil {
				recordName := strings.TrimSuffix(tc.whenChallenge.ResolvedFQDN, "."+tc.whenChallenge.ResolvedZone)
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

//func Test_Present(t *testing.T) {
//	logger, err := zap.NewDevelopment()
//	require.NoError(t, err)
//
//	zonesAPIMock := mocks.NewZonesAPI(t)
//	apiZonesRequest := dnsclient.ApiZonesGetRequest{
//		ApiService: zonesAPIMock,
//	}
//	zonesAPIMock.EXPECT().ZonesGet(context.Background()).Return(apiZonesRequest)
//	zoneReadList := &dnsclient.ZoneReadList{
//		Items: []dnsclient.ZoneRead{
//			{
//				Id: "test",
//			},
//		},
//	}
//	resp := &http.Response{
//		StatusCode: 200,
//	}
//	apiZonesRequest = apiZonesRequest.FilterZoneName("test")
//	zonesAPIMock.EXPECT().ZonesGetExecute(apiZonesRequest).Return(zoneReadList, resp, nil)
//
//	mockClient := &dnsclient.APIClient{
//		ZonesAPI:   zonesAPIMock,
//		RecordsAPI: mocks.NewRecordsAPI(t),
//	}
//
//	resolver := NewResolver(mockClient, logger)
//
//	challenge := &v1alpha1.ChallengeRequest{
//		UID:          "test",
//		Key:          "test",
//		DNSName:      "test",
//		ResolvedZone: "test",
//		ResolvedFQDN: "test",
//	}
//
//	err = resolver.Present(challenge)
//	require.NoError(t, err)
//
//}
