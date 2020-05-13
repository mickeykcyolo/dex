package repo

import (
	"database/sql"
	"github.com/cyolo-core/cmd/dex/model"
)

// LDAPRepository is a repo for ldap objects.
type LDAPRepository interface {
	// Retrieve finds all ldaps that match an optional where clause.
	Retrieve(where string, args ...interface{}) ([]*model.LDAP, error)
	// Retrieve finds the first ldap that matches an optional where clause.
	RetrieveOne(where string, args ...interface{}) (*model.LDAP, error)
	// Create creates an ldap with information from from map m.
	Create(m M) (*model.LDAP, error)
	// CreateIfNotExists creates an ldap with information from map m
	// if it does not already exist.
	CreateIfNotExists(m M) (*model.LDAP, error)
	// Upsert updates an ldap with name from m if exists, creates it otherwise.
	Upsert(m M) (ldap *model.LDAP, err error)
	// Update updates objects that match where with values from m.
	Update(m M, where string, args ...interface{}) error
	// Delete deletes all ldaps that match an optional where clause.
	Delete(where string, args ...interface{}) error
	// CreateTable creates the ldaps table if it does not exist.
	CreateTable() error
}

// sqlLDAPRepository is a sql implementation for LDAPRepository.
type sqlLDAPRepository struct{ base }

// Retrieve finds all ldaps that match an optional where clause.
func (repo sqlLDAPRepository) Retrieve(where string, args ...interface{}) (ldaps []*model.LDAP, err error) {
	var (
		rows *sql.Rows
		ldap *model.LDAP
	)
	ldaps = make([]*model.LDAP, 0)
	if rows, err = repo.Query(buildStmt("select * from ldaps", where), args...); err == nil {
		defer rows.Close()
		for rows.Next() && err == nil {
			if ldap, err = repo.ScanLDAP(rows); err == nil {
				ldaps = append(ldaps, ldap)
			}
		}
	}
	return
}

// Retrieve finds the first ldap that matches an optional where clause.
func (repo sqlLDAPRepository) RetrieveOne(where string, args ...interface{}) (ldap *model.LDAP, err error) {
	if row := repo.QueryRow(buildStmt("select * from ldaps", where), args...); row != nil {
		if ldap, err = repo.ScanLDAP(row); err != nil {
			ldap = nil
		}
	}
	return
}

// Create creates an ldap with information from from map m.
func (repo sqlLDAPRepository) Create(m M) (ldap *model.LDAP, err error) {
	insert, args := buildInsertStmt("ldaps", m)
	if err = repo.Exec(insert, args...); err == nil {
		ldap, err = repo.RetrieveOne("name = ?", m["name"])
	}
	return
}

// CreateIfNotExists creates a ldap with information from map m,
// if it does not exist.
func (repo sqlLDAPRepository) CreateIfNotExists(m M) (ldap *model.LDAP, err error) {
	if ldap, err = repo.RetrieveOne("name = ?", m["name"]); err == nil {
		return
	}
	return repo.Create(m)
}

// Upsert updates a ldap with name from m if exists, creates it otherwise.
func (repo sqlLDAPRepository) Upsert(m M) (ldap *model.LDAP, err error) {
	if ldap, err = repo.RetrieveOne("name = ?", m["name"]); err == nil && ldap != nil {
		if err = repo.Update(m, "name = ?", m["name"]); err == nil {
			return repo.RetrieveOne("name = ?", m["name"])
		}
	}
	return repo.Create(m)
}

// Update updates objects that match where with values from m.
func (repo sqlLDAPRepository) Update(m M, where string, args ...interface{}) error {
	updateStmt, updateArgs := buildUpdateStmt("ldaps", m)
	args = append(updateArgs, args...)
	return repo.Exec(buildStmt(updateStmt, where), args...)
}

// Delete deletes all ldaps that match an optional where clause.
func (repo sqlLDAPRepository) Delete(where string, args ...interface{}) error {
	return repo.Exec(buildStmt("delete from ldaps", where), args...)
}

// CreateTable creates the ldaps table if it does not exist.
func (repo sqlLDAPRepository) CreateTable() error {
	return repo.Exec("" +
		"create table if not exists ldaps (" +
		"id integer primary key, " +
		"name text unique not null," +
		"system integer default 0," +
		"enabled integer default 0," +
		"ctime timestamp default current_timestamp," +
		"mtime timestamp default current_timestamp," +
		"attributes text default 'sAMAccountName,'," +
		"base_dn text default ''," +
		"bind_dn text default ''," +
		"bind_password text default ''," +
		"user_filter text default '(sAMAccountName=%s)'," +
		"group_filter text default '(memberUid=%s)'," +
		"host text default ''," +
		"port integer default 389," +
		"server_name text default ''," +
		"use_ssl integer default 0," +
		"skip_tls integer default 1," +
		"insecure integer default 1," +
		"auto_create_shadow_users integer default 1," +
		// type assertions:
		"check(system in (0,1,'true','false'))," +
		"check(enabled in (0,1,'true','false'))," +
		"check(use_ssl in (0,1,'true','false'))," +
		"check(skip_tls in (0,1,'true','false'))," +
		"check(insecure in (0,1,'true','false'))," +
		"check(auto_create_shadow_users in (0,1,2))" +
		");" +
		"" +
		"create trigger if not exists update_ldaps_mtime after update on ldaps " +
		"begin " +
		"	update ldaps set mtime = current_timestamp where id = old.id;" +
		"end;",
	)
}

// ScanLDAP attempts to scan out a ldap from scanner s.
func (repo sqlLDAPRepository) ScanLDAP(s Scanner) (ldap *model.LDAP, err error) {
	ldap = model.NewLDAP("")
	err = s.Scan(
		&ldap.Id,
		&ldap.Name,
		&ldap.System,
		&ldap.Enabled,
		&ldap.Ctime,
		&ldap.Mtime,
		&ldap.Attributes,
		&ldap.BaseDN,
		&ldap.BindDN,
		&ldap.BindPassword,
		&ldap.UserFilter,
		&ldap.GroupFilter,
		&ldap.Host,
		&ldap.Port,
		&ldap.ServerName,
		&ldap.UseSSL,
		&ldap.SkipTLS,
		&ldap.InsecureSkipVerify,
		&ldap.AutoCreateShadowUsers,
	)
	return
}

// NewSQLLDAPRepository returns a new ldap repo with db.
func NewSQLLDAPRepository(db *sql.DB) LDAPRepository {
	return sqlLDAPRepository{base{db}}
}
