package repo

import (
	"database/sql"
	"github.com/cyolo-core/cmd/dex/model"
	"time"
)

// LogEntryRepository is a repo for entry objects.
type LogEntryRepository interface {
	// Retrieve finds all entries that match an optional where clause.
	Retrieve(where string, args ...interface{}) ([]*model.LogEntry, error)
	// Create creates a entry.
	Create(username, remoteAddr, mapping, text string) error
	// Delete deletes all entries that match an optional where clause.
	Delete(where string, args ...interface{}) error
	// CreateTable creates the entries table if it does not exist.
	CreateTable() error
}

// sqlLogEntryRepository is a sql implementation for LogEntryRepository.
type sqlLogEntryRepository struct{ base }

// Retrieve finds all entries that match an optional where clause.
func (repo sqlLogEntryRepository) Retrieve(where string, args ...interface{}) (entries []*model.LogEntry, err error) {
	var (
		rows  *sql.Rows
		entry *model.LogEntry
	)
	entries = make([]*model.LogEntry, 0)
	if rows, err = repo.Query(buildStmt("select * from log_entries", where), args...); err == nil {
		defer rows.Close()
		for rows.Next() && err == nil {
			if entry, err = repo.ScanLogEntry(rows); err == nil {
				entries = append(entries, entry)
			}
		}
	}
	return
}

// Create creates a entry with information from from map m.
func (repo sqlLogEntryRepository) Create(username, addr, mapping, text string) error {
	return repo.Exec("insert into log_entries (timestamp, username, remote_addr, mapping, text) values (?,?,?,?,?)", time.Now(), username, addr, mapping, text)
}

// Delete deletes all entries that match an optional where clause.
func (repo sqlLogEntryRepository) Delete(where string, args ...interface{}) error {
	return repo.Exec(buildStmt("delete from log_entries", where), args...)
}

// CreateTable creates the entries table if it does not exist.
func (repo sqlLogEntryRepository) CreateTable() error {
	return repo.Exec("" +
		"create table if not exists log_entries (" +
		"id integer primary key, " +
		"timestamp timestamp not null," +
		"username text default ''," +
		"remote_addr text default ''," +
		"mapping text default ''," +
		"text text default ''" +
		");",
	)
}

// ScanLogEntry attempts to scan out a entry from scanner s.
func (repo sqlLogEntryRepository) ScanLogEntry(s Scanner) (entry *model.LogEntry, err error) {
	entry = model.NewLogEntry()
	err = s.Scan(
		&entry.Id,
		&entry.Timestamp,
		&entry.Username,
		&entry.RemoteAddr,
		&entry.Mapping,
		&entry.Text,
	)
	return
}

// NewSQLLogEntryRepository returns a new entry repo with db.
func NewSQLLogEntryRepository(db *sql.DB) LogEntryRepository {
	return sqlLogEntryRepository{base{db}}
}
