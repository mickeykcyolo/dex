package service

import (
	"context"
	"github.com/cyolo-core/cmd/idac/model"
	"sync"
	"time"
)

// SessionService maintains user sessions.
type SessionService interface {
	// GenerateSession generates a session for a user.
	GenerateSession(user *model.User) *model.Session
	// Retrieve retrieves a session for user
	// from the session store.
	RetrieveSession(id string) (*model.Session, bool)
	// Destroy destroys a session with id.
	DestroySession(id string)
	// StartGC starts session garbage collection
	// in the background.
	StartGC(ctx context.Context)
}

// memSessionService implements an in-memory SessionService.
type memSessionService struct {
	// idle represents the maximum idle time
	// that a session can have.
	idle time.Duration
	// total represents the maximum lifetime
	// that a session can have.
	total time.Duration
	// mu protects the sessions map below
	mu sync.RWMutex
	// sessions is a mapping between ids and user sessions.
	sessions map[string]*model.Session
}

// GenerateSession generates a session for a user.
func (svc *memSessionService) GenerateSession(user *model.User) (ses *model.Session) {
	ses = model.GenerateSession(user)
	svc.mu.Lock()
	svc.sessions[ses.Id] = ses
	svc.mu.Unlock()
	return
}

// Retrieve retrieves a session for user from the session store.
func (svc *memSessionService) RetrieveSession(id string) (s *model.Session, ok bool) {
	svc.mu.RLock()
	s, ok = svc.sessions[id]
	svc.mu.RUnlock()

	if ok && svc.isValid(s) {
		// touch session atime
		s.Atime = now()
		return
	}

	svc.DestroySession(id)
	return
}

// DestroySession destroys a session with id.
func (svc *memSessionService) DestroySession(id string) {
	svc.mu.Lock()
	delete(svc.sessions, id)
	svc.mu.Unlock()
}

// StartGC starts session garbage collection in the background.
func (svc *memSessionService) StartGC(ctx context.Context) {
	go func() {
		for {
			select {
			case <-ctx.Done():
				return
			default:
				svc.mu.RLock()
				sessions := make([]*model.Session, len(svc.sessions))
				sessions = sessions[:0]

				for _, s := range svc.sessions {
					if !svc.isValid(s) {
						sessions = append(sessions, s)
					}
				}

				svc.mu.RUnlock()

				if len(sessions) > 0 {
					svc.mu.Lock()
					for _, s := range sessions {
						delete(svc.sessions, s.Id)
					}
					svc.mu.Unlock()
				}

				time.Sleep(svc.idle)
			}
		}
	}()
}

// isValid returns true if session s is still valid under svc.total and svc.idle.
func (svc *memSessionService) isValid(s *model.Session) bool {
	return now().Sub(s.Ctime) < svc.total && now().Sub(s.Atime) < svc.idle
}

// NewInMemSessionService returns a new in-mem session service with total
// and idle session lifetimes.
func NewInMemSessionService(total, idle time.Duration) SessionService {
	return &memSessionService{
		idle:     idle,
		total:    total,
		sessions: map[string]*model.Session{},
	}
}
