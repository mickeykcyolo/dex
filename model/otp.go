package model

import (
	"crypto/rand"
	"fmt"
	uuid "github.com/satori/go.uuid"
	"math/big"
	"time"
)

// OTP represents a one-time code.
type OTP struct {
	// id
	Id string
	// User is the user that the otp was created for.
	User *User
	// Ctime records when was the otp created.
	Ctime time.Time
}

// GenerateOTP generates an OTP for user u.
func GenerateOTP(u *User, strong bool) *OTP {
	max := big.NewInt(999999)
	n, _ := rand.Int(rand.Reader, max)
	id := fmt.Sprintf("%06d", n.Int64())

	if strong {
		id = uuid.Must(uuid.NewV4(), nil).String()
	}

	return &OTP{
		Id:    id,
		User:  u,
		Ctime: time.Now(),
	}
}
