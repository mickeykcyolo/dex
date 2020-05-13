package model

// Policy represents which mappings are
// accessible to which users.
type Policy struct {
	// Meta ...
	Meta `json:",inline"`
	// Mappings is a list of mappings that
	// this policy applies to.
	Mappings []*Mapping `json:"mappings"`
	// Users is a list of users that
	// this policy applies to,
	Users []*User `json:"users"`
	// AuthLevel maintains the authentication
	// level required from users in this policy
	// to access the mappings in this policy.
	AuthLevel `json:"auth_level"`
}

// AppliesTo returns true if policy p applies to user u.
func (p *Policy) AppliesTo(u *User) bool {
	if p.AuthLevel == Anonymous {
		return true
	}
	for _, policyUser := range p.Users {
		if !(policyUser.Enabled && p.Enabled) {
			continue
		}
		if policyUser.Name == "anonymous" {
			return true
		}
		if (policyUser.Name == "authenticated" || policyUser.Equal(u.Meta)) &&
			u.AuthLevel >= p.AuthLevel {
			return true
		}
	}
	return false
}

// TheoreticallyAppliesTo returns true if policy p applies to
// user u, not taking into account the user's authentication level.
func (p *Policy) TheoreticallyAppliesTo(u *User) bool {
	if p.AuthLevel == Anonymous {
		return true
	}
	for _, policyUser := range p.Users {
		if !(policyUser.Enabled && p.Enabled) {
			continue
		}
		if policyUser.Name == "anonymous" {
			return true
		}
		if policyUser.Name == "authenticated" {
			return true
		}
		if policyUser.Equal(u.Meta) {
			return true
		}
	}
	return false
}

// NewPolicy returns a ne policy object with name.
func NewPolicy(name string) *Policy {
	return &Policy{
		Meta:     Meta{Name: name, Kind: "Policy"},
		Mappings: []*Mapping{},
		Users:    []*User{},
	}
}
