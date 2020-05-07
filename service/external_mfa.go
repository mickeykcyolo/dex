package service

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	authgw "github.com/cyolo-core/cmd/external_mfa/model"
	"github.com/cyolo-core/cmd/idac/model"
	"net/http"
	"net/url"
	"strings"
	"sync"
	"time"
)

// ExternalMFAService manages interactions with an
// external multi-factor authentication systems.
type ExternalMFAService interface {
	// InitiateSMS starts an external otp
	// process for user u via SMS.
	InitiateSMS(u *model.User, phoneNumber string, redirect *url.URL) (err error)
	// Authenticate authenticates a code.
	Authenticate(code string) (*model.User, bool)
	// StartGC starts garbage collection
	// for stale OTP objects.
	StartGC(ctx context.Context)
}

// make sure that cyoloExternalMFAService implements ExternalMFAService
var _ ExternalMFAService = cyoloExternalMFAService{}

// cyoloExternalMFAService is uses the external_mfa package.
type cyoloExternalMFAService struct {
	// secret is used to authenticate
	// with the external service.
	secret string
	// serviceURL is the url of the external
	// authentication gateway.
	serviceURL string
	// client is the http client used to
	// talk to the authentication gateway.
	client *http.Client
	// lifetime controls the max lifetime
	// of otp objects.
	lifetime time.Duration
	// mu protects the otps map below.
	mu sync.RWMutex
	// otps is a mapping between
	// codes an otp objects.
	otps map[string]*model.OTP
}

// InitiateSMS starts an external otp process with sms for user u.
func (svc cyoloExternalMFAService) InitiateSMS(u *model.User, phoneNumber string, redirect *url.URL) error {
	if len(phoneNumber) == 0 {
		return errors.New("empty phone number")
	}
	if strings.HasPrefix("+", phoneNumber) {
		return errors.New("phone number must start with a plus sign")
	}

	otp := model.GenerateOTP(u, redirect != nil)
	uri := otp.Id

	if redirect != nil {
		redirect.RawQuery += fmt.Sprintf("code=%s", otp.Id)
		uri = redirect.String()
	}

	// if the user has a supervisor, send the user's name
	// to the supervisor.
	name := ""
	if u.Supervisor != nil {
		name = u.Name
	}

	buf, _ := json.Marshal(authgw.SMS{phoneNumber, uri, name})

	req, err := http.NewRequest("POST", svc.serviceURL+"/v1/sms", bytes.NewReader(buf))

	if err != nil {
		return err
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", svc.secret))

	res, err := svc.client.Do(req)

	if err != nil {
		return err
	}
	if code := res.StatusCode; code != 204 {
		return fmt.Errorf("server returned unexpected code: %d", code)
	}

	svc.mu.Lock()
	svc.otps[otp.Id] = otp
	svc.mu.Unlock()

	return nil
}

// Authenticate authenticates a code against the otp store.
func (svc cyoloExternalMFAService) Authenticate(code string) (*model.User, bool) {
	svc.mu.RLock()
	otp, ok := svc.otps[code]
	svc.mu.RUnlock()

	svc.mu.Lock()
	delete(svc.otps, code)
	svc.mu.Unlock()

	if ok && svc.isValid(otp) {
		return otp.User, true
	}
	return nil, false
}

// StartGC starts garbage collection of otp objects.
func (svc cyoloExternalMFAService) StartGC(ctx context.Context) {
	go func() {
		for {
			select {
			case <-ctx.Done():
				return
			default:
				svc.mu.RLock()
				otps := make([]*model.OTP, len(svc.otps))
				otps = otps[:0]

				for _, otp := range svc.otps {
					if !svc.isValid(otp) {
						otps = append(otps, otp)
					}
				}
				svc.mu.RUnlock()

				if len(otps) > 0 {
					svc.mu.Lock()
					for _, otp := range otps {
						delete(svc.otps, otp.Id)
					}
					svc.mu.Unlock()
				}
			}
		}
	}()
}

// isValid returns true if otp is within the allowed lifetime.
func (svc cyoloExternalMFAService) isValid(otp *model.OTP) bool {
	return now().Sub(otp.Ctime) < svc.lifetime
}

// NewCyoloExternalMFAService returns a new ExternalMFAService that users the external_mfa
func NewCyoloExternalMFAService(url, secret string, lifetime time.Duration, client *http.Client) ExternalMFAService {
	if client == nil {
		client = http.DefaultClient
	}
	return &cyoloExternalMFAService{
		secret:     secret,
		serviceURL: url,
		client:     client,
		lifetime:   lifetime,
		otps:       map[string]*model.OTP{},
	}
}
