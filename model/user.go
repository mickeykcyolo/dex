package model

import (
	"crypto/rand"
	"encoding/base32"
	"github.com/dgryski/dgoogauth"
	"golang.org/x/crypto/bcrypt"
)

type User struct {
	// Meta ...
	Meta `json:",inline"`
	// PhoneNumber is the users phone
	// number for SMS verification.
	PhoneNumber string `json:"phone_number"`
	// TOTPSecret is the secret used
	// authenticate the user's time-based-otp.
	TOTPSecret string `json:"-"`
	// TOTPEnabled maintains whether
	// or not the user is allowed
	// time-based-otp for MFA.
	TOTPEnabled bool `json:"totp_enabled"`
	// TOTPEnrolled maintains whether
	// or not the user had enrolled
	// an authenticator app with their
	// account.
	TOTPEnrolled bool `json:"totp_enrolled"`
	// SSOPrivateKey is the user's private
	// key.
	PrivateKey string `json:"-"`
	// PasswordHash is the user's password
	// hash in bcrypt.
	PasswordHash string `json:"-"`
	// AdditionalInfo maintains additional
	// information about the user, for example
	// ldap.Attributes.
	AdditionalInfo map[string]interface{} `json:"-"`
	// AuthLevel represents the authentication state
	// of the user.
	AuthLevel `json:"auth_level"`
	// Supervisor is an optional
	// entity that can approve
	// actions for a user.
	Supervisor *User `json:"supervisor"`
	// PersonalDesktop is the ip address
	// or hostname of the user's
	// personal workstation.
	PersonalDesktop string `json:"personal_desktop"`
	// Enrolled maintains whether or not
	// the user has completed enrollment.
	Enrolled bool `json:"enrolled"`
	// Origin maintains the entity that
	// authenticated the user.
	Origin interface{} `json:"-"`
}

// StorePassword stores password as the new password hash.
func (user *User) StorePassword(password string) {
	hash, _ := bcrypt.GenerateFromPassword([]byte(password), bcrypt.MinCost)
	user.PasswordHash = string(hash)
}

// AuthenticatePassword authenticates password against the user's password hash.
func (user *User) AuthenticatePassword(password string) bool {
	return bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(password)) == nil
}

// TOTPConfig gets the totp configuration for the user.
func (user *User) TOTPConfig() *dgoogauth.OTPConfig {
	return &dgoogauth.OTPConfig{
		Secret:     user.TOTPSecret,
		WindowSize: 3,
		UTC:        true,
	}
}

// AuthenticateTOTP authenticates a user's one time password.
func (user *User) AuthenticateTOTP(otp string) (ok bool) {
	ok, _ = user.UpsertTOTP().Authenticate(otp)
	return
}

// ProvisionTotpKeyURI returns the user's totp key uri.
func (user *User) ProvisionTotpKeyURI() string {
	return user.UpsertTOTP().ProvisionURIWithIssuer(user.Name, "cyolo")
}

// UpsertTOTP generates a totp configuration for the user.
func (user *User) UpsertTOTP() *dgoogauth.OTPConfig {
	if user.TOTPSecret != "" {
		return user.TOTPConfig()
	}

	// generate totp configuration for user:
	random := make([]byte, 10)
	n, _ := rand.Read(random)
	user.TOTPSecret = base32.StdEncoding.EncodeToString(random[:n])
	user.TOTPEnabled = true
	return user.TOTPConfig()
}

// HasMFA returns true if the user can perform multi-factor authentication
func (user *User) HasMFA() bool {
	return user.TOTPEnrolled || len(user.PhoneNumber) > 0
}

// NewUser creates a new user object.
func NewUser(name string) *User {
	return &User{
		Meta:           Meta{Kind: "User", Name: name},
		AdditionalInfo: map[string]interface{}{},
	}
}
