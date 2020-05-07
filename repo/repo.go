package repo

import (
	"fmt"
	"strings"
)

var (
// make sure implementations are complete
//_ DomainRepository        = sqlDomainRepository{}
//_ LDAPRepository          = sqlLDAPRepository{}
//_ MappingRepository       = sqlMappingRepository{}
//_ PolicyRepository        = sqlPolicyRepository{}
//_ UserRepository          = sqlUserRepository{}
//_ EncryptionKeyRepository = sqlEncryptionKeyRepository{}
//_ LogEntryRepository      = sqlLogEntryRepository{}
)

// Scanner is an interface that represents both sql.Row and sql.Rows.
type Scanner interface {
	Scan(dst ...interface{}) error
}

// buildStmt takes a statement and a where clause and concatenates them appropriately.
func buildStmt(query string, where string) string {
	if len(where) > 0 {
		query += fmt.Sprintf(" where %s", where)
	}
	return query
}

// buildInsertStmt builds an insert statement for kind with map m.
func buildInsertStmt(kind string, m M) (insert string, args []interface{}) {
	var (
		keys   []string
		values []string
	)

	for key, value := range m {
		keys = append(keys, key)
		args = append(args, value)
		values = append(values, "?")
	}

	insert = fmt.Sprintf("insert into %s (%s) values (%s)",
		kind,
		strings.Join(keys, ","),
		strings.Join(values, ","))
	return
}

// buildUpdateStmt builds an update statement for kind with map m.
func buildUpdateStmt(kind string, m M) (update string, args []interface{}) {
	var keys []string

	for key, value := range m {
		keys = append(keys, fmt.Sprintf("%s=?", key))
		args = append(args, value)
	}

	update = fmt.Sprintf("update %s set %s", kind, strings.Join(keys, ","))
	return
}
