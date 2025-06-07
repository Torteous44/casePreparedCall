package sessionstate

import "sync"

// Store manages session states
type Store struct {
	sessions map[string]*SessionState
	mu       sync.RWMutex
}

// NewStore creates a new session store
func NewStore() *Store {
	return &Store{
		sessions: make(map[string]*SessionState),
	}
}

// CreateSession creates a new session
func (s *Store) CreateSession(id string) *SessionState {
	s.mu.Lock()
	defer s.mu.Unlock()

	session := NewSessionState(id)
	s.sessions[id] = session
	return session
}

// GetSession retrieves a session by ID
func (s *Store) GetSession(id string) (*SessionState, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	session, exists := s.sessions[id]
	return session, exists
}

// DeleteSession removes a session
func (s *Store) DeleteSession(id string) {
	s.mu.Lock()
	defer s.mu.Unlock()

	delete(s.sessions, id)
}
