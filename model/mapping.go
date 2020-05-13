package model

import (
	"crypto/tls"
	"encoding/json"
	"fmt"
	"net"
	"strings"
)

var (
	// insecureTLSConfig is the tls configuration
	// used for mappings configured with
	// the InsecureSkipVerify.
	insecureTLSConfig = &tls.Config{
		InsecureSkipVerify: true,
	}
)

// Mapping represents a mapping between an external
// fully-qualified domain name (FQDN) and an
// internal hostname, protocol and port.
// mapping also contains sso information.
type Mapping struct {
	// Meta ...
	Meta `json:",inline"`
	// External is represents an external
	// hostname for an application.
	External string `json:"external"`
	// Domain is the external domain name
	// for an application.
	Domain *Domain `json:"domain"`
	// Internal is the internal hostname of the
	// application.
	Internal string `json:"internal"`
	// Port is the port in which the internal
	// application executes.
	Port int `json:"port"`
	// Protocol is the application protocol.
	Protocol Protocol `json:"protocol"`
	// SSOType is the type of sso configured
	// for the mapping.
	SSOType SSOType `json:"sso"`
	// SSOFormsTemplate is a template
	// for forms-sso.
	SSOFormsTemplate string `json:"sso_forms_tpl"`
	// SSOFormsURL is the url to
	// send the forms template to.
	SSOFormsURL string `json:"sso_forms_url"`
	// SSOPrivateKey is the users private
	// key for private-key-sso.
	SSOPrivateKey string `json:"-"`
	// SSOUseMappingCredentials maintains which
	// credentials should be used when performing
	// SSO, the user's or the mapping's.
	SSOUseMappingCredentials bool `json:"sso_use_mapping_credentials"`
	// SSOUsername is the mapping username to
	// use when SSOUseMappingCredentials = true.
	SSOUsername string `json:"sso_username"`
	// SSOPassword is the mapping password to
	// use when SSOUseMappingCredentials = true.
	SSOPassword string `json:"-"`
	// InsecureSkipVerify skips tls verification
	// for HTTPS mappings.
	InsecureSkipVerify bool `json:"insecure_skip_verify"`
}

// FQDN returns the fully-qualified domain name for mapping m.
func (m Mapping) FQDN() string { return fmt.Sprintf("%s.%s", m.External, m.Domain.Name) }

// InternalAddr returns the joint internal address associated with mapping m.
func (m *Mapping) InternalAddr() string {
	host := m.Internal
	// We assume that host is a literal IPv6 address
	// if host has colons:
	if strings.Contains(host, ":") {
		host = fmt.Sprintf("[%s]", host)
	}
	return fmt.Sprintf("%s:%d", host, m.Port)
}

// TLSConfig returns the tls context for mapping m,
// based on the InsecureSkipVerify parameter.
func (m *Mapping) TLSConfig() *tls.Config {
	if m.InsecureSkipVerify {
		return insecureTLSConfig
	}
	// tls cannot verify ip addresses,
	// so if the internal host is an ip address
	// we also need to turn off tls verification:
	if ip := net.ParseIP(m.Internal); ip != nil {
		return insecureTLSConfig
	}
	return nil
}

// Dial attempts to dial the mapping.
func (m *Mapping) Dial() (net.Conn, error) {
	if m.Protocol == HTTPS {
		return tls.Dial("tcp", m.InternalAddr(), m.TLSConfig())
	}
	return net.Dial("tcp", m.InternalAddr())
}

// MarshalJSON implements json.Marshaller to simplify the json that is returned to the client.
func (m *Mapping) MarshalJSON() ([]byte, error) {
	return json.Marshal(struct {
		Meta
		External              string   `json:"external"`
		Domain                int      `json:"domain"`
		Internal              string   `json:"internal"`
		Port                  int      `json:"port"`
		Protocol              Protocol `json:"protocol"`
		SSOType               SSOType  `json:"sso"`
		FormsTemplate         string   `json:"sso_forms_tpl"`
		FormsURL              string   `json:"sso_forms_url"`
		UseMappingCredentials bool     `json:"sso_use_mapping_credentials"`
		Username              string   `json:"sso_username"`
		InsecureSkipVerify    bool     `json:"insecure_skip_verify"`
	}{
		m.Meta,
		m.External,
		int(m.Domain.Id),
		m.Internal,
		m.Port,
		m.Protocol,
		m.SSOType,
		m.SSOFormsTemplate,
		m.SSOFormsURL,
		m.SSOUseMappingCredentials,
		m.SSOUsername,
		m.InsecureSkipVerify,
	})
}

// NewMapping returns a new mapping object.
func NewMapping(name string) *Mapping {
	return &Mapping{Meta: Meta{Kind: "Mapping", Name: name}}
}
