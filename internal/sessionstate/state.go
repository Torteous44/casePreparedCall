package sessionstate

// SessionState represents the current state of a call session
type SessionState struct {
	ID     string                 `json:"id"`
	Status string                 `json:"status"`
	Data   map[string]interface{} `json:"data"`
	// TODO: Add more state fields
}

// NewSessionState creates a new session state
func NewSessionState(id string) *SessionState {
	return &SessionState{
		ID:     id,
		Status: "initialized",
		Data:   make(map[string]interface{}),
	}
}

// UpdateState updates the session state
func (s *SessionState) UpdateState(key string, value interface{}) {
	s.Data[key] = value
}

// GetState retrieves a value from the session state
func (s *SessionState) GetState(key string) (interface{}, bool) {
	value, exists := s.Data[key]
	return value, exists
}
