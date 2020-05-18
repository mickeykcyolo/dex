package model

import (
	"database/sql/driver"
	"errors"
	"fmt"
	"github.com/dgrijalva/jwt-go"
	"strings"
	"time"
)

// licenseKey is the public key used to validate system licenses.
var licenseKey = []byte(`
-----BEGIN PUBLIC KEY-----
MIICIjANBgkqhkiG9w0BAQEFAAOCAg8AMIICCgKCAgEAvopphJxheo+N2Dx67hdl
ypr4eivgoDkPbFUkByso8+cxMd8TQd4+XdTHMhdCpTlaq2at1S8iXaMXNSU1P0rm
7/ZObFXGKNG5VhzJph2m35wWsvlVaLRlN5MCmpBfCdM22i+Gft/w8nB+IV+vOa4l
hkX38cOoR9Ahm9/ZQGjmuly7EEFsxCEXhaTcUfUH/XnIHs5MH5Ta3n5fNXJQfEP2
7ylpGZnKTLdUNf3OR3VSHGyQn11IKmB2yAVdo6qjklXk4b3oIzg0qTKGr/gUsTau
c/VH/ZLZgTIpbtqMHCvShLuM5TCIJqqz778HIgVndDvMtOG05cXUEGqKd6PHR8o1
l/ozKtPlA794E34XD+5/vzkSdAJqZjCQBQ3tmzqbpkrmKAt5q2Zp87p7+RDE70yR
y1ZYYoQByPbtFddaJ2TMifa3R+frPlGOxfTnKs/aCDIOQDCQof4rpGglzD9p2dsJ
xTxiRtc3n1T1Wn2hiM9VKvsmBFCFbuovwSxb7IYGXQrzbM1iEKjFD5DDObXDzayG
KI1a0fGhcDQDq7HI44WC+LauK+rg7JhbGLkUVhGB48whwYBDpHuYQsZPw1lJgEX9
JgS3dHtWCl2KwL3SsXZEvz6IY6ONFADVrrSu7qJGzyoJ4FHne2K8tIfy39n1az81
G56KpgVf9MrJDOpkkKjie7MCAwEAAQ==
-----END PUBLIC KEY-----
`)

// License represents a license to use the system.
type License struct {
	// Users represents the amount of available user seats.
	Users int `json:"users"`
	// NotBefore is a timestamp before which the license is
	// not yet valid, in unix time.
	NotBefore int64 `json:"validation_start"`
	// NotAfter is a timestamp after which the  license is
	// no longer valid, in unix time.
	NotAfter int64 `json:"validation_end"`
	// Raw is the raw-byte representation of the license.
	Raw []byte `json:"-"`
}

// Valid implements jwt.Claims so that the license can be used
// with the jwt library.
func (lic *License) Valid() (err error) {
	if now := time.Now().UTC().Unix(); now > lic.NotAfter {
		err = errors.New("license expired")
	} else if now < lic.NotBefore {
		err = errors.New("license not yet valid")
	}
	return
}

// Unmarshal attempts to parse and validate raw bytes into license.
func (lic *License) Unmarshal(data string) (err error) {
	// prepare the key selection function:
	key := func(_ *jwt.Token) (interface{}, error) {
		return jwt.ParseRSAPublicKeyFromPEM(licenseKey)
	}

	// try to parse lic JWT:
	if _, err = jwt.ParseWithClaims(strings.TrimSpace(data), lic, key); err != nil {
		return
	}

	// maintain a copy of the raw lic:
	lic.Raw = append(lic.Raw[:0], data...)
	return
}

// Scan implements db.Scanner so that a license can be read idiomatically from a database.
func (lic *License) Scan(src interface{}) error {
	switch src := src.(type) {
	case []byte:
		return lic.Unmarshal(string(src))
	case string:
		return lic.Unmarshal(src)
	default:
		return fmt.Errorf("unsupporetd source of type %T", src)
	}
}

// Value implements driver.Valuer so that license can be written idiomatically into a database.
func (lic *License) Value() (driver.Value, error) {
	if len(lic.Raw) > 0 {
		return lic.Raw, nil
	} else {
		return nil, errors.New("invalid license")
	}
}

// ParseLicense attempts to parse a license from data.
func ParseLicense(data string) (license *License, err error) {
	// assign a new license:
	license = new(License)

	// try to parse and validate license:
	if err = license.Unmarshal(data); err != nil {
		// something went wrong, nullify license:
		license = nil
	}
	return
}
