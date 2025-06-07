package transport

// Message represents a transport message
type Message struct {
	Type      string      `json:"type"`
	Payload   interface{} `json:"payload"`
	Timestamp int64       `json:"timestamp"`
}

// AudioMessage represents an audio data message
type AudioMessage struct {
	SessionID string `json:"session_id"`
	Data      []byte `json:"data"`
	Format    string `json:"format"`
}

// TextMessage represents a text message
type TextMessage struct {
	SessionID string `json:"session_id"`
	Text      string `json:"text"`
	Speaker   string `json:"speaker"`
}

// StatusMessage represents a status update message
type StatusMessage struct {
	SessionID string `json:"session_id"`
	Status    string `json:"status"`
	Details   string `json:"details"`
}
