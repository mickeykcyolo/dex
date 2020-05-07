package service

import (
	"errors"
	"fmt"
	"github.com/cyolo-core/cmd/idac/model"
	"github.com/cyolo-core/cmd/idac/repo"
	"strconv"
	"strings"
)

// CrudCallback is an optional callback of crud operations
type CrudCallback func(kind string)

// CrudService enables creation, retrieval, deletion and updating of objects.
type CrudService interface {
	// Create creates an object with kind and params.
	Create(kind string, params repo.M) (interface{}, error)
	// CreateBulk creates objects in bulk.
	CreateBulk(kind string, params []repo.M) (interface{}, error)
	// Upsert creates an object with kind if it does not exist, updates otherwise.
	Upsert(kind string, params repo.M) (interface{}, error)
	// Retrieve retrieves objects of kind with filter.
	Retrieve(kind string, filter repo.M) (interface{}, error)
	// Retrieve retrieves an object of kind with id or name.
	RetrieveOne(kind string, idOrName interface{}) (interface{}, error)
	// Update updates an object with kind with id or name to params.
	Update(kind string, idOrName interface{}, params repo.M) (interface{}, error)
	// UpdatePolicy updates policy.
	UpdatePolicy(policy *model.Policy) error
	// UpdateProcess updates the current process.
	UpdateProcess() error
	// Delete deletes an object of kind with id or name.
	Delete(kind string, idOrName string) error
}

// crudServiceImpl is the implementation of CrudService
type crudServiceImpl struct {
	domains  repo.DomainRepository
	ldaps    repo.LDAPRepository
	mappings repo.MappingRepository
	policies repo.PolicyRepository
	users    repo.UserRepository
	logs     repo.LogEntryRepository
	licenses repo.LicenseRepository

	createCallback CrudCallback
	updateCallback CrudCallback
	deleteCallback CrudCallback
}

// make sure that the implementation actually implements the interface.
var _ CrudService = crudServiceImpl{}

// Create creates an object with kind and params.
func (svc crudServiceImpl) Create(kind string, params repo.M) (out interface{}, err error) {
	delete(params, "system")
	defer svc.callCreateCallback(kind)
	switch kind {
	case "domain", "domains":
		return svc.domains.Create(params)
	case "ldap", "ldaps":
		if ldap, err := svc.ldaps.Create(params); err == nil {
			return ldap, ldap.Verify()
		} else {
			return nil, err
		}
	case "mapping", "mappings":
		return svc.mappings.Create(params)
	case "policy", "policies":
		return svc.policies.Create(params)
	case "user", "users":
		license, err := svc.licenses.Retrieve()
		// check license validity:
		if err != nil || license == nil {
			return nil, errors.New("license: invalid or missing license")
		}
		// check that the license has enough available seats:
		if svc.users.Count("enabled = ? and system = ?", true, false) > license.Users {
			return nil, errors.New("license: not enough available seats")
		}
		return svc.users.Create(params)
	case "license", "licenses":
		return svc.licenses.Update(params)
	default:
		return nil, fmt.Errorf("svc: unknown kind %q", kind)
	}
}

// CreateBulk creates objects in bulk.
func (svc crudServiceImpl) CreateBulk(kind string, params []repo.M) (interface{}, error) {
	out := make([]interface{}, len(params))
	out = out[:0]

	for _, params := range params {
		delete(params, "system")
		if _, err := svc.Create(kind, params); err != nil {
			out = append(out, repo.M{"created": false, "reason": err.Error()})
			continue
		}
		out = append(out, repo.M{"created": true, "reason": ""})
	}

	return out, nil
}

// Upsert creates an object with kind if it does not exist, updates otherwise.
func (svc crudServiceImpl) Upsert(kind string, params repo.M) (out interface{}, err error) {
	if out, err = svc.RetrieveOne(kind, params["name"].(string)); err == nil {
		return svc.Update(kind, params["name"].(string), params)
	}
	return svc.Create(kind, params)
}

// Retrieve retrieves objects of kind with filter.
func (svc crudServiceImpl) Retrieve(kind string, filter repo.M) (interface{}, error) {
	switch kind {
	case "domain", "domains":
		where, args := buildWhereClause(filter)
		return svc.domains.Retrieve(where, args...)
	case "ldap", "ldaps":
		where, args := buildWhereClause(filter)
		return svc.ldaps.Retrieve(where, args...)
	case "mapping", "mappings":
		where, args := buildWhereClause(filter)
		return svc.mappings.Retrieve(where, args...)
	case "policy", "policies":
		where, args := buildWhereClause(filter)
		return svc.policies.Retrieve(where, args...)
	case "user", "users":
		where, args := buildWhereClause(filter)
		return svc.users.Retrieve(where, args...)
	case "protocol", "protocols":
		return model.SupportedProtocols, nil
	case "sso", "ssos":
		return model.SupportedSSO, nil
	case "auth_level", "auth_levels":
		return model.SupportedAuthLevels, nil
	case "log", "logs":
		where, args := buildWhereClause(filter)
		return svc.logs.Retrieve(where, args...)
	case "process", "processes":
		return model.SelfProcess(nil), nil
	case "license", "licenses":
		return svc.licenses.Retrieve()
	default:
		return nil, fmt.Errorf("svc: unknown kind %q", kind)
	}
}

// Retrieve retrieves an object of kind with id or name.
func (svc crudServiceImpl) RetrieveOne(kind string, idOrName interface{}) (interface{}, error) {
	where, arg := buildWhereClauseForIdOrName(idOrName)
	switch kind {
	case "domain", "domains":
		return svc.domains.RetrieveOne(where, arg)
	case "ldap", "ldaps":
		return svc.ldaps.RetrieveOne(where, arg)
	case "mapping", "mappings":
		return svc.mappings.RetrieveOne(where, arg)
	case "policy", "policies":
		return svc.policies.RetrieveOne(where, arg)
	case "user", "users":
		return svc.users.RetrieveOne(where, arg)
	case "process", "processes":
		return model.SelfProcess(nil), nil
	case "license", "licenses":
		return svc.licenses.Retrieve()
	default:
		return nil, fmt.Errorf("svc: unknown kind %q", kind)
	}
}

// Update updates an object with kind with id or name to params.
func (svc crudServiceImpl) Update(kind string, idOrName interface{}, params repo.M) (out interface{}, err error) {
	delete(params, "system")
	defer svc.callUpdateCallback(kind)
	where, arg := buildWhereClauseForIdOrName(idOrName)
	switch kind {
	case "domain", "domains":
		if err = svc.domains.Update(params, where, arg); err == nil {
			return svc.domains.RetrieveOne(where, arg)
		}
	case "ldap", "ldaps":
		if err = svc.ldaps.Update(params, where, arg); err == nil {
			if ldap, err := svc.ldaps.RetrieveOne(where, arg); err == nil {
				return ldap, ldap.Verify()
			}
		}
	case "mapping", "mappings":
		if err = svc.mappings.Update(params, where, arg); err == nil {
			return svc.mappings.RetrieveOne(where, arg)
		}
	case "policy", "policies":
		if err = svc.policies.Update(params, where, arg); err == nil {
			return svc.policies.RetrieveOne(where, arg)
		}
	case "user", "users":
		if enabled, ok := params["enabled"].(bool); ok && enabled {
			// a user is being enabled, check licensing:
			license, err := svc.licenses.Retrieve()
			// check license validity:
			if err != nil || license == nil {
				return nil, errors.New("license: invalid or missing license")
			}
			// check that the license has enough available seats:
			if svc.users.Count("enabled = ? and system = ?", true, false) > license.Users {
				return nil, errors.New("license: not enough available seats")
			}
		}
		if err = svc.users.Update(params, where, arg); err == nil {
			return svc.users.RetrieveOne(where, arg)
		}
	case "license", "licenses":
		if _, err = svc.licenses.Update(params); err == nil {
			return svc.licenses.Retrieve()
		}
	default:
		return nil, fmt.Errorf("svc: unknown kind %q", kind)
	}

	return
}

// UpdatePolicy updates policy.
func (svc crudServiceImpl) UpdatePolicy(policy *model.Policy) error {
	defer svc.callUpdateCallback("policy")
	return svc.policies.Save(policy)
}

// UpdateProcess updates the current process.
func (svc crudServiceImpl) UpdateProcess() error {
	return model.SelfProcess(nil).Update(nil)
}

// Delete deletes an object of kind with id or name.
func (svc crudServiceImpl) Delete(kind string, idOrName string) error {
	defer svc.callDeleteCallback(kind)
	where, arg := buildWhereClauseForIdOrName(idOrName)
	switch kind {
	case "domain", "domains":
		return svc.domains.Delete(where, arg)
	case "ldap", "ldaps":
		return svc.ldaps.Delete(where, arg)
	case "mapping", "mappings":
		return svc.mappings.Delete(where, arg)
	case "policy", "policies":
		return svc.policies.Delete(where, arg)
	case "user", "users":
		return svc.users.Delete(where, arg)
	case "process", "processes":
		return model.SelfProcess(nil).Kill()
	default:
		return fmt.Errorf("svc: unknown kind %q", kind)
	}
}

// callCreateCallback calls the create callback if it exists.
func (svc crudServiceImpl) callCreateCallback(kind string) {
	if svc.createCallback != nil {
		svc.createCallback(kind)
	}
}

// callUpdateCallback calls the update callback if it exists.
func (svc crudServiceImpl) callUpdateCallback(kind string) {
	if svc.updateCallback != nil {
		svc.updateCallback(kind)
	}
}

// callDeleteCallback calls the delete callback if it exists.
func (svc crudServiceImpl) callDeleteCallback(kind string) {
	if svc.deleteCallback != nil {
		svc.deleteCallback(kind)
	}
}

// buildWhereClauseForIdOrName builds the correct where clause for id or name.
func buildWhereClauseForIdOrName(idOrName interface{}) (string, interface{}) {
	switch idOrName := idOrName.(type) {
	case int:
		return "id = ?", idOrName
	case string:
		if i, err := strconv.Atoi(idOrName); err == nil {
			return "id = ?", i
		}
		return "name = ?", idOrName
	}
	return "", nil
}

// buildWhereClause builds a where clause according to filter.
func buildWhereClause(filter map[string]interface{}) (where string, args []interface{}) {
	if len(filter) > 0 {
		var keys []string
		for key, value := range filter {
			keys = append(keys, fmt.Sprintf("%s=?", key))
			args = append(args, value)
		}
		where = strings.Join(keys, " and ")
	}
	return
}

// NewCrudService returns a new CrudService
func NewCrudService(
	domains repo.DomainRepository,
	ldaps repo.LDAPRepository,
	mappings repo.MappingRepository,
	policies repo.PolicyRepository,
	users repo.UserRepository,
	logs repo.LogEntryRepository,
	licenses repo.LicenseRepository,
	create CrudCallback,
	update CrudCallback,
	delete CrudCallback) CrudService {
	return &crudServiceImpl{
		domains,
		ldaps,
		mappings,
		policies,
		users,
		logs,
		licenses,
		create,
		update,
		delete,
	}
}
