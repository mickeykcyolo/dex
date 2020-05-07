package model

import "time"

// Meta struct is embedded in all other models.
type Meta struct {
	// id represents the unique identifier of the object.
	Id Identity `json:"id"`
	// Kind represents the kind of the object.
	Kind string `json:"kind"`
	// Name represents the unique name of the object.
	Name string `json:"name"`
	// System maintains whether the object is a system object.
	System bool `json:"system"`
	// Ctime is the object creation time.
	Ctime time.Time `json:"ctime"`
	// Mtime is the object last modified time.
	Mtime time.Time `json:"mtime"`
	// Enabled maintains whether or not the
	// object is effectively enabled.
	Enabled bool `json:"enabled"`
}

// Equal returns true if meta m equals other meta.
func (m Meta) Equal(other Meta) bool {
	return m.Kind == other.Kind && // objects of different kinds cannot be equal
		(m.Id == other.Id || m.Name == other.Name) && // id and name are unique
		m.System == other.System // system objects cannot equal non-system objects
}
