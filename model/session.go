package model

import (
	"github.com/cyolo-core/pkg/dict"
	uuid "github.com/satori/go.uuid"
	"sync"
	"time"
)

// Session represents a user session.
type Session struct {
	mu sync.RWMutex
	// id is the session unique identifier.
	Id string `json:"-"`
	// User is the user that this session belongs to.
	*User `json:",inline"`
	// Ctime is the session creation time.
	Ctime time.Time `json:"-"`
	// Atime is the session last access time.
	Atime time.Time `json:"-"`
	// Data is arbitrary data that can be stored
	// within the session.
	data dict.Dict `json:"data"`
}

// GenerateSession returns a new session object.
func GenerateSession(user *User) *Session {
	now := time.Now()
	return &Session{
		Id:    uuid.Must(uuid.NewV4(), nil).String(),
		User:  user,
		Ctime: now,
		Atime: now,
		data:  make(dict.Dict),
	}
}

func (s *Session) BoolValue(key string, defaultValue ...bool) bool {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.data.BoolValue(key, defaultValue...)
}

func (s *Session) SetBoolValue(key string, value bool) {
	s.mu.Lock()
	s.data.SetBoolValue(key, value)
	s.mu.Unlock()
}

func (s *Session) StringValue(key string, defaultValue ...string) string {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.data.StringValue(key, defaultValue...)
}

func (s *Session) SetStringValue(key string, value string) {
	s.mu.Lock()
	s.data.SetStringValue(key, value)
	s.mu.Unlock()
}

func (s *Session) TimeValue(key string, defaultValue ...time.Time) time.Time {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.data.TimeValue(key, defaultValue...)
}

func (s *Session) SetTimeValue(key string, value time.Time) {
	s.mu.Lock()
	s.data.SetTimeValue(key, value)
	s.mu.Unlock()
}

func (s *Session) IntValue(key string, defaultValue ...int) int {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.data.IntValue(key, defaultValue...)
}

func (s *Session) SetIntValue(key string, value int) {
	s.mu.Lock()
	s.data.SetIntValue(key, value)
	s.mu.Unlock()
}

func (s *Session) InterfaceValue(key string, defaultValue ...interface{}) interface{} {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.data.InterfaceValue(key, defaultValue...)
}

func (s *Session) SetInterfaceValue(key string, value interface{}) {
	s.mu.Lock()
	s.data.SetInterfaceValue(key, value)
	s.mu.Unlock()
}
