package controller

import (
	"github.com/cyolo-core/cmd/dex/model"
	"github.com/cyolo-core/cmd/dex/service"
	"net/http"
	"time"
)

var (
	// SessionCookieName is the name of the session cookie generated by the login controller.
	SessionCookieName = "cyolo-sid"
	// zeroTime is the zero unix time for expiring cookies.
	zeroTime = time.Unix(0, 0)
)

// baseController implements the basic functionality of an http controller.
type baseController struct {
	service.SessionService
}

// UpsertSession returns a the user's session if it exists, creates
// an anonymous session otherwise.
func (ctrl baseController) UpsertSession(w http.ResponseWriter, r *http.Request) *model.Session {
	if session, ok := ctrl.RetrieveSession(r); ok {
		return session
	}
	return ctrl.GenerateSession(model.NewUser("anonymous"), w, r)
}

// RetrieveSession attempts to retrieve a session for request r.
func (ctrl baseController) RetrieveSession(r *http.Request) (session *model.Session, ok bool) {
	if cookie, err := r.Cookie(SessionCookieName); err == nil && cookie != nil {
		session, ok = ctrl.SessionService.RetrieveSession(cookie.Value)
	}
	return
}

// Me returns information about the currently logged-on user, if any.
//
// Method = {GET}
// Path = /v1/users/me
// Produces = application/json
func (ctrl baseGuacTunnelController) Me(w http.ResponseWriter, r *http.Request) {
	if session := ctrl.UpsertSession(w, r); session.User.Name == "anonymous" {
		encode(403, session.User, w, r)
	} else {
		encode(200, session.User, w, r)
	}
}

// DestroySession destroys the session (if exists) for request r.
func (ctrl baseController) DestroySession(w http.ResponseWriter, r *http.Request) {
	if cookie, err := r.Cookie(SessionCookieName); err == nil && cookie != nil {
		ctrl.SessionService.DestroySession(cookie.Value)
	}

	// reset the session-cookie:
	http.SetCookie(w, &http.Cookie{
		Name:     SessionCookieName,
		Value:    "",
		Path:     "/",
		Domain:   cookieDomain(r.Host),
		Secure:   r.TLS != nil,
		HttpOnly: true,
		SameSite: http.SameSiteLaxMode,
		Expires:  zeroTime,
	})
}

// GenerateSession generates a session for user u.
func (ctrl baseController) GenerateSession(u *model.User, w http.ResponseWriter, r *http.Request) *model.Session {
	s := ctrl.SessionService.GenerateSession(u)
	http.SetCookie(w, &http.Cookie{
		Name:     SessionCookieName,
		Value:    s.Id,
		Path:     "/",
		Domain:   cookieDomain(r.Host),
		Secure:   r.TLS != nil,
		HttpOnly: true,
		SameSite: http.SameSiteLaxMode,
	})
	return s
}
