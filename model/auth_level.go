package model

// AuthLevel represents the authentication level of the user.
type AuthLevel int

const (
	// Anonymous is the starting authentication state
	// for any user.
	Anonymous AuthLevel = iota
	// PreAuthenticated represents a state in which
	// a user was successfully authenticated
	// using a single factor.
	PreAuthenticated
	// PreAuthenticated represents a state in which
	// a user was successfully authenticated
	// using multiple factors.
	Authenticated
)

// SupportedAuthLevels maintains supported user authentication levels.
var SupportedAuthLevels = []struct {
	Id   Identity `json:"id"`
	Name string   `json:"name"`
}{
	{Identity(Anonymous), "Anonymous"},
	{Identity(PreAuthenticated), "PreAuthenticated"},
	{Identity(Authenticated), "Authenticated"},
}
