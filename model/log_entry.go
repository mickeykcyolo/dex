package model

import "time"

// LogEntry represents an entry in the system log.
type LogEntry struct {
	// id is the unique identity
	// of the log entry.
	Id Identity `json:"id"`
	// Kind represents the kind
	// of the log entry object.
	Kind string `json:"kind"`
	// Timestamp is the exact date and
	// time when the entry was recorded.
	Timestamp time.Time `json:"timestamp"`
	// SSOUsername is the name of a user associated
	// with the log entry.
	Username string `json:"username"`
	// RemoteAddr is a remote address associated
	// with the log entry.
	RemoteAddr string `json:"remote_addr"`
	// Mapping is a name of the mapping
	// associated with the log entry.
	Mapping string `json:"mapping"`
	// Text represents the log entry data.
	Text string `json:"text"`
}

// NewLogEntry creates a new LogEntry object.
func NewLogEntry() *LogEntry { return &LogEntry{Kind: "LogEntry"} }
