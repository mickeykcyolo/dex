package repo

import (
	"database/sql"
	"github.com/cyolo-core/cmd/dex/model"
)

// UserRepository is a repo for user objects.
type UserRepository interface {
	// Count returns the number of users that match an optional where clause.
	Count(where string, args ...interface{}) int
	// Retrieve finds all users that match an optional where clause.
	Retrieve(where string, args ...interface{}) ([]*model.User, error)
	// Retrieve finds the first user that matches an optional where clause.
	RetrieveOne(where string, args ...interface{}) (*model.User, error)
	// Create creates a user with information from map m.
	Create(m M) (*model.User, error)
	// CreateIfNotExists creates a group with information from map m,
	// if it does not exist.
	CreateIfNotExists(m M) (user *model.User, err error)
	// Upsert updates a user with name from m if exists, creates it otherwise.
	Upsert(m M) (user *model.User, err error)
	// Update updates objects that match where with values from m.
	Update(m M, where string, args ...interface{}) error
	// Delete deletes all users that match an optional where clause.
	Delete(where string, args ...interface{}) error
	// CreateTable creates the users table if it does not exist.
	CreateTable() error
}

// sqlUserRepository is a sql implementation for UserRepository.
type sqlUserRepository struct{ base }

// Count returns the number of users that match an optional where clause.
func (repo sqlUserRepository) Count(where string, args ...interface{}) (count int) {
	if row := repo.QueryRow(buildStmt("select count(*) from users", where), args...); row != nil {
		row.Scan(&count)
	}
	return
}

// Retrieve finds all users that match an optional where clause.
func (repo sqlUserRepository) Retrieve(where string, args ...interface{}) (users []*model.User, err error) {
	var (
		rows *sql.Rows
		user *model.User
	)
	users = make([]*model.User, 0)
	if rows, err = repo.Query(buildStmt("select * from users", where), args...); err == nil {
		defer rows.Close()
		for rows.Next() && err == nil {
			if user, err = repo.ScanUser(rows); err == nil {
				users = append(users, user)
			}
		}
	}
	return
}

// Retrieve finds the first user that matches an optional where clause.
func (repo sqlUserRepository) RetrieveOne(where string, args ...interface{}) (user *model.User, err error) {
	if row := repo.QueryRow(buildStmt("select * from users", where), args...); row != nil {
		if user, err = repo.ScanUser(row); err != nil {
			user = nil
		}
	}
	return
}

// Create creates a user with information from from map m.
func (repo sqlUserRepository) Create(m M) (user *model.User, err error) {
	insert, args := buildInsertStmt("users", m)
	if err = repo.Exec(insert, args...); err == nil {
		user, err = repo.RetrieveOne("name = ?", m["name"])
	}
	return
}

// CreateIfNotExists creates a group with information from map m,
// if it does not exist.
func (repo sqlUserRepository) CreateIfNotExists(m M) (user *model.User, err error) {
	if user, err = repo.RetrieveOne("name = ?", m["name"]); err == nil {
		return
	}
	return repo.Create(m)
}

// Upsert updates a user with name from m if exists, creates it otherwise.
func (repo sqlUserRepository) Upsert(m M) (user *model.User, err error) {
	if user, err = repo.RetrieveOne("name = ?", m["name"]); err == nil && user != nil {
		if err = repo.Update(m, "name = ?", m["name"]); err == nil {
			return repo.RetrieveOne("name = ?", m["name"])
		}
	}
	return repo.Create(m)
}

// Update updates objects that match where with values from m.
func (repo sqlUserRepository) Update(m M, where string, args ...interface{}) error {
	updateStmt, updateArgs := buildUpdateStmt("users", m)
	args = append(updateArgs, args...)
	return repo.Exec(buildStmt(updateStmt, where), args...)
}

// Delete deletes all users that match an optional where clause.
func (repo sqlUserRepository) Delete(where string, args ...interface{}) error {
	return repo.Exec(buildStmt("delete from users", where), args...)
}

// CreateTable creates the users table if it does not exist.
func (repo sqlUserRepository) CreateTable() error {
	return repo.Exec("" +
		"create table if not exists users (" +
		"id integer primary key," +
		"name text unique not null," +
		"system integer default 0," +
		"enabled integer default 0," +
		"ctime timestamp default current_timestamp," +
		"mtime timestamp default current_timestamp," +
		"phone_number text default ''," +
		"totp_secret text default ''," +
		"totp_enabled integer default 1," +
		"totp_enrolled integer default 0," +
		"private_key text default ''," +
		"password_hash text default ''," +
		"supervisor integer references users(id) on delete set null," +
		"personal_desktop text default ''," +
		"enrolled integer default 0," +
		// users cannot be their own supervisors:
		"check(supervisor <> id)," +
		// type assertions:
		"check(system in (0,1,'true','false'))," +
		"check(enabled in (0,1,'true','false'))," +
		"check(totp_enabled in (0,1,'true','false'))," +
		"check(totp_enrolled in (0,1,'true','false'))," +
		"check(enrolled in (0,1,'true','false'))" +
		");" +
		"" +
		"create trigger if not exists update_users_mtime after update on users " +
		"begin " +
		"	update users set mtime = current_timestamp where id = old.id;" +
		"end;",
	)
}

// ScanUser attempts to scan out a user from scanner s.
func (repo sqlUserRepository) ScanUser(s Scanner) (user *model.User, err error) {

	// supervisorId may be null
	var supervisorId sql.NullInt64

	user = model.NewUser("")

	err = s.Scan(
		&user.Id,
		&user.Name,
		&user.System,
		&user.Enabled,
		&user.Ctime,
		&user.Mtime,
		&user.PhoneNumber,
		&user.TOTPSecret,
		&user.TOTPEnabled,
		&user.TOTPEnrolled,
		&user.PrivateKey,
		&user.PasswordHash,
		&supervisorId,
		&user.PersonalDesktop,
		&user.Enrolled,
	)

	if supervisorId.Valid {
		if super, err := repo.RetrieveOne("id = ?", supervisorId.Int64); err == nil {
			user.Supervisor = super
		}
	}

	return
}

// NewSQLUserRepository returns a new user repo.
func NewSQLUserRepository(db *sql.DB) UserRepository {
	return sqlUserRepository{base{db}}
}
