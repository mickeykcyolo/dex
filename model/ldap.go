package model

import (
	"fmt"
	"github.com/jtblin/go-ldap-client"
	"strings"
)

// AutoCreateShadowUsersLevel represents the
// level at which shadow users are automatically
// created from an IDP.
type AutoCreateShadowUsersLevel int

const (
	// Disabled does not create shadow users.
	Disabled AutoCreateShadowUsersLevel = iota
	// CreateDisabled creates shadow users as disabled.
	// note: this is the default.
	CreateDisabled
	// CreateEnabled creates shadow users as enabled.
	CreateEnabled
)

// LDAP represents an LDAP directory.
type LDAP struct {
	// Meta ...
	Meta `json:",inline"`
	// Attributes are which ldap attributes
	// should be request from the ldap host.
	Attributes string `json:"attributes"`
	// BaseDN is the base of the ldap tree.
	BaseDN string `json:"base_dn"`
	// BindDN is the distinguished name of the
	// user used to search the ldap tree.
	BindDN string `json:"bind_dn"`
	// BindPassword is the password for the
	// user specified in BaseDN.
	BindPassword string `json:"-"`
	// UserFilter is the filter that is used
	// to search for ldap users.
	UserFilter string `json:"user_filer"`
	// GroupFilter is the filter that is
	// used to search for ldap groups.
	GroupFilter string `json:"group_filter"`
	// Host is the server that hosts the
	// ldap tree.
	Host string `json:"host"`
	// Port is the port that the ldap
	// tree is served on.
	Port int `json:"port"`
	// ServerName is the name of the server
	// where the ldap tree is served.
	ServerName string `json:"server_name"`
	// UseSSL is whether or not to use
	// ssl for the ldap connection.
	UseSSL bool `json:"use_ssl"`
	// SkipTLS maintains whether or not
	// to skip tls phase in ldap.
	SkipTLS bool `json:"skip_tls"`
	// InsecureSkipVerify maintains whether
	// or not any certificate should be
	// trusted by the ldap client.
	InsecureSkipVerify bool `json:"insecure"`
	// AutoCreateShadowUsers maintains the
	// level of creating shadow users from
	// this LDAP: 0 = disabled; 1 = created disabled; and 2 = created enabled.
	AutoCreateShadowUsers AutoCreateShadowUsersLevel `json:"auto_create_shadow_users"`
}

// CreateClient creates a new ldap.LDAPClient based on the information in ldp.
func (ldp *LDAP) CreateClient() *ldap.LDAPClient {
	return &ldap.LDAPClient{
		Attributes:         strings.Split(ldp.Attributes, ","),
		Base:               ldp.BaseDN,
		BindDN:             ldp.BindDN,
		BindPassword:       ldp.BindPassword,
		GroupFilter:        ldp.GroupFilter,
		Host:               ldp.Host,
		UserFilter:         ldp.UserFilter,
		Port:               ldp.Port,
		ServerName:         ldp.ServerName,
		SkipTLS:            ldp.SkipTLS,
		UseSSL:             ldp.UseSSL,
		InsecureSkipVerify: ldp.InsecureSkipVerify,
	}
}

// Authenticate authenticates a user name and password against LDAP.
func (ldp *LDAP) AuthenticateUser(username, password string) (bool, map[string]string, error) {
	return ldp.CreateClient().Authenticate(username, password)
}

// Configured returns true if LDAP contains all required information
// in order to do its job and authenticate users.
func (ldp *LDAP) Configured() bool {
	return ldp.Host != "" &&
		ldp.Port > 0 &&
		ldp.BindDN != "" &&
		ldp.BaseDN != "" &&
		ldp.BindPassword != ""
}

// Verify verifies ldap configuration
func (ldp *LDAP) Verify() error {
	client := ldp.CreateClient()
	if err := client.Connect(); err != nil {
		return fmt.Errorf("LDAP: cannot connect to %s:%d", ldp.Host, ldp.Port)
	}
	if err := client.Conn.Bind(ldp.BindDN, ldp.BindPassword); err != nil {
		return fmt.Errorf("LDAP: cannot bind to server: invalid credentials or bad DN")
	}
	return nil
}

// NewLDAP creates a new LDAP object.
func NewLDAP(name string) *LDAP { return &LDAP{Meta: Meta{Kind: "LDAP", Name: name}} }
