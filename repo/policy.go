package repo

import (
	"database/sql"
	"github.com/cyolo-core/cmd/idac/model"
)

// PolicyRepository is a repo for policy objects.
type PolicyRepository interface {
	// Retrieve finds all policies that match an optional where clause.
	Retrieve(where string, args ...interface{}) ([]*model.Policy, error)
	// ForMapping finds all policies that apply to a mapping id.
	ForMapping(id model.Identity) ([]*model.Policy, error)
	// ForMapping finds all policies that apply to a user id.
	ForUser(id model.Identity) (policies []*model.Policy, err error)
	// Retrieve finds the first policy that matches an optional where clause.
	RetrieveOne(where string, args ...interface{}) (*model.Policy, error)
	// Create creates a policy with information from map m.
	Create(m M) (*model.Policy, error)
	// CreateIfNotExists creates a policy with information from m if it
	// does not already exist.
	CreateIfNotExists(m M) (policy *model.Policy, created bool, err error)
	// Upsert updates a policy with name from m if exists, creates it otherwise.
	Upsert(m M) (policy *model.Policy, err error)
	// Update updates objects that match where with values from m.
	Update(m M, where string, args ...interface{}) error
	// Save persists a policy object as-is.
	Save(policy *model.Policy) error
	// Delete deletes all policies that match an optional where clause.
	Delete(where string, args ...interface{}) error
	// CreateTable creates the policies table if it does not exist.
	CreateTable() error
}

// sqlPolicyRepository is a sql implementation for PolicyRepository.
type sqlPolicyRepository struct {
	base
	users    UserRepository
	mappings MappingRepository
}

// Retrieve finds all policies that match an optional where clause.
func (repo sqlPolicyRepository) Retrieve(where string, args ...interface{}) (policies []*model.Policy, err error) {
	var (
		rows   *sql.Rows
		policy *model.Policy
	)
	policies = make([]*model.Policy, 0)
	if rows, err = repo.Query(buildStmt("select * from policies", where), args...); err == nil {
		defer rows.Close()
		for rows.Next() && err == nil {
			if policy, err = repo.ScanPolicy(rows); err == nil {
				policies = append(policies, policy)
			}
		}
	}
	return
}

// Retrieve finds the first policy that matches an optional where clause.
func (repo sqlPolicyRepository) RetrieveOne(where string, args ...interface{}) (policy *model.Policy, err error) {
	if row := repo.QueryRow(buildStmt("select * from policies", where), args...); row != nil {
		if policy, err = repo.ScanPolicy(row); err != nil {
			policy = nil
		}
	}
	return
}

// ForMapping finds all policies that apply to a mapping id.
func (repo sqlPolicyRepository) ForMapping(id model.Identity) (policies []*model.Policy, err error) {
	if rows, err := repo.Query("select policy_id from policy_to_mapping where mapping_id = ?", id); err == nil {
		defer rows.Close()
		for rows.Next() && err == nil {
			if err = rows.Scan(&id); err == nil {
				if policy, err := repo.RetrieveOne("id = ? and enabled = ?", id, true); err == nil {
					policies = append(policies, policy)
				}
			}
		}
	}
	return
}

// ForUser finds all policies that apply to a user id.
func (repo sqlPolicyRepository) ForUser(id model.Identity) (policies []*model.Policy, err error) {
	if rows, err := repo.Query("select policy_id from policy_to_user where user_id = ?", id); err == nil {
		defer rows.Close()
		for rows.Next() && err == nil {
			if err = rows.Scan(&id); err == nil {
				if policy, err := repo.RetrieveOne("id = ? and enabled = ?", id, true); err == nil {
					policies = append(policies, policy)
				}
			}
		}
	}
	return
}

// Create creates a policy with information from from map m.
func (repo sqlPolicyRepository) Create(m M) (policy *model.Policy, err error) {
	insert, args := buildInsertStmt("policies", m)
	if err = repo.Exec(insert, args...); err == nil {
		policy, err = repo.RetrieveOne("name = ?", m["name"])
	}
	return
}

// CreateIfNotExists creates a policy with information from map m,
// if it does not exist.
func (repo sqlPolicyRepository) CreateIfNotExists(m M) (policy *model.Policy, created bool, err error) {
	if policy, err = repo.RetrieveOne("name = ?", m["name"]); err == nil {
		return
	}
	if policy, err = repo.Create(m); err == nil {
		created = true
	}
	return
}

// Upsert updates a policy with name from m if exists, creates it otherwise.
func (repo sqlPolicyRepository) Upsert(m M) (policy *model.Policy, err error) {
	if policy, err = repo.RetrieveOne("name = ?", m["name"]); err == nil && policy != nil {
		if err = repo.Update(m, "name = ?", m["name"]); err == nil {
			return repo.RetrieveOne("name = ?", m["name"])
		}
	}
	return repo.Create(m)
}

// Update updates objects that match where with values from m.
func (repo sqlPolicyRepository) Update(m M, where string, args ...interface{}) error {
	updateStmt, updateArgs := buildUpdateStmt("policies", m)
	args = append(updateArgs, args...)
	return repo.Exec(buildStmt(updateStmt, where), args...)
}

// Save persists a policy object as-is.
func (repo sqlPolicyRepository) Save(policy *model.Policy) (err error) {
	return repo.DoTransaction(func(tx baseTx) (err error) {
		if err = tx.Exec("delete from policy_to_user where policy_id = ?", policy.Id); err != nil {
			return
		}
		if err = tx.Exec("delete from policy_to_mapping where policy_id = ?", policy.Id); err != nil {
			return
		}
		for _, user := range policy.Users {
			if err = tx.Exec("insert into policy_to_user (policy_id,user_id) values (?,?)", policy.Id, user.Id); err != nil {
				return
			}
		}
		for _, mapping := range policy.Mappings {
			if err = tx.Exec("insert into policy_to_mapping (policy_id,mapping_id) values (?,?)", policy.Id, mapping.Id); err != nil {
				return
			}
		}
		return tx.Exec("update policies set "+
			"name = ?,"+
			"system = ?,"+
			"auth_level = ?,"+
			"enabled = ? "+
			"where id = ?",
			policy.Name,
			policy.System,
			policy.AuthLevel,
			policy.Enabled,
			policy.Id,
		)
	})
}

// Delete deletes all policies that match an optional where clause.
func (repo sqlPolicyRepository) Delete(where string, args ...interface{}) error {
	return repo.Exec(buildStmt("delete from policies", where), args...)
}

// CreateTable creates the policies table if it does not exist.
func (repo sqlPolicyRepository) CreateTable() error {
	return repo.Exec("" +
		"create table if not exists policies (" +
		"id integer primary key," +
		"name text unique not null," +
		"system integer default 0," +
		"enabled integer default 0," +
		"ctime timestamp default current_timestamp," +
		"mtime timestamp default current_timestamp," +
		"auth_level integer default 0," +
		// type assertions:
		"check(system in (0,1,'true','false'))," +
		"check(enabled in (0,1,'true','false'))," +
		"check(auth_level in (0,1,2))" +
		");" +
		"" +
		"create table if not exists policy_to_mapping (" +
		"	policy_id integer references policies(id) on delete cascade," +
		"	mapping_id integer references mappings(id) on delete cascade" +
		");" +
		"" +
		"create table if not exists policy_to_user (" +
		"	policy_id integer references policies(id) on delete cascade," +
		"	user_id integer references users(id) on delete cascade" +
		");" +
		"" +
		"create unique index if not exists idx_policy_mapping on policy_to_mapping(policy_id, mapping_id);" +
		"create unique index if not exists idx_policy_user on policy_to_user(policy_id, user_id);" +
		"" +
		"create trigger if not exists update_policies_mtime after update on policies " +
		"begin " +
		"	update policies set mtime = current_timestamp where id = old.id;" +
		"end;",
	)
}

// ScanPolicy attempts to scan out a policy from scanner s.
func (repo sqlPolicyRepository) ScanPolicy(s Scanner) (policy *model.Policy, err error) {
	policy = model.NewPolicy("")
	err = s.Scan(
		&policy.Id,
		&policy.Name,
		&policy.System,
		&policy.Enabled,
		&policy.Ctime,
		&policy.Mtime,
		&policy.AuthLevel,
	)

	if err == nil {
		var (
			id      int
			rows    *sql.Rows
			user    *model.User
			mapping *model.Mapping
		)

		if rows, err = repo.Query("select mapping_id from policy_to_mapping where policy_id = ?", policy.Id); err == nil {
			defer rows.Close()
			for rows.Next() && err == nil {
				if err = rows.Scan(&id); err == nil {
					if mapping, err = repo.mappings.RetrieveOne("id = ?", id); err == nil {
						policy.Mappings = append(policy.Mappings, mapping)
					}
				}
			}
		}

		if rows, err = repo.Query("select user_id from policy_to_user where policy_id = ?", policy.Id); err == nil {
			defer rows.Close()
			for rows.Next() && err == nil {
				if err = rows.Scan(&id); err == nil {
					if user, err = repo.users.RetrieveOne("id = ?", id); err == nil {
						policy.Users = append(policy.Users, user)
					}
				}
			}
		}

	}
	return
}

// NewSQLPolicyRepository returns a new policy repo with db,
// mapping and user repositories.
func NewSQLPolicyRepository(
	db *sql.DB,
	mappings MappingRepository,
	users UserRepository) PolicyRepository {

	return sqlPolicyRepository{
		base:     base{db},
		mappings: mappings,
		users:    users,
	}
}
