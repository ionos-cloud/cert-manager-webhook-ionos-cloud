package clouddns

import (
	"context"
	"fmt"
	dnsclient "github.com/ionos-cloud/sdk-go-dns"
	"net/http"
)

const typeTxtRecord = "TXT"

type DNSAPI interface {
	GetZones(name string) (dnsclient.ZoneReadList, error)
	CreateZone(name string) (dnsclient.ZoneRead, error)
	GetRecords(zoneId string, name string) (dnsclient.RecordReadList, error)
	CreateTXTRecord(zoneId string, recordName string, content string) (dnsclient.RecordRead, error)
	DeleteRecord(zoneId string, recordId string) error
}

func CreateDNSAPI(client *dnsclient.APIClient) DNSAPI {
	return &APIClient{
		client: client,
	}
}

type APIClient struct {
	client *dnsclient.APIClient
}

func (c *APIClient) GetZones(name string) (dnsclient.ZoneReadList, error) {
	zoneList, resp, err := c.client.ZonesApi.ZonesGet(context.Background()).FilterZoneName(name).Execute()
	if err != nil {
		return dnsclient.ZoneReadList{}, err
	}
	if resp.StatusCode != http.StatusOK {
		return dnsclient.ZoneReadList{}, fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}
	return zoneList, nil
}

func (c *APIClient) CreateZone(name string) (dnsclient.ZoneRead, error) {
	zoneCreate := *dnsclient.NewZoneCreate(*dnsclient.NewZone(name))
	zone, resp, err := c.client.ZonesApi.ZonesPost(context.Background()).ZoneCreate(zoneCreate).Execute()
	if err != nil {
		return dnsclient.ZoneRead{}, err
	}
	if resp.StatusCode != http.StatusCreated {
		return dnsclient.ZoneRead{}, fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}
	return zone, nil
}

func (c *APIClient) GetRecords(zoneId string, name string) (dnsclient.RecordReadList, error) {
	recordList, resp, err := c.client.RecordsApi.RecordsGet(context.Background()).FilterZoneId(zoneId).FilterName(name).
		Execute()
	if err != nil {
		return dnsclient.RecordReadList{}, err
	}
	if resp.StatusCode != http.StatusOK {
		return dnsclient.RecordReadList{}, fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}
	return recordList, nil
}

func (c *APIClient) CreateTXTRecord(zoneId string, recordName string, content string) (dnsclient.RecordRead, error) {
	recordCreate := *dnsclient.NewRecordCreate(*dnsclient.NewRecord(recordName, typeTxtRecord, content)) // RecordCreate | record
	record, resp, err := c.client.RecordsApi.ZonesRecordsPost(context.Background(), zoneId).RecordCreate(recordCreate).Execute()
	if err != nil {
		return dnsclient.RecordRead{}, err
	}
	if resp.StatusCode != http.StatusCreated {
		return dnsclient.RecordRead{}, fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}
	return record, nil
}

func (c *APIClient) DeleteRecord(zoneId string, recordId string) error {
	_, resp, err := c.client.RecordsApi.ZonesRecordsDelete(context.Background(), zoneId, recordId).Execute()
	if err != nil {
		return err
	}
	if resp.StatusCode != http.StatusNoContent {
		return fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}
	return nil
}
