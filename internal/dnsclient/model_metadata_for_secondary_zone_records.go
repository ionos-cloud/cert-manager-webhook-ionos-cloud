/*
IONOS Cloud - DNS API

Cloud DNS service helps IONOS Cloud customers to automate DNS Zone and Record management.

API version: 1.17.0
Contact: support@cloud.ionos.com
*/

// Code generated by OpenAPI Generator (https://openapi-generator.tech); DO NOT EDIT.

package dnsclient

import (
	"bytes"
	"encoding/json"
	"fmt"
)

// checks if the MetadataForSecondaryZoneRecords type satisfies the MappedNullable interface at compile time
var _ MappedNullable = &MetadataForSecondaryZoneRecords{}

// MetadataForSecondaryZoneRecords Metadata for records of secondary zones.
type MetadataForSecondaryZoneRecords struct {
	// A fully qualified domain name. FQDN consists of two parts - the hostname and the domain name.
	Fqdn string `json:"fqdn"`
	// The ID (UUID) of the DNS zone of which record belongs to.
	ZoneId string `json:"zoneId"`
	// Indicates the root name (from the primary zone) for the record
	RootName string `json:"rootName"`
}

type _MetadataForSecondaryZoneRecords MetadataForSecondaryZoneRecords

// NewMetadataForSecondaryZoneRecords instantiates a new MetadataForSecondaryZoneRecords object
// This constructor will assign default values to properties that have it defined,
// and makes sure properties required by API are set, but the set of arguments
// will change when the set of required properties is changed
func NewMetadataForSecondaryZoneRecords(fqdn string, zoneId string, rootName string) *MetadataForSecondaryZoneRecords {
	this := MetadataForSecondaryZoneRecords{}
	this.Fqdn = fqdn
	this.ZoneId = zoneId
	this.RootName = rootName
	return &this
}

// NewMetadataForSecondaryZoneRecordsWithDefaults instantiates a new MetadataForSecondaryZoneRecords object
// This constructor will only assign default values to properties that have it defined,
// but it doesn't guarantee that properties required by API are set
func NewMetadataForSecondaryZoneRecordsWithDefaults() *MetadataForSecondaryZoneRecords {
	this := MetadataForSecondaryZoneRecords{}
	return &this
}

// GetFqdn returns the Fqdn field value
func (o *MetadataForSecondaryZoneRecords) GetFqdn() string {
	if o == nil {
		var ret string
		return ret
	}

	return o.Fqdn
}

// GetFqdnOk returns a tuple with the Fqdn field value
// and a boolean to check if the value has been set.
func (o *MetadataForSecondaryZoneRecords) GetFqdnOk() (*string, bool) {
	if o == nil {
		return nil, false
	}
	return &o.Fqdn, true
}

// SetFqdn sets field value
func (o *MetadataForSecondaryZoneRecords) SetFqdn(v string) {
	o.Fqdn = v
}

// GetZoneId returns the ZoneId field value
func (o *MetadataForSecondaryZoneRecords) GetZoneId() string {
	if o == nil {
		var ret string
		return ret
	}

	return o.ZoneId
}

// GetZoneIdOk returns a tuple with the ZoneId field value
// and a boolean to check if the value has been set.
func (o *MetadataForSecondaryZoneRecords) GetZoneIdOk() (*string, bool) {
	if o == nil {
		return nil, false
	}
	return &o.ZoneId, true
}

// SetZoneId sets field value
func (o *MetadataForSecondaryZoneRecords) SetZoneId(v string) {
	o.ZoneId = v
}

// GetRootName returns the RootName field value
func (o *MetadataForSecondaryZoneRecords) GetRootName() string {
	if o == nil {
		var ret string
		return ret
	}

	return o.RootName
}

// GetRootNameOk returns a tuple with the RootName field value
// and a boolean to check if the value has been set.
func (o *MetadataForSecondaryZoneRecords) GetRootNameOk() (*string, bool) {
	if o == nil {
		return nil, false
	}
	return &o.RootName, true
}

// SetRootName sets field value
func (o *MetadataForSecondaryZoneRecords) SetRootName(v string) {
	o.RootName = v
}

func (o MetadataForSecondaryZoneRecords) MarshalJSON() ([]byte, error) {
	toSerialize, err := o.ToMap()
	if err != nil {
		return []byte{}, err
	}
	return json.Marshal(toSerialize)
}

func (o MetadataForSecondaryZoneRecords) ToMap() (map[string]interface{}, error) {
	toSerialize := map[string]interface{}{}
	toSerialize["fqdn"] = o.Fqdn
	toSerialize["zoneId"] = o.ZoneId
	toSerialize["rootName"] = o.RootName
	return toSerialize, nil
}

func (o *MetadataForSecondaryZoneRecords) UnmarshalJSON(data []byte) (err error) {
	// This validates that all required properties are included in the JSON object
	// by unmarshalling the object into a generic map with string keys and checking
	// that every required field exists as a key in the generic map.
	requiredProperties := []string{
		"fqdn",
		"zoneId",
		"rootName",
	}

	allProperties := make(map[string]interface{})

	err = json.Unmarshal(data, &allProperties)

	if err != nil {
		return err
	}

	for _, requiredProperty := range requiredProperties {
		if _, exists := allProperties[requiredProperty]; !exists {
			return fmt.Errorf("no value given for required property %v", requiredProperty)
		}
	}

	varMetadataForSecondaryZoneRecords := _MetadataForSecondaryZoneRecords{}

	decoder := json.NewDecoder(bytes.NewReader(data))
	decoder.DisallowUnknownFields()
	err = decoder.Decode(&varMetadataForSecondaryZoneRecords)

	if err != nil {
		return err
	}

	*o = MetadataForSecondaryZoneRecords(varMetadataForSecondaryZoneRecords)

	return err
}

type NullableMetadataForSecondaryZoneRecords struct {
	value *MetadataForSecondaryZoneRecords
	isSet bool
}

func (v NullableMetadataForSecondaryZoneRecords) Get() *MetadataForSecondaryZoneRecords {
	return v.value
}

func (v *NullableMetadataForSecondaryZoneRecords) Set(val *MetadataForSecondaryZoneRecords) {
	v.value = val
	v.isSet = true
}

func (v NullableMetadataForSecondaryZoneRecords) IsSet() bool {
	return v.isSet
}

func (v *NullableMetadataForSecondaryZoneRecords) Unset() {
	v.value = nil
	v.isSet = false
}

func NewNullableMetadataForSecondaryZoneRecords(val *MetadataForSecondaryZoneRecords) *NullableMetadataForSecondaryZoneRecords {
	return &NullableMetadataForSecondaryZoneRecords{value: val, isSet: true}
}

func (v NullableMetadataForSecondaryZoneRecords) MarshalJSON() ([]byte, error) {
	return json.Marshal(v.value)
}

func (v *NullableMetadataForSecondaryZoneRecords) UnmarshalJSON(src []byte) error {
	v.isSet = true
	return json.Unmarshal(src, &v.value)
}
