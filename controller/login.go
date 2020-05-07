package controller

import (
	"errors"
	"fmt"
	"github.com/cyolo-core/cmd/dex/model"
	"github.com/cyolo-core/cmd/dex/repo"
	"github.com/cyolo-core/cmd/dex/service"
	"github.com/gorilla/mux"
	"net"
	"net/http"
	"net/url"
	"strings"
	"time"
)

var (
	// totpIssuer is the issuer value for totp registrations.
	totpIssuer = "cyolo"
)

// loginController is responsible for user handling user authentication.
type loginController struct {
	baseController
	service.LogService
	service.LoginService
	service.CrudService
	service.ExternalMFAService
}

// Stage1 authenticates a user with credentials against a user-store.
//
// Methods = {POST}
// Path = /v1/auth/stage/1
// Consumes = {application/x-www-form-urlencoded}
func (ctrl loginController) Stage1(w http.ResponseWriter, r *http.Request) {
	if err := r.ParseForm(); err != nil {
		http.Error(w, err.Error(), 400)
		return
	}

	session, ok := ctrl.RetrieveSession(r)
	username := r.Form.Get("username")
	username = strings.ToLower(username)
	password := r.Form.Get("password")

	if len(username) == 0 || len(password) == 0 {
		encode(400, errors.New("missing username or password"), w, r)
		return
	}

	user, err := ctrl.DoLogin(username, password)

	if err != nil {
		ctrl.Logf(username, r.RemoteAddr, "login", "log in failed with password")
		encode(403, err, w, r)
		return
	}

	ctrl.Logf(username, r.RemoteAddr, "login", "log in success with password")

	if ok {
		// update an existing session:
		session.User = user
	} else {
		session = ctrl.GenerateSession(user, w, r)
	}

	// store username and password for single-sign on:
	session.SetStringValue("username", username)
	session.SetStringValue("password", password)

	// add redirect uri to new session:
	redirectURI := ctrl.buildRedirectURI(r)
	session.SetStringValue("redirect_uri", redirectURI)

	// set user pre-authenticated:
	user.AuthLevel = model.PreAuthenticated

	// tell the client to redirect the user:
	encode(200, map[string]interface{}{"user": user, "redirect_uri": redirectURI}, w, r)
}

// Stage2 authenticates a user with a second factor such as totp.
//
// Methods = {GET/POST}
// Path = /v1/auth/stage/2
// Consumes = {application/x-www-form-urlencoded}
func (ctrl loginController) Stage2(w http.ResponseWriter, r *http.Request) {

	// parse the form, query and body:
	if err := r.ParseForm(); err != nil {
		encode(400, err, w, r)
		return
	}

	// retrieve the session:
	session, ok := ctrl.RetrieveSession(r)

	// for sms callbacks, there is no user session:
	if code := r.Form.Get("code"); len(code) > 0 && ctrl.ExternalMFAService != nil {
		if user, ok := ctrl.Authenticate(code); ok {
			user.AuthLevel = model.Authenticated

			if user.Supervisor != nil {
				ctrl.Logf(user.Name, r.RemoteAddr, "login", "log in approved by %s", user.Supervisor.Name)
			} else {
				ctrl.Logf(user.Name, r.RemoteAddr, "login", "log in success with sms")
			}

			w.Header().Set("content/type", "text/html")
			w.WriteHeader(200)
			w.Write([]byte("<html>" +
				"<head>" +
				"<script>setTimeout(close, 3500);</script>" +
				"</head>" +
				"<body>" +
				"<h1>Login Was Approved For " + user.Name + " </h1>" +
				"</body>" +
				"</html>"))
		}
		return
	}

	if !ok {
		encode(403, errors.New("forbidden"), w, r)
		return
	}
	if err := r.ParseForm(); err != nil {
		encode(400, err, w, r)
		return
	}
	if totp := r.Form.Get("totp"); len(totp) > 0 {
		if session.User.Supervisor == nil && session.User.AuthenticateTOTP(totp) {
			session.User.AuthLevel = model.Authenticated
			ctrl.Logf(session.User.Name, r.RemoteAddr, "login", "log in success with totp")
		}
	}
	if session.User.AuthLevel == model.Authenticated {
		encode(204, nil, w, r)
		return
	}
	if ctrl.ExternalMFAService != nil {

		phoneNumber := session.User.PhoneNumber

		// if user has a supervisor, send them the sms:
		if supervisor := session.User.Supervisor; supervisor != nil {
			phoneNumber = supervisor.PhoneNumber
		}

		now := time.Now()

		// sms was already sent for this session, abort:
		if passed := now.Sub(session.TimeValue("sms_sent", now)); passed > time.Minute {
			http.Error(w, fmt.Sprintf("try again in: %v", passed), 403)
			return
		}

		// build a uri for the SMS:
		uri := &url.URL{Scheme: "https", Host: r.Host, Path: r.URL.Path}

		// send sms to user/supervisor:
		if err := ctrl.ExternalMFAService.InitiateSMS(session.User, phoneNumber, uri); err != nil {
			ctrl.Logf(session.User.Name, r.RemoteAddr, "login", "log in failed with sms: %v", err)
			encode(500, err, w, r)
			return
		}

		// mark session as pending external authentication:
		session.SetTimeValue("sms_sent", now)
		encode(204, nil, w, r)
		return

	}

	ctrl.DestroySession(w, r)
	http.Error(w, "forbidden", 403)
}

// Me returns information about the currently logged-on user, if any,
// about anonymous otherwise.
//
// Methods = {GET}
// Path = /v1/users/me
// Produces = {application/json}
func (ctrl loginController) Me(w http.ResponseWriter, r *http.Request) {
	if session, ok := ctrl.RetrieveSession(r); ok {
		encode(200, session.User, w, r)
		return
	}
	encode(403, nil, w, r)
}

// EditMe edits information about the currently logged on user.
//
// Methods = {PUT}
// Path = /v1/users/me
// Consumes = {application/json}
// Produces = {application/json}
func (ctrl loginController) EditMe(w http.ResponseWriter, r *http.Request) {
	if session, ok := ctrl.RetrieveSession(r); ok {
		if params, err := decodeParams(w, r); err == nil {
			// currently, we only let editing the personal desktop
			// and phone number parameters if they were not previously set by an admin:
			if personalDesktop, ok := params["personal_desktop"].(string); ok {
				if session.User.PersonalDesktop == "" {
					session.User.PersonalDesktop = personalDesktop
				}
			}
			if phoneNumber, ok := params["phone_number"].(string); ok {
				if session.User.PhoneNumber == "" {
					session.User.PhoneNumber = phoneNumber
				}
			}
		}
		encode(200, session.User, w, r)
		return
	}
	encode(403, model.NewUser("anonymous"), w, r)
}

// TotpKeyURI returns the the currently logged-on user's totp key uri.
//
// Methods = {GET}
// Path = /v1/users/me/totp-key-uri
// Produces = {application/json}
func (ctrl loginController) TotpKeyURI(w http.ResponseWriter, r *http.Request) {
	if session, ok := ctrl.RetrieveSession(r); ok {
		// enrolled users cannot obtain their totp uri once again:
		if session.User.TOTPEnrolled {
			encode(403, errors.New("access denied"), w, r)
			return
		}
		encode(200, map[string]string{"uri": session.User.ProvisionTotpKeyURI()}, w, r)
		return
	}
	encode(401, errors.New("unauthorized"), w, r)
}

// InitiateSMS sends an initial SMS to the user
//
// Method = {POST}
// Path = /v1/users/me/initiate-sms
// Consumes = {application/json}
func (ctrl loginController) InitiateSMS(w http.ResponseWriter, r *http.Request) {
	if session, ok := ctrl.RetrieveSession(r); ok {
		// user cannot initiate sms flow if they are already enrolled with a phone number:
		if session.User.Enrolled && session.User.PhoneNumber != "" {
			encode(403, errors.New("access denied"), w, r)
			return
		}
		// throttle sms messages:
		if lastSMS, ok := session.InterfaceValue("last_sms_request").(time.Time); ok {
			if remaining := time.Now().Sub(lastSMS); remaining < time.Minute {
				encode(403, fmt.Errorf("try again in: %v", remaining), w, r)
				return
			}
		}

		if params, err := decodeParams(w, r); err == nil {
			if phone, ok := params["phone_number"].(string); ok {
				if err := ctrl.ExternalMFAService.InitiateSMS(session.User, phone, nil); err != nil {
					encode(500, err, w, r)
					return
				}
				session.SetTimeValue("last_sms_request", time.Now())
				session.SetStringValue("phone_number", phone)
				encode(204, nil, w, r)
				return
			}
			encode(400, errors.New("missing phone number"), w, r)
			return
		}
	}
	encode(403, errors.New("forbidden"), w, r)
}

// Redirect redirects an authenticated user to the place they came from.
//
// Method = {GET}
// Path = /v1/redirect
func (ctrl loginController) Redirect(w http.ResponseWriter, r *http.Request) {
	http.Redirect(w, r, ctrl.buildRedirectURI(r), 302)
}

// Verify verifies a totp or code received via SMS.
//
// Methods = {POST}
// Path = /v1/users/me/verify
// Consumes = {application/json}
// Produces = {application/json}
func (ctrl loginController) Verify(w http.ResponseWriter, r *http.Request) {
	if session, ok := ctrl.RetrieveSession(r); ok {

		if params, err := decodeParams(w, r); err == nil {

			kind := params["kind"]
			code := params["code"]
			user := session.User

			if kind == nil || code == nil {
				encode(400, errors.New("kind and code required"), w, r)
				return
			}
			if kind == "sms" {
				if user.Enrolled {
					// users cannot re-enroll with sms here:
					encode(403, errors.New("access denied"), w, r)
					return
				}
				if user, ok = ctrl.Authenticate(code.(string)); !ok {
					encode(401, errors.New("wrong code"), w, r)
					return
				}
				user.PhoneNumber = session.StringValue("phone_number")
			}
			if kind == "totp" {
				if ok = user.AuthenticateTOTP(code.(string)); !ok {
					encode(401, errors.New("wrong code"), w, r)
					return
				}
				user.TOTPEnrolled = true
			}

			user.AuthLevel = model.Authenticated
			encode(204, nil, w, r)
			return
		}
	}

	encode(401, errors.New("unauthorized"), w, r)
}

// Commit creates the user in the repository and finishes the enrollment process.
//
// Methods = {POST}
// Path = /v1/users/me/commit
// Produces = {application/json}
func (ctrl loginController) Commit(w http.ResponseWriter, r *http.Request) {
	if session, ok := ctrl.RetrieveSession(r); ok && session.User.AuthLevel == model.Authenticated && !session.User.Enrolled {

		// an enrolled user cannot commit
		if session.User.Enrolled {
			encode(403, errors.New("access denied"), w, r)
			return
		}

		enabled := session.User.Enabled

		if ldap, ok := session.User.Origin.(*model.LDAP); ok && !enabled {
			if ldap.AutoCreateShadowUsers == model.Disabled {
				// administrator has disabled auto shadow user creation:
				encode(403, errors.New("administrator has disabled enrollment"), w, r)
				return
			}
			if ldap.AutoCreateShadowUsers == model.CreateEnabled {
				// administrator has configured the users are auto enabled:
				enabled = true
			}
		}

		userM := repo.M{
			"name":         session.User.Name,
			"phone_number": session.User.PhoneNumber,
			"totp_secret":  session.User.TOTPSecret,
			// totp_enabled and totp_enrolled are both
			// determined from TOTPEnrolled:
			"totp_enabled":     session.User.TOTPEnrolled,
			"totp_enrolled":    session.User.TOTPEnrolled,
			"personal_desktop": session.User.PersonalDesktop,
			"supervisor":       nil,
			"enrolled":         true,
			"enabled":          enabled,
		}

		if session.User.System {
			// system users don't need to be enabled
			delete(userM, "enabled")
		}

		user, err := ctrl.Upsert("user", userM)

		if err != nil {
			encode(500, err, w, r)
			return
		}

		msg := struct {
			User        interface{} `json:"user"`
			RedirectURI string      `json:"redirect_uri"`
		}{
			user,
			ctrl.buildRedirectURI(r),
		}

		ctrl.Logf(session.User.Name, r.RemoteAddr, "enroll", "user successfully enrolled")
		session.User.Enrolled = true
		session.User.Enabled = enabled
		encode(201, msg, w, r)
		return
	}

	encode(401, errors.New("unauthorized"), w, r)
}

// buildRedirectURI builds the redirect URI from login
func (ctrl loginController) buildRedirectURI(r *http.Request) string {

	// default value for redirect uri is the user's console:
	redirectURI := "https://users" + cookieDomain(r.Host)

	// try to obtain a redirect uri from the session:
	if session, ok := ctrl.RetrieveSession(r); ok {
		redirectURI = session.StringValue("redirect_uri", redirectURI)
	}

	// protect against redirecting to login again:
	if strings.HasPrefix(redirectURI, "https://login") {
		redirectURI = "https://users" + cookieDomain(r.Host)
	}

	// tack on port if non-empty:
	if _, port, _ := net.SplitHostPort(r.Host); len(port) > 0 {
		if uri, err := url.Parse(redirectURI); err == nil {
			if uri.Port() == "" {
				uri.Host += fmt.Sprintf(":%s", port)
				redirectURI = uri.String()
			}
		}
	}

	return redirectURI
}

func NewLoginController(login service.LoginService, mfa service.ExternalMFAService, session service.SessionService, log service.LogService, view http.Handler, crud service.CrudService) http.Handler {
	ctrl := loginController{
		baseController:     baseController{session},
		LoginService:       login,
		ExternalMFAService: mfa,
		LogService:         log,
		CrudService:        crud,
	}

	rtr := mux.NewRouter()
	rtr.HandleFunc("/v1/auth/stage/1", ctrl.Stage1).Methods("POST")
	rtr.HandleFunc("/v1/auth/stage/2", ctrl.Stage2).Methods("POST", "GET")
	rtr.HandleFunc("/v1/users/me", ctrl.Me).Methods("GET")
	rtr.HandleFunc("/v1/users/me", ctrl.EditMe).Methods("PUT")
	rtr.HandleFunc("/v1/users/me/totp-key-uri", ctrl.TotpKeyURI).Methods("GET")
	rtr.HandleFunc("/v1/users/me/initiate-sms", ctrl.InitiateSMS).Methods("POST")
	rtr.HandleFunc("/v1/users/me/verify", ctrl.Verify).Methods("POST")
	rtr.HandleFunc("/v1/users/me/commit", ctrl.Commit).Methods("POST")
	rtr.HandleFunc("/v1/redirect", ctrl.Redirect).Methods("GET")
	rtr.PathPrefix("/mfa").Handler(http.StripPrefix("/mfa", compress(view)))
	rtr.PathPrefix("/enroll").Handler(http.StripPrefix("/enroll", compress(view)))
	rtr.PathPrefix("/").Handler(compress(view))

	return rtr
}
