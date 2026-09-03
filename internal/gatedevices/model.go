package gatedevices
type Device struct{ID int64 `json:"id"`;DeviceKey string `json:"device_key"`;Name string `json:"name"`;Description *string `json:"description,omitempty"`;GateID int64 `json:"gate_id"`;Active bool `json:"active"`;LastSeenAt *string `json:"last_seen_at,omitempty"`}
type Input struct{DeviceKey string `json:"device_key"`;Name string `json:"name"`;Description *string `json:"description"`;GateID int64 `json:"gate_id"`;Active *bool `json:"active"`}
type ActivateInput struct{DeviceKey string `json:"device_key"`;DeviceSecret string `json:"device_secret"`}