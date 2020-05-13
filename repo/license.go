package repo

import (
	"database/sql"
	"fmt"
	"github.com/cyolo-core/cmd/dex/model"
)

// LicenseRepository is a repo for license objects.
type LicenseRepository interface {
	// Retrieve finds the first license in the db.
	Retrieve() (*model.License, error)
	// Update updates license to license from m.
	Update(m M) (*model.License, error)
	// CreateIfNotExists creates a license from data if it does not exist.
	CreateIfNotExists(m M) (*model.License, error)
	// CreateTable creates the licenses table if it does not exist.
	CreateTable() error
}

// sqlLicenseRepository is a sql implementation for LicenseRepository.
type sqlLicenseRepository struct{ base }

// CreateIfNotExists creates a license from data if it does not exist.
func (repo sqlLicenseRepository) CreateIfNotExists(m M) (license *model.License, err error) {
	// check if license already exists:
	if license, err = repo.Retrieve(); err == nil {
		return
	}
	// no valid license exists, update:
	return repo.Update(m)
}

// Update updates license to license from data.
func (repo sqlLicenseRepository) Update(m M) (license *model.License, err error) {
	// get license data from m:
	data, ok := m["data"].(string)

	if !ok {
		err = fmt.Errorf("missing %q paramter", "data")
		return
	}

	// parse license to assert validity:
	if license, err = model.ParseLicense(data); err != nil {
		return
	}

	// replace the current license:
	if err = repo.Exec("replace into licenses (id, license) values (?,?)", 0, license); err == nil {
		license, err = repo.Retrieve()
	}
	return
}

// CreateTable creates the licenses table if it does not exist.
func (repo sqlLicenseRepository) CreateTable() error {
	return repo.Exec("" +
		"create table if not exists licenses (" +
		"id integer primary key," +
		"license blob not null" +
		");",
	)
}

// ScanLicense attempts to scan out a license from scanner s.
func (repo sqlLicenseRepository) ScanLicense(s Scanner) (license *model.License, err error) {
	license = &model.License{}
	err = s.Scan(license)
	return
}

// NewSQLLicenseRepository returns a new license repo with db.
func NewSQLLicenseRepository(db *sql.DB) LicenseRepository {
	return sqlLicenseRepository{base{db}}
}
