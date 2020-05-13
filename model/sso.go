package model

import "encoding/json"

const (
	None SSOType = iota
	Basic
	Windows
	Forms
	PrivateKey
)

// SSOType represents type of sso configured for a mapping.
type SSOType int

// SupportedSSO is the list of supported sso types.
var SupportedSSO = []struct {
	Id   Identity `json:"id"`
	Name string   `json:"name"`
}{
	{Identity(None), "None"},
	{Identity(Basic), "Basic"},
	{Identity(Windows), "Windows"},
	{Identity(Forms), "Forms"},
	{Identity(PrivateKey), "PrivateKey"},
}

// MarshalJSON marshals the sso-type with its name.
func (sso SSOType) MarshalJSON() (buf []byte, err error) {
	return json.Marshal(struct {
		Id   Identity `json:"id"`
		Name string   `json:"name"`
	}{
		Identity(sso),
		sso.String(),
	})
}

// String returns the textual representation of the sso.
func (sso SSOType) String() string {
	switch sso {
	case None:
		return "None"
	case Basic:
		return "Basic"
	case Windows:
		return "Windows"
	case Forms:
		return "Forms"
	case PrivateKey:
		return "PrivateKey"
	default:
		return "UNKNOWN"
	}
}
