package resolver

import (
	"context"
	"fmt"
	"github.com/cert-manager/cert-manager/pkg/acme/webhook/apis/acme/v1alpha1"
	"github.com/ionos-cloud/cert-manager-webhook-ionos-cloud/internal/dnsclient"
	"github.com/ionos-cloud/cert-manager-webhook-ionos-cloud/internal/dnsclient/mocks"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
	"go.uber.org/zap"
	"net/http"
	"strings"
	"testing"
)

//type apiZonePostRequestMatcher struct {
//	expected dnsclient.ApiZonesPostRequest
//	message  string
//}
//
//func (e *apiZonePostRequestMatcher) Matches(x interface{}) bool {
//	actual, ok := x.(dnsclient.ApiZonesPostRequest)
//	if !ok {
//		return false
//	}
//	if !e.matchZoneCreate(actual) {
//
//
//
//	return true
//}

type ResolverTestSuite struct {
	suite.Suite
	apiClient *dnsclient.APIClient
	logger    *zap.Logger
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
	zonesAPIMock := mocks.NewZonesAPI(s.T())
	recordsAPIMock := mocks.NewRecordsAPI(s.T())
	s.apiClient = &dnsclient.APIClient{
		ZonesAPI:   zonesAPIMock,
		RecordsAPI: recordsAPIMock,
	}
	s.logger.Debug("apiClient with mocks is created")
}

func (s *ResolverTestSuite) apiZonesGetRequest() dnsclient.ApiZonesGetRequest {
	return dnsclient.ApiZonesGetRequest{
		ApiService: s.apiClient.ZonesAPI,
	}
}

func (s *ResolverTestSuite) apiZonesPostRequest() dnsclient.ApiZonesPostRequest {
	return dnsclient.ApiZonesPostRequest{
		ApiService: s.apiClient.ZonesAPI,
	}
}

func (s *ResolverTestSuite) apiRecordsGetRequest() dnsclient.ApiRecordsGetRequest {
	return dnsclient.ApiRecordsGetRequest{
		ApiService: s.apiClient.RecordsAPI,
	}
}

func (s *ResolverTestSuite) apiRecordsPostRequest() dnsclient.ApiZonesRecordsPostRequest {
	return dnsclient.ApiZonesRecordsPostRequest{
		ApiService: s.apiClient.RecordsAPI,
	}
}

func (s *ResolverTestSuite) zonesAPIMock() *mocks.ZonesAPI {
	return s.apiClient.ZonesAPI.(*mocks.ZonesAPI)
}

func (s *ResolverTestSuite) recordsAPIMock() *mocks.RecordsAPI {
	return s.apiClient.RecordsAPI.(*mocks.RecordsAPI)
}

func (s *ResolverTestSuite) TestPresent() {
	testCases := []struct {
		name                     string
		givenZones               []dnsclient.ZoneRead
		givenRecords             []dnsclient.RecordRead
		whenChallenge            *v1alpha1.ChallengeRequest
		whenZonesReadResponse    *http.Response
		whenZonesReadError       error
		whenZoneCreateResponse   *http.Response
		whenZoneCreateError      error
		whenRecordsReadResponse  *http.Response
		whenRecordsReadError     error
		whenRecordCreateResponse *http.Response
		whenRecordCreateError    error
		thenError                string
		thenZoneCreate           *dnsclient.ZoneCreate
		thenRecordCreate         *dnsclient.RecordCreate
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
			thenZoneCreate: dnsclient.NewZoneCreate(*dnsclient.NewZone("test.com")),
			thenRecordCreate: dnsclient.NewRecordCreate(*dnsclient.NewRecord("_acme-challenge", typeTxtRecord,
				"test-key")),
		},
		{
			name: "zone already exists",
			givenZones: []dnsclient.ZoneRead{
				{
					Id: "test-zone-id",
					Properties: dnsclient.Zone{
						ZoneName: "test.com",
					},
					Type: "NATIVE",
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
			thenRecordCreate: dnsclient.NewRecordCreate(*dnsclient.NewRecord("_acme-challenge", typeTxtRecord,
				"test-key")),
		},
		{
			name: "record with the same name already exists",
			givenZones: []dnsclient.ZoneRead{
				{
					Id: "test-zone-id",
					Properties: dnsclient.Zone{
						ZoneName: "test.com",
					},
					Type: "NATIVE",
				},
			},
			givenRecords: []dnsclient.RecordRead{
				{
					Id: "test-record-id",
					Properties: dnsclient.Record{
						Name:    "_acme-challenge",
						Type:    typeTxtRecord,
						Content: "test-key",
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
			thenZoneCreate: dnsclient.NewZoneCreate(*dnsclient.NewZone("test.com")),
			thenError:      "error creating zone",
		},
		{
			name: "error fetching records",
			givenZones: []dnsclient.ZoneRead{
				{
					Id: "test-zone-id",
					Properties: dnsclient.Zone{
						ZoneName: "test.com",
					},
					Type: "NATIVE",
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
					Id: "test-zone-id",
					Properties: dnsclient.Zone{
						ZoneName: "test.com",
					},
					Type: "NATIVE",
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
			thenRecordCreate: dnsclient.NewRecordCreate(*dnsclient.NewRecord("_acme-challenge", typeTxtRecord,
				"test-key")),
			thenError: "error creating record",
		},
	}
	for _, tc := range testCases {
		s.Run(tc.name, func() {
			s.setupMocks()
			if tc.givenZones != nil {
				apiZonesGetRequest := s.apiZonesGetRequest()
				s.zonesAPIMock().EXPECT().ZonesGet(context.Background()).Return(apiZonesGetRequest)
				zoneReadList := &dnsclient.ZoneReadList{
					Items: tc.givenZones,
				}
				zonesGetResponse := &http.Response{
					StatusCode: 200,
				}
				if tc.whenZonesReadResponse != nil {
					zonesGetResponse = tc.whenZonesReadResponse
				}
				filterName := strings.TrimSuffix(tc.whenChallenge.ResolvedZone, ".")
				apiZonesGetRequest = apiZonesGetRequest.FilterZoneName(filterName)
				s.zonesAPIMock().EXPECT().ZonesGetExecute(apiZonesGetRequest).Return(zoneReadList, zonesGetResponse, tc.whenZonesReadError)
			}
			if tc.thenZoneCreate != nil {
				zoneCreateResponse := &http.Response{
					StatusCode: http.StatusCreated,
				}
				if tc.whenZoneCreateResponse != nil {
					zoneCreateResponse = tc.whenZoneCreateResponse
				}
				apiZoneCreateRequest := s.apiZonesPostRequest()
				s.zonesAPIMock().EXPECT().ZonesPost(context.Background()).Return(apiZoneCreateRequest)
				apiZoneCreateRequest = apiZoneCreateRequest.ZoneCreate(*tc.thenZoneCreate)
				s.zonesAPIMock().EXPECT().ZonesPostExecute(apiZoneCreateRequest).Return(&dnsclient.ZoneRead{
					Id: "test-zone-id",
				},
					zoneCreateResponse, tc.whenZoneCreateError)
			}
			if tc.givenRecords != nil {
				apiRecordsGetRequest := s.apiRecordsGetRequest()
				recordName := strings.TrimSuffix(tc.whenChallenge.ResolvedFQDN, "."+tc.whenChallenge.ResolvedZone)
				s.recordsAPIMock().EXPECT().RecordsGet(context.Background()).Return(apiRecordsGetRequest)
				apiRecordsGetRequest = apiRecordsGetRequest.FilterZoneId("test-zone-id").
					FilterName(recordName).FilterType(typeTxtRecord)
				recordsGetResponse := &http.Response{
					StatusCode: 200,
				}
				if tc.whenRecordsReadResponse != nil {
					recordsGetResponse = tc.whenRecordsReadResponse
				}
				s.recordsAPIMock().EXPECT().RecordsGetExecute(apiRecordsGetRequest).Return(&dnsclient.RecordReadList{
					Items: tc.givenRecords,
				}, recordsGetResponse, tc.whenRecordsReadError)
			}
			if tc.thenRecordCreate != nil {
				apiRecordsPostRequest := s.apiRecordsPostRequest()
				s.recordsAPIMock().EXPECT().ZonesRecordsPost(context.Background(), "test-zone-id").Return(apiRecordsPostRequest)
				apiRecordsPostRequest = apiRecordsPostRequest.RecordCreate(*tc.thenRecordCreate)
				recordCreateResponse := &http.Response{
					StatusCode: http.StatusCreated,
				}
				if tc.whenRecordCreateResponse != nil {
					recordCreateResponse = tc.whenRecordCreateResponse
				}
				s.recordsAPIMock().EXPECT().ZonesRecordsPostExecute(apiRecordsPostRequest).Return(&dnsclient.RecordRead{
					Id: "test-record-id",
				}, recordCreateResponse, tc.whenRecordCreateError)
			}

			resolver := NewResolver(s.apiClient, s.logger)
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
