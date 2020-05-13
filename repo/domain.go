package repo

import (
	"database/sql"
	"github.com/cyolo-core/cmd/idac/model"
)

// DomainRepository is a repo for domain objects.
type DomainRepository interface {
	// Retrieve finds all domains that match an optional where clause.
	Retrieve(where string, args ...interface{}) ([]*model.Domain, error)
	// Retrieve finds the first domain that matches an optional where clause.
	RetrieveOne(where string, args ...interface{}) (*model.Domain, error)
	// Create creates a domain with information from from map m.
	Create(m M) (*model.Domain, error)
	// CreateIfNotExists creates a domain with information from map m
	// if it does not already exist.
	CreateIfNotExists(m M) (*model.Domain, error)
	// Upsert updates a domain with name from m if exists, creates it otherwise.
	Upsert(m M) (domain *model.Domain, err error)
	// Update updates domain objects that match an optional where clause with information from map m.
	Update(m M, where string, args ...interface{}) error
	// Delete deletes all domains that match an optional where clause.
	Delete(where string, args ...interface{}) error
	// CreateTable creates the domains table if it does not exist.
	CreateTable() error
}

// sqlDomainRepository is a sql implementation for DomainRepository.
type sqlDomainRepository struct{ base }

// Retrieve finds all domains that match an optional where clause.
func (repo sqlDomainRepository) Retrieve(where string, args ...interface{}) (domains []*model.Domain, err error) {
	var (
		rows   *sql.Rows
		domain *model.Domain
	)
	domains = make([]*model.Domain, 0)
	if rows, err = repo.Query(buildStmt("select * from domains", where), args...); err == nil {
		defer rows.Close()
		for rows.Next() && err == nil {
			if domain, err = repo.ScanDomain(rows); err == nil {
				domains = append(domains, domain)
			}
		}
	}
	return
}

// Retrieve finds the first domain that matches an optional where clause.
func (repo sqlDomainRepository) RetrieveOne(where string, args ...interface{}) (domain *model.Domain, err error) {
	if row := repo.QueryRow(buildStmt("select * from domains", where), args...); row != nil {
		if domain, err = repo.ScanDomain(row); err != nil {
			domain = nil
		}
	}
	return
}

// Upsert updates a domain with name from m if exists, creates it otherwise.
func (repo sqlDomainRepository) Upsert(m M) (domain *model.Domain, err error) {
	if domain, err = repo.RetrieveOne("name = ?", m["name"]); err == nil && domain != nil {
		if err = repo.Update(m, "name = ?", m["name"]); err == nil {
			return repo.RetrieveOne("name = ?", m["name"])
		}
	}
	return repo.Create(m)
}

// Create creates a domain with information from from map m.
func (repo sqlDomainRepository) Create(m M) (domain *model.Domain, err error) {
	insert, args := buildInsertStmt("domains", m)
	if err = repo.Exec(insert, args...); err == nil {
		domain, err = repo.RetrieveOne("name = ?", m["name"])
	}
	return
}

// CreateIfNotExists creates a domain with information from map m
// if it does not exist.
func (repo sqlDomainRepository) CreateIfNotExists(m M) (domain *model.Domain, err error) {
	if domain, err = repo.RetrieveOne("name = ?", m["name"]); err == nil {
		return
	}
	return repo.Create(m)
}

// Update updates domain objects that match an optional where clause with information from map m.
func (repo sqlDomainRepository) Update(m M, where string, args ...interface{}) error {
	updateStmt, updateArgs := buildUpdateStmt("domains", m)
	args = append(updateArgs, args...)
	return repo.Exec(buildStmt(updateStmt, where), args...)
}

// Delete deletes all domains that match an optional where clause.
func (repo sqlDomainRepository) Delete(where string, args ...interface{}) error {
	return repo.Exec(buildStmt("delete from domains", where), args...)
}

// CreateTable creates the domains table if it does not exist.
func (repo sqlDomainRepository) CreateTable() error {
	return repo.Exec("" +
		"create table if not exists domains (" +
		"id integer primary key, " +
		"name text unique not null," +
		"system integer default 0," +
		"enabled integer default 0," +
		"ctime timestamp default current_timestamp," +
		"mtime timestamp default current_timestamp," +
		"certificate text not null," +
		"private_key text not null," +
		"check(system in (0,1,'true','false'))," +
		"check(enabled in (0,1,'true','false'))" +
		");" +
		"" +
		"create trigger if not exists update_domains_mtime after update on domains " +
		"begin " +
		"	update domains set mtime = current_timestamp where id = old.id;" +
		"end;",
	)
}

// ScanDomain attempts to scan out a domain from scanner s.
func (repo sqlDomainRepository) ScanDomain(s Scanner) (domain *model.Domain, err error) {
	domain = model.NewDomain("")
	err = s.Scan(
		&domain.Id,
		&domain.Name,
		&domain.System,
		&domain.Enabled,
		&domain.Ctime,
		&domain.Mtime,
		&domain.Certificate,
		&domain.PrivateKey,
	)
	return
}

// NewSQLDomainRepository returns a new domain repo with db.
func NewSQLDomainRepository(db *sql.DB) DomainRepository {
	return sqlDomainRepository{base{db}}
}
