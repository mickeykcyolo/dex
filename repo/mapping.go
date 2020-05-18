package repo

import (
	"database/sql"
	"github.com/cyolo-core/cmd/dex/model"
)

// MappingRepository is a repo for mapping objects.
type MappingRepository interface {
	// Retrieve finds all mappings that match an optional where clause.
	Retrieve(where string, args ...interface{}) ([]*model.Mapping, error)
	// Retrieve finds the first mapping that matches an optional where clause.
	RetrieveOne(where string, args ...interface{}) (*model.Mapping, error)
	// Create creates a mapping with information from map m.
	Create(m M) (*model.Mapping, error)
	// CreateIfNotExists creates a mapping with information from m if
	// it does not already exist.
	CreateIfNotExists(m M) (*model.Mapping, error)
	// Upsert updates a mapping with name from m if exists, creates it otherwise.
	Upsert(m M) (mapping *model.Mapping, err error)
	// Update updates objects that match where with values from m.
	Update(m M, where string, args ...interface{}) error
	// Delete deletes all mappings that match an optional where clause.
	Delete(where string, args ...interface{}) error
	// CreateTable creates the mappings table if it does not exist.
	CreateTable() error
}

// sqlMappingRepository is a sql implementation for MappingRepository.
type sqlMappingRepository struct {
	base
	domains DomainRepository
}

// Retrieve finds all mappings that match an optional where clause.
func (repo sqlMappingRepository) Retrieve(where string, args ...interface{}) (mappings []*model.Mapping, err error) {
	var (
		rows    *sql.Rows
		mapping *model.Mapping
	)
	mappings = make([]*model.Mapping, 0)
	if rows, err = repo.Query(buildStmt("select * from mappings", where), args...); err == nil {
		defer rows.Close()
		for rows.Next() && err == nil {
			if mapping, err = repo.ScanMapping(rows); err == nil {
				mappings = append(mappings, mapping)
			}
		}
	}
	return
}

// Retrieve finds the first mapping that matches an optional where clause.
func (repo sqlMappingRepository) RetrieveOne(where string, args ...interface{}) (mapping *model.Mapping, err error) {
	if row := repo.QueryRow(buildStmt("select * from mappings", where), args...); row != nil {
		if mapping, err = repo.ScanMapping(row); err != nil {
			mapping = nil
		}
	}
	return
}

// Create creates a mapping with information from from map m.
func (repo sqlMappingRepository) Create(m M) (mapping *model.Mapping, err error) {
	insert, args := buildInsertStmt("mappings", m)
	if err = repo.Exec(insert, args...); err == nil {
		mapping, err = repo.RetrieveOne("name = ?", m["name"])
	}
	return
}

// CreateIfNotExists creates a mapping with information from map m,
// if it does not exist.
func (repo sqlMappingRepository) CreateIfNotExists(m M) (mapping *model.Mapping, err error) {
	if mapping, err = repo.RetrieveOne("name = ?", m["name"]); err == nil {
		return
	}
	return repo.Create(m)
}

// Upsert updates a mapping with name from m if exists, creates it otherwise.
func (repo sqlMappingRepository) Upsert(m M) (mapping *model.Mapping, err error) {
	if mapping, err = repo.RetrieveOne("name = ?", m["name"]); err == nil && mapping != nil {
		if err = repo.Update(m, "name = ?", m["name"]); err == nil {
			return repo.RetrieveOne("name = ?", m["name"])
		}
	}
	return repo.Create(m)
}

// Update updates objects that match where with values from m.
func (repo sqlMappingRepository) Update(m M, where string, args ...interface{}) error {
	updateStmt, updateArgs := buildUpdateStmt("mappings", m)
	args = append(updateArgs, args...)
	return repo.Exec(buildStmt(updateStmt, where), args...)
}

// Delete deletes all mappings that match an optional where clause.
func (repo sqlMappingRepository) Delete(where string, args ...interface{}) error {
	return repo.Exec(buildStmt("delete from mappings", where), args...)
}

// CreateTable creates the mappings table if it does not exist.
func (repo sqlMappingRepository) CreateTable() error {
	return repo.Exec("" +
		"create table if not exists mappings (" +
		"id integer primary key, " +
		"name text unique not null," +
		"system integer default 0," +
		"enabled integer default 0," +
		"ctime timestamp default current_timestamp," +
		"mtime timestamp default current_timestamp," +
		"external text not null," +
		"domain integer references domains(id) on delete cascade," +
		"internal text not null," +
		"port integer default 0," +
		"protocol integer default 0," +
		"sso integer default 0," +
		"sso_forms_tpl text default ''," +
		"sso_forms_url text default ''," +
		"sso_private_key text default ''," +
		"sso_use_mapping_credentials integer default 0," +
		"sso_username text default ''," +
		"sso_password text default ''," +
		"insecure_skip_verify default 0," +
		// type assertions:
		"check(system in (0,1,'true','false'))," +
		"check(enabled in (0,1,'true','false'))," +
		"check(insecure_skip_verify in (0,1,'true','false'))" +
		");" +
		"" +
		"create trigger if not exists update_mappings_mtime after update on mappings " +
		"begin " +
		"	update mappings set mtime = current_timestamp where id = old.id;" +
		"end;",
	)
}

// ScanMapping attempts to scan out a mapping from scanner s.
func (repo *sqlMappingRepository) ScanMapping(s Scanner) (mapping *model.Mapping, err error) {
	sso := 0
	domain := 0
	protocol := 0

	mapping = model.NewMapping("")

	err = s.Scan(
		&mapping.Id,
		&mapping.Name,
		&mapping.System,
		&mapping.Enabled,
		&mapping.Ctime,
		&mapping.Mtime,
		&mapping.External,
		&domain,
		&mapping.Internal,
		&mapping.Port,
		&protocol,
		&sso,
		&mapping.SSOFormsTemplate,
		&mapping.SSOFormsURL,
		&mapping.SSOPrivateKey,
		&mapping.SSOUseMappingCredentials,
		&mapping.SSOUsername,
		&mapping.SSOPassword,
		&mapping.InsecureSkipVerify,
	)

	mapping.SSOType = model.SSOType(sso)
	mapping.Protocol = model.Protocol(protocol)
	mapping.Domain, err = repo.domains.RetrieveOne("id = ?", domain)

	return
}

// NewSQLMappingRepository returns a new mapping repo with db and Mappings.
func NewSQLMappingRepository(db *sql.DB, domains DomainRepository) MappingRepository {
	return sqlMappingRepository{
		base:    base{db},
		domains: domains,
	}
}
