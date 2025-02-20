/*
IONOS Cloud - DNS API

Cloud DNS service helps IONOS Cloud customers to automate DNS Zone and Record management.

API version: 1.17.0
Contact: support@cloud.ionos.com
*/

// Code generated by OpenAPI Generator (https://openapi-generator.tech); DO NOT EDIT.

package dnsclient

import (
	"encoding/json"
	"fmt"
)

// KskBits Key signing key length in bits. kskBits >= zskBits
type KskBits int32

// List of kskBits
const (
	_1024 KskBits = 1024
	_2048 KskBits = 2048
	_4096 KskBits = 4096
)

// All allowed values of KskBits enum
var AllowedKskBitsEnumValues = []KskBits{
	1024,
	2048,
	4096,
}

func (v *KskBits) UnmarshalJSON(src []byte) error {
	var value int32
	err := json.Unmarshal(src, &value)
	if err != nil {
		return err
	}
	enumTypeValue := KskBits(value)
	for _, existing := range AllowedKskBitsEnumValues {
		if existing == enumTypeValue {
			*v = enumTypeValue
			return nil
		}
	}

	return fmt.Errorf("%+v is not a valid KskBits", value)
}

// NewKskBitsFromValue returns a pointer to a valid KskBits
// for the value passed as argument, or an error if the value passed is not allowed by the enum
func NewKskBitsFromValue(v int32) (*KskBits, error) {
	ev := KskBits(v)
	if ev.IsValid() {
		return &ev, nil
	} else {
		return nil, fmt.Errorf("invalid value '%v' for KskBits: valid values are %v", v, AllowedKskBitsEnumValues)
	}
}

// IsValid return true if the value is valid for the enum, false otherwise
func (v KskBits) IsValid() bool {
	for _, existing := range AllowedKskBitsEnumValues {
		if existing == v {
			return true
		}
	}
	return false
}

// Ptr returns reference to kskBits value
func (v KskBits) Ptr() *KskBits {
	return &v
}

type NullableKskBits struct {
	value *KskBits
	isSet bool
}

func (v NullableKskBits) Get() *KskBits {
	return v.value
}

func (v *NullableKskBits) Set(val *KskBits) {
	v.value = val
	v.isSet = true
}

func (v NullableKskBits) IsSet() bool {
	return v.isSet
}

func (v *NullableKskBits) Unset() {
	v.value = nil
	v.isSet = false
}

func NewNullableKskBits(val *KskBits) *NullableKskBits {
	return &NullableKskBits{value: val, isSet: true}
}

func (v NullableKskBits) MarshalJSON() ([]byte, error) {
	return json.Marshal(v.value)
}

func (v *NullableKskBits) UnmarshalJSON(src []byte) error {
	v.isSet = true
	return json.Unmarshal(src, &v.value)
}
