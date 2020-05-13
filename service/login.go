package service

import (
	"errors"
	"github.com/cyolo-core/cmd/dex/model"
	"github.com/cyolo-core/cmd/dex/repo"
	"regexp"
)

var (
	// ErrIllegalUsername is returned whenever a username does
	// not pass usernameRegex validations.
	ErrIllegalUsername = errors.New("illegal characters in user name")
	// ErrUsernameOrPasswordIncorrect is returned whenever a username
	// or password is incorrect.
	ErrUsernameOrPasswordIncorrect = errors.New("username of password incorrect")
)

// usernameRegex is used to validate a username and protect against
// LDAP injection attacks.
var usernameRegex, _ = regexp.Compile(`^[\w\s.$\-]+$`)

// ErrUserDisabled occurs when a user successfully logs in but is not currently enabled.
var ErrUserDisabled = errors.New("user disabled")

type LoginService interface {
	// DoLocal attempts to authenticate a user against a local password hash.
	DoLocal(username string, password string) (*model.User, error)
	// DoLDAP attempts to authenticate a user against an LDAP server.
	DoLDAP(username string, password string) (*model.User, error)
	// DoLogin attempts to authenticate a user against both local and LDAP user stores.
	DoLogin(username string, password string) (*model.User, error)
}

// make sure that the implementation follows the interface.
var _ LoginService = loginServiceImpl{}

// loginServiceImpl is the implementation of LoginService.
type loginServiceImpl struct {
	users repo.UserRepository
	ldaps repo.LDAPRepository
}

// DoLocal attempts to authenticate a user against a local password hash.
func (svc loginServiceImpl) DoLocal(username, password string) (user *model.User, err error) {
	if user, err = svc.users.RetrieveOne("name = ? and length(password_hash) > 0", username); err == nil {
		if ok := user.AuthenticatePassword(password); !ok {
			user = nil
			err = ErrUsernameOrPasswordIncorrect
			return
		}
		if !user.Enabled {
			user = nil
			err = ErrUserDisabled
		}
	}
	return
}

// DoLDAP attempts to authenticate a user against an LDAP server.
func (svc loginServiceImpl) DoLDAP(username, password string) (user *model.User, err error) {
	var ldaps []*model.LDAP
	authenticated := false

	if !usernameRegex.MatchString(username) {
		err = ErrIllegalUsername
		return
	}

	if ldaps, err = svc.ldaps.Retrieve(""); err != nil {
		return
	}

	if len(ldaps) == 0 {
		err = ErrUsernameOrPasswordIncorrect
		return
	}

	if user, err = svc.users.RetrieveOne("name = ?", username); err != nil {
		err = nil
		user = model.NewUser(username)
	}

	for _, ldap := range ldaps {
		// protect against authenticating with ldap
		// objects that have not been configured yet:
		if !ldap.Configured() {
			continue
		}
		if ok, attribs, err := ldap.AuthenticateUser(username, password); err == nil && ok {
			for key, value := range attribs {
				user.AdditionalInfo[key] = value
				user.Origin = ldap
				authenticated = true
			}
			break
		}
	}

	if !authenticated {
		err = ErrUsernameOrPasswordIncorrect
		user = nil
		return
	}

	if !user.Enrolled {
		return
	}

	if !user.Enabled {
		err = ErrUserDisabled
		user = nil
		return
	}

	return
}

// DoLogin attempts to authenticate a user against both local and LDAP user stores.
func (svc loginServiceImpl) DoLogin(username, password string) (user *model.User, err error) {
	if user, err = svc.DoLocal(username, password); err == nil {
		return
	}
	return svc.DoLDAP(username, password)
}

// NewLoginService returns a new login service.
func NewLoginService(users repo.UserRepository, ldaps repo.LDAPRepository) LoginService {
	return &loginServiceImpl{
		users: users,
		ldaps: ldaps,
	}
}
