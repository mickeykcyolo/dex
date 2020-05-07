package controller

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/cyolo-core/cmd/dex/repo"
	"golang.org/x/net/publicsuffix"
	"net"
	"net/http"
	"regexp"
	"strconv"
	"strings"
)

var (
	// KeyRegex is the regexp that is used to sanitize parameter keys.
	KeyRegex, _ = regexp.Compile("^\\w+$")
)

const (
	// hdrContentType represents the content-type header.
	hdrContentType = "content-type"
)

const (
	// applicationJSON represents the application/json type.
	applicationJSON = "application/json"
	// applicationURLEncoded represents the application/x-www-form-urlencoded type.
	applicationURLEncoded = "application/x-www-form-urlencoded"
)

// decodeParams attempts to decode request body based on content-type
// into []repo.M.
//
// note that decodeBulk will automatically notify the client of failure.
func decodeBulk(w http.ResponseWriter, r *http.Request) (bulk []repo.M, err error) {

	// assign a new empty slice into bulk:
	bulk = make([]repo.M, 0)

	// currently, decodeBulk only supports json:
	if !strings.HasPrefix(strings.ToLower(r.Header.Get(hdrContentType)), applicationJSON) {
		encode(415, errors.New("unsupported content-type"), w, r)
		return
	}

	// attempt to parse bulk:
	if err = json.NewDecoder(r.Body).Decode(&bulk); err != nil {
		encode(400, err, w, r)
		return
	}

	// sanitize parameter keys:
	for i := range bulk {
		bulk[i] = sanitize(bulk[i])
	}

	return
}

// decodeParams attempts to decode request body based on content-type
// into repo.M.
//
// note that decodeParams will automatically notify the client of failure.
func decodeParams(w http.ResponseWriter, r *http.Request) (params repo.M, err error) {

	// assign a new empty map to params:
	params = repo.M{}

	// decode params based on the content-type header:
	ct := strings.ToLower(r.Header.Get(hdrContentType))
	switch {
	case strings.HasPrefix(ct, applicationJSON):
		err = json.NewDecoder(r.Body).Decode(&params)
	case strings.HasPrefix(ct, applicationURLEncoded):
		err = r.ParseForm()
		params = decodeForm(r)
	default:
		http.Error(w, fmt.Sprintf("unsupporetd content type: %q", ct), 422)
		return
	}

	if err != nil {
		encode(400, err, w, r)
		return
	}

	// sanitize:
	params = sanitize(params)
	return
}

// decodeForm parses and sanitizes request form into the form of repo.M.
func decodeForm(r *http.Request) (params repo.M) {

	// assign an empty map in params:
	params = repo.M{}

	// decode url parameters if not already decoded:
	if err := r.ParseForm(); err != nil {
		return
	}

	// decode request form:
	for key, values := range r.Form {
		for _, value := range values {
			switch value {
			case "true":
				params[key] = true
			case "false":
				params[key] = false
			default:
				if ivalue, err := strconv.Atoi(value); err == nil {
					params[key] = ivalue
				} else {
					params[key] = value
				}
			}
		}
	}

	// sanitize:
	params = sanitize(params)
	return
}

// decodeIntArray attempts decode an array of integers from request body.
//
// note that decodeIntArray will automatically notify the client of failure.
func decodeIntArray(w http.ResponseWriter, r *http.Request) (ids []int, err error) {
	ids = make([]int, 0)

	switch ct := r.Header.Get(hdrContentType); strings.ToLower(ct) {
	case applicationJSON:
		err = json.NewDecoder(r.Body).Decode(&ids)
	default:
		http.Error(w, fmt.Sprintf("unsupporetd content type: %q", ct), http.StatusUnprocessableEntity)
		return
	}

	if err != nil {
		encode(400, err, w, r)
	}

	return
}

// encode json-encodes a value and sends it back to the client in wire format.
func encode(code int, value interface{}, w http.ResponseWriter, r *http.Request) {
	w.Header().Set("content-type", "application/json")
	w.WriteHeader(code)

	if value != nil {
		if err, ok := value.(error); ok {
			// value is an error, transform it into
			// the form {"code": N, "error", "..."}
			value = repo.M{"code": code, "error": err.Error()}
		}
		json.NewEncoder(w).Encode(value)
	}
}

// sanitize sanitizes a parameter array keys to only keys that match KeyRegex.
func sanitize(in repo.M) (out repo.M) {
	out = repo.M{}
	for key, value := range in {
		if KeyRegex.MatchString(key) {
			out[key] = value
		}
	}
	return
}

// cookieDomain strips away the first component of a given sub-domain
// and returns a parent domain suitable for setting cookies.
// e.g: a domain "foo.bar.amazon.com" will return ".bar.amazon.com"
// and "amazon.co.uk" will return "amazon.co.uk".
func cookieDomain(domain string) string {
	if strings.Contains(domain, ":") {
		// domain contains a port, strip it:
		domain, _, _ = net.SplitHostPort(domain)
	}
	if tldPlusOne, _ := publicsuffix.EffectiveTLDPlusOne(domain); tldPlusOne == domain {
		// domain is not a sub-domain
		return domain
	}
	if dot := strings.IndexByte(domain, '.'); dot > 0 {
		// domain is a sub-domain:
		return domain[dot:]
	}
	return "." + domain
}

// hostname returns the actual hostname from a request host.
func hostname(r *http.Request) string {
	if host, _, err := net.SplitHostPort(r.Host); err == nil && host != "" {
		return host
	}
	return r.Host
}
