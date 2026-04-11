package api

import "encoding/json"

// Device represents a TCL IoT device.
type Device struct {
	DeviceID     string  `json:"deviceId"`
	NickName     string  `json:"nickName,omitempty"`
	Category     string  `json:"category,omitempty"`
	LocationName string  `json:"locationName,omitempty"`
	IsOnline     string  `json:"isOnline,omitempty"`
	ProductKey   string  `json:"productKey,omitempty"`
	Family       *Family `json:"family,omitempty"`
}

// IsDeviceOnline reports whether the device is currently online.
func (d *Device) IsDeviceOnline() bool {
	return d.IsOnline == "1"
}

// Family represents home/family info attached to a device.
type Family struct {
	FamilyID   string `json:"familyId,omitempty"`
	FamilyName string `json:"familyName,omitempty"`
}

// TSLProperty represents a single property in the Thing Model.
type TSLProperty struct {
	Identifier string   `json:"identifier"`
	Name       string   `json:"name"`
	AccessMode string   `json:"accessMode"`
	DataType   DataType `json:"dataType"`
}

// DataType describes the type and constraints of a TSL property.
// Specs is kept as json.RawMessage because the TCL API returns different formats
// depending on the property type (flat map for int/float, nested object for struct,
// array for enum, etc.). Preserving raw JSON ensures no parsing failures and
// lossless pass-through for --json output consumed by AI agents.
type DataType struct {
	Type  string          `json:"type"`
	Specs json.RawMessage `json:"specs,omitempty"`
}

// TSLModel represents the Thing Service Language model for a product.
type TSLModel struct {
	ProductKey string        `json:"productKey,omitempty"`
	Properties []TSLProperty `json:"properties,omitempty"`
}

// apiResponse is the generic TCL API response envelope.
// TCL API is inconsistent across endpoints: some return "message", others "msg";
// some return "code" as a number, others as a string; some use "success" bool,
// others use "status" int. All fields are kept to handle every endpoint correctly.
type apiResponse struct {
	Code         interface{}     `json:"code,omitempty"`
	Success      bool            `json:"success,omitempty"`
	Message      string          `json:"message,omitempty"`
	Msg          string          `json:"msg,omitempty"`
	Data         json.RawMessage `json:"data,omitempty"`
	Status       int             `json:"status,omitempty"`
	AccessToken  string          `json:"accessToken,omitempty"`
	RefreshToken string          `json:"refreshToken,omitempty"`
}
