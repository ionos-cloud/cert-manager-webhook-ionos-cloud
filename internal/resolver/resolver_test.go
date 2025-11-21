//go:build unit

package resolver

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"testing"

	"github.com/cert-manager/cert-manager/pkg/acme/webhook/apis/acme/v1alpha1"
	"github.com/ionos-cloud/cert-manager-webhook-ionos-cloud/internal/clouddns"
	"github.com/ionos-cloud/cert-manager-webhook-ionos-cloud/internal/mocks"
	dnsclient "github.com/ionos-cloud/sdk-go-dns"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
	"go.uber.org/zap"
	corev1 "k8s.io/api/core/v1"
	apiextensionsv1 "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/rest"
)

var (
	typeTxtRecord = toPTR(dnsclient.RecordType("TXT"))
	errK8Client   = errors.New("k8 client error")
)

const testNamespace = "unit-test"

type ResolverTestSuite struct {
	suite.Suite
	dnsAPIMock *mocks.DNSAPI
	k8Client   *mocks.K8Client
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
	s.k8Client = mocks.NewK8Client(s.T())
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
		whenK8ClientError     error
		thenError             string
		whenConfigParseError  bool
		thenRecordCreateKey   string
	}{
		{
			name:       "invalid config json",
			givenZones: []dnsclient.ZoneRead{},
			whenChallenge: &v1alpha1.ChallengeRequest{
				UID:          "test-UID",
				Key:          "test-key",
				DNSName:      "*.test.com",
				ResolvedZone: "test.com.",
				ResolvedFQDN: "_acme-challenge.test.com.",
				Config:       &apiextensionsv1.JSON{Raw: []byte("{")},
			},
			whenConfigParseError: true,
			thenError:            "failed to create IONOS Cloud API client: failed to parse config: unexpected end of JSON input",
		},
		{
			name:       "k8 client error",
			givenZones: []dnsclient.ZoneRead{},
			whenChallenge: &v1alpha1.ChallengeRequest{
				UID:          "test-UID",
				Key:          "test-key",
				DNSName:      "*.test.com",
				ResolvedZone: "test.com.",
				ResolvedFQDN: "_acme-challenge.test.com.",
			},
			whenK8ClientError: errK8Client,
			thenError:         "failed to create IONOS Cloud API client: failed to get secret cert-manager-webhook-ionos-cloud from namespace unit-test: k8 client error",
		},
		{
			name:       "no zones",
			givenZones: []dnsclient.ZoneRead{},
			whenChallenge: &v1alpha1.ChallengeRequest{
				UID:          "test-UID",
				Key:          "test-key",
				DNSName:      "*.test.com",
				ResolvedZone: "test.com.",
				ResolvedFQDN: "_acme-challenge.test.com.",
			},
			thenError: "zone 'test.com' not found",
		},
		{
			name: "zone already exists",
			givenZones: []dnsclient.ZoneRead{
				{
					Id: toPTR("test-zone-id"),
					Properties: &dnsclient.Zone{
						ZoneName: toPTR("test.com"),
					},
					Type: toPTR("NATIVE"),
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
			thenRecordCreateKey: "test-key",
		},
		{
			name: "record with the same name and key already exists",
			givenZones: []dnsclient.ZoneRead{
				{
					Id: toPTR("test-zone-id"),
					Properties: &dnsclient.Zone{
						ZoneName: toPTR("test.com"),
					},
					Type: toPTR("NATIVE"),
				},
			},
			givenRecords: []dnsclient.RecordRead{
				{
					Id: toPTR("test-record-id"),
					Properties: &dnsclient.Record{
						Name:    toPTR("_acme-challenge"),
						Type:    typeTxtRecord,
						Content: toPTR("test-key"),
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
			thenRecordCreateKey: "", // no record should be created
		},
		{
			name: "record with the same name but different key already exists",
			givenZones: []dnsclient.ZoneRead{
				{
					Id: toPTR("test-zone-id"),
					Properties: &dnsclient.Zone{
						ZoneName: toPTR("test.com"),
					},
					Type: toPTR("NATIVE"),
				},
			},
			givenRecords: []dnsclient.RecordRead{
				{
					Id: toPTR("test-record-id"),
					Properties: &dnsclient.Record{
						Name:    toPTR("_acme-challenge"),
						Type:    typeTxtRecord,
						Content: toPTR("different-key"),
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
					Id: toPTR("test-zone-id"),
					Properties: &dnsclient.Zone{
						ZoneName: toPTR("test.com"),
					},
					Type: toPTR("NATIVE"),
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
					Id: toPTR("test-zone-id"),
					Properties: &dnsclient.Zone{
						ZoneName: toPTR("test.com"),
					},
					Type: toPTR("NATIVE"),
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
			if !tc.whenConfigParseError {
				setUpK8ClientExpectations(s.k8Client, tc.whenK8ClientError, s.T())
			}
			if !tc.whenConfigParseError && tc.whenK8ClientError == nil {
				zoneName := strings.TrimSuffix(tc.whenChallenge.ResolvedZone, ".")
				if tc.givenZones != nil {
					zoneReadList := dnsclient.ZoneReadList{
						Items: &tc.givenZones,
					}
					s.dnsAPIMock.EXPECT().GetZones(zoneName).Return(zoneReadList, tc.whenZonesReadError)
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
							Id: toPTR("test-record-id"),
						}, tc.whenRecordCreateError)
				}
			}

			resolver := NewResolver(testNamespace, createTestK8Factory(s.k8Client), createTestDNSFactory(s.dnsAPIMock), s.logger)
			resolver.Initialize(&rest.Config{}, nil)
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
		whenK8ClientError     error
		whenConfigParseError  bool
		thenDeleteRecordId    string
		thenError             string
	}{
		{
			name:       "invalid config json",
			givenZones: []dnsclient.ZoneRead{},
			whenChallenge: &v1alpha1.ChallengeRequest{
				UID:          "test-UID",
				Key:          "test-key",
				DNSName:      "*.test.com",
				ResolvedZone: "test.com.",
				ResolvedFQDN: "_acme-challenge.test.com.",
				Config:       &apiextensionsv1.JSON{Raw: []byte("{")},
			},
			whenConfigParseError: true,
			thenError:            "failed to create IONOS Cloud API client: failed to parse config: unexpected end of JSON input",
		},
		{
			name:       "k8 client error",
			givenZones: []dnsclient.ZoneRead{},
			whenChallenge: &v1alpha1.ChallengeRequest{
				UID:          "test-UID",
				Key:          "test-key",
				DNSName:      "*.test.com",
				ResolvedZone: "test.com.",
				ResolvedFQDN: "_acme-challenge.test.com.",
			},
			whenK8ClientError: errK8Client,
			thenError:         "failed to create IONOS Cloud API client: failed to get secret cert-manager-webhook-ionos-cloud from namespace unit-test: k8 client error",
		},
		{
			name:       "no zones",
			givenZones: []dnsclient.ZoneRead{},
			whenChallenge: &v1alpha1.ChallengeRequest{
				UID:          "test-UID",
				Key:          "test-key",
				DNSName:      "*.test.com",
				ResolvedZone: "test.com.",
				ResolvedFQDN: "_acme-challenge.test.com.",
			},
			thenError:          "", // no error
			thenDeleteRecordId: "", // no record to delete
		},
		{
			name: "zone exists, but no record",
			givenZones: []dnsclient.ZoneRead{
				{
					Id: toPTR("test-zone-id"),
					Properties: &dnsclient.Zone{
						ZoneName: toPTR("test.com"),
					},
					Type: toPTR("NATIVE"),
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
			thenError:          "", // no error
			thenDeleteRecordId: "", // no record to delete
		},
		{
			name: "zone and record with same name exists, but has a different key",
			givenZones: []dnsclient.ZoneRead{
				{
					Id: toPTR("test-zone-id"),
					Properties: &dnsclient.Zone{
						ZoneName: toPTR("test.com"),
					},
					Type: toPTR("NATIVE"),
				},
			},
			givenRecords: []dnsclient.RecordRead{
				{
					Id: toPTR("test-record-id"),
					Properties: &dnsclient.Record{
						Name:    toPTR("_acme-challenge"),
						Type:    typeTxtRecord,
						Content: toPTR("different-key"),
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
				ResolvedZone: "test.com.",
				ResolvedFQDN: "_acme-challenge.test.com.",
			},
			thenError: "error fetching zones",
		},
		{
			name: "record read error",
			givenZones: []dnsclient.ZoneRead{
				{
					Id: toPTR("test-zone-id"),
					Properties: &dnsclient.Zone{
						ZoneName: toPTR("test.com"),
					},
					Type: toPTR("NATIVE"),
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
			name: "record delete error",
			givenZones: []dnsclient.ZoneRead{
				{
					Id: toPTR("test-zone-id"),
					Properties: &dnsclient.Zone{
						ZoneName: toPTR("test.com"),
					},
					Type: toPTR("NATIVE"),
				},
			},
			givenRecords: []dnsclient.RecordRead{
				{
					Id: toPTR("test-record-id"),
					Properties: &dnsclient.Record{
						Name:    toPTR("_acme-challenge"),
						Type:    typeTxtRecord,
						Content: toPTR("test-key"),
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
			whenRecordDeleteError: fmt.Errorf("error deleting record"),
			thenError:             "error deleting record",
			thenDeleteRecordId:    "test-record-id",
		},
		{
			name: "record with key exists",
			givenZones: []dnsclient.ZoneRead{
				{
					Id: toPTR("test-zone-id"),
					Properties: &dnsclient.Zone{
						ZoneName: toPTR("test.com"),
					},
					Type: toPTR("NATIVE"),
				},
			},
			givenRecords: []dnsclient.RecordRead{
				{
					Id: toPTR("test-record-id"),
					Properties: &dnsclient.Record{
						Name:    toPTR("_acme-challenge"),
						Type:    typeTxtRecord,
						Content: toPTR("test-key"),
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
			thenDeleteRecordId: "test-record-id",
		},
	}
	for _, tc := range testCases {
		s.Run(tc.name, func() {
			s.setupMocks()
			if !tc.whenConfigParseError {
				setUpK8ClientExpectations(s.k8Client, tc.whenK8ClientError, s.T())
				if tc.whenK8ClientError == nil {
					zoneName := strings.TrimSuffix(tc.whenChallenge.ResolvedZone, ".")
					zoneReadList := dnsclient.ZoneReadList{
						Items: &tc.givenZones,
					}
					s.dnsAPIMock.EXPECT().GetZones(zoneName).Return(zoneReadList, tc.whenZonesReadError)
					if len(tc.givenZones) > 0 {
						zoneId := *tc.givenZones[0].GetId()
						if tc.givenRecords != nil {
							recordName := strings.TrimSuffix(tc.whenChallenge.ResolvedFQDN, "."+tc.whenChallenge.ResolvedZone)
							s.dnsAPIMock.EXPECT().GetRecords(zoneId, recordName).Return(dnsclient.RecordReadList{
								Items: &tc.givenRecords,
							}, tc.whenRecordsReadError)
						}
						if tc.thenDeleteRecordId != "" {
							s.dnsAPIMock.EXPECT().DeleteRecord(zoneId, tc.thenDeleteRecordId).Return(tc.whenRecordDeleteError)
						}
					}
				}
			}
			resolver := NewResolver(testNamespace, createTestK8Factory(s.k8Client), createTestDNSFactory(s.dnsAPIMock), s.logger)
			resolver.Initialize(&rest.Config{}, nil)
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

func createTestDNSFactory(dnsAPIMock *mocks.DNSAPI) DNSAPIFactory {
	return func(_ string) clouddns.DNSAPI {
		return dnsAPIMock
	}
}

func createTestK8Factory(k8Client *mocks.K8Client) K8ClientFactory {
	return func(_ *rest.Config) (K8Client, error) {
		return k8Client, nil
	}
}

func setUpK8ClientExpectations(k8Client *mocks.K8Client, err error, t *testing.T) {
	secretsInterface := mocks.NewSecretInterface(t)
	coreV1Interface := mocks.NewCoreV1Interface(t)
	k8Secret := &corev1.Secret{}
	k8Secret.StringData = map[string]string{defaultAuthTokenSecretKey: "token"}
	secretsInterface.EXPECT().Get(context.Background(), defaultSecretName, v1.GetOptions{}).
		Return(k8Secret, err)

	coreV1Interface.EXPECT().Secrets(testNamespace).Return(secretsInterface)

	k8Client.EXPECT().CoreV1().Return(coreV1Interface)
}

func toPTR[C any](c C) *C {
	return &c
}
