package model

import (
	"crypto/tls"
)

// Domain represents one of the organization's domains.
type Domain struct {
	Meta        `json:",inline"`
	Certificate string `json:"-"`
	PrivateKey  string `json:"-"`
}

// X509KeyPair attempts to returns the tls key-pair the corresponds to domain d.
func (d Domain) X509KeyPair() (tls.Certificate, error) {
	return tls.X509KeyPair([]byte(d.Certificate), []byte(d.PrivateKey))
}

// NewDomain creates a new domain object.
func NewDomain(name string) *Domain {
	return &Domain{Meta: Meta{Kind: "Domain", Name: name}}
}
