package stt

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"os"
	"sync"
	"time"

	"github.com/coder/websocket"
)

// StreamingSTT handles real-time speech-to-text using AssemblyAI's streaming API
type StreamingSTT struct {
	apiKey      string
	conn        *websocket.Conn
	mu          sync.RWMutex
	isConnected bool
	transcripts chan StreamingResult
	errors      chan error
	config      StreamingConfig
}

// StreamingConfig holds configuration for the streaming session
type StreamingConfig struct {
	SampleRate                       int     `json:"sample_rate"`
	Encoding                         string  `json:"encoding,omitempty"`
	FormatTurns                      bool    `json:"format_turns,omitempty"`
	EndOfTurnConfidenceThreshold     float64 `json:"end_of_turn_confidence_threshold,omitempty"`
	MinEndOfTurnSilenceWhenConfident int     `json:"min_end_of_turn_silence_when_confident,omitempty"`
	MaxTurnSilence                   int     `json:"max_turn_silence,omitempty"`
}

// StreamingResult represents a transcription result from the streaming API
type StreamingResult struct {
	Type       string  `json:"type"`
	Text       string  `json:"text,omitempty"`
	Confidence float64 `json:"confidence,omitempty"`
	IsFinal    bool    `json:"is_final,omitempty"`
	TurnID     string  `json:"turn_id,omitempty"`
	StartTime  int64   `json:"start_time,omitempty"`
	EndTime    int64   `json:"end_time,omitempty"`
}

// SessionBegins represents the session start message
type SessionBegins struct {
	Type      string    `json:"type"`
	ID        string    `json:"id"`
	ExpiresAt time.Time `json:"expires_at"`
}

// AudioMessage represents an audio data message
type AudioMessage struct {
	Type string `json:"type"`
	Data string `json:"data"` // base64 encoded audio data
}

// ConfigUpdateMessage represents a configuration update message
type ConfigUpdateMessage struct {
	Type   string          `json:"type"`
	Config StreamingConfig `json:"config"`
}

// NewStreamingSTT creates a new streaming STT instance
func NewStreamingSTT(config StreamingConfig) *StreamingSTT {
	apiKey := os.Getenv("ASSEMBLYAI_API_KEY")
	if apiKey == "" {
		panic("ASSEMBLYAI_API_KEY environment variable is not set")
	}

	return &StreamingSTT{
		apiKey:      apiKey,
		config:      config,
		transcripts: make(chan StreamingResult, 100),
		errors:      make(chan error, 10),
	}
}

// Connect establishes a WebSocket connection to AssemblyAI streaming API
func (s *StreamingSTT) Connect(ctx context.Context) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if s.isConnected {
		return fmt.Errorf("already connected")
	}

	// Build WebSocket URL with query parameters
	u, err := url.Parse("wss://streaming.assemblyai.com/v3/ws")
	if err != nil {
		return fmt.Errorf("failed to parse WebSocket URL: %w", err)
	}

	q := u.Query()
	q.Set("sample_rate", fmt.Sprintf("%d", s.config.SampleRate))
	if s.config.Encoding != "" {
		q.Set("encoding", s.config.Encoding)
	}
	if s.config.FormatTurns {
		q.Set("format_turns", "true")
	}
	u.RawQuery = q.Encode()

	// Set up headers
	headers := http.Header{}
	headers.Set("Authorization", s.apiKey)

	// Connect to WebSocket
	conn, _, err := websocket.Dial(ctx, u.String(), &websocket.DialOptions{
		HTTPHeader: headers,
	})
	if err != nil {
		return fmt.Errorf("failed to connect to WebSocket: %w", err)
	}

	s.conn = conn
	s.isConnected = true

	// Start message handler
	go s.handleMessages(ctx)

	return nil
}

// SendAudio sends audio data to the streaming API
func (s *StreamingSTT) SendAudio(audioData []byte) error {
	s.mu.RLock()
	defer s.mu.RUnlock()

	if !s.isConnected || s.conn == nil {
		return fmt.Errorf("not connected")
	}

	// Encode audio data to base64
	encodedData := base64.StdEncoding.EncodeToString(audioData)

	// Send audio message
	return s.conn.Write(context.Background(), websocket.MessageText, []byte(encodedData))
}

// UpdateConfig sends a configuration update during the session
func (s *StreamingSTT) UpdateConfig(config StreamingConfig) error {
	s.mu.RLock()
	defer s.mu.RUnlock()

	if !s.isConnected || s.conn == nil {
		return fmt.Errorf("not connected")
	}

	msg := ConfigUpdateMessage{
		Type:   "UpdateConfiguration",
		Config: config,
	}

	data, err := json.Marshal(msg)
	if err != nil {
		return fmt.Errorf("failed to marshal config update: %w", err)
	}

	return s.conn.Write(context.Background(), websocket.MessageText, data)
}

// ForceEndpoint manually forces an endpoint in the transcription
func (s *StreamingSTT) ForceEndpoint() error {
	s.mu.RLock()
	defer s.mu.RUnlock()

	if !s.isConnected || s.conn == nil {
		return fmt.Errorf("not connected")
	}

	msg := map[string]string{
		"type": "ForceEndpoint",
	}

	data, err := json.Marshal(msg)
	if err != nil {
		return fmt.Errorf("failed to marshal force endpoint: %w", err)
	}

	return s.conn.Write(context.Background(), websocket.MessageText, data)
}

// GetTranscripts returns a channel for receiving transcription results
func (s *StreamingSTT) GetTranscripts() <-chan StreamingResult {
	return s.transcripts
}

// GetErrors returns a channel for receiving errors
func (s *StreamingSTT) GetErrors() <-chan error {
	return s.errors
}

// Close gracefully terminates the streaming session
func (s *StreamingSTT) Close() error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if !s.isConnected || s.conn == nil {
		return nil
	}

	// Send session termination message
	msg := map[string]string{
		"type": "SessionTermination",
	}

	data, err := json.Marshal(msg)
	if err == nil {
		s.conn.Write(context.Background(), websocket.MessageText, data)
	}

	// Close WebSocket connection
	err = s.conn.Close(websocket.StatusNormalClosure, "")
	s.isConnected = false
	s.conn = nil

	// Close channels
	close(s.transcripts)
	close(s.errors)

	return err
}

// handleMessages processes incoming WebSocket messages
func (s *StreamingSTT) handleMessages(ctx context.Context) {
	defer func() {
		if r := recover(); r != nil {
			s.errors <- fmt.Errorf("message handler panic: %v", r)
		}
	}()

	for {
		select {
		case <-ctx.Done():
			return
		default:
			if s.conn == nil {
				return
			}

			_, message, err := s.conn.Read(ctx)
			if err != nil {
				s.errors <- fmt.Errorf("failed to read message: %w", err)
				return
			}

			// Parse message
			var baseMsg map[string]interface{}
			if err := json.Unmarshal(message, &baseMsg); err != nil {
				s.errors <- fmt.Errorf("failed to parse message: %w", err)
				continue
			}

			msgType, ok := baseMsg["type"].(string)
			if !ok {
				s.errors <- fmt.Errorf("invalid message type")
				continue
			}

			switch msgType {
			case "SessionBegins":
				var sessionBegins SessionBegins
				if err := json.Unmarshal(message, &sessionBegins); err != nil {
					s.errors <- fmt.Errorf("failed to parse SessionBegins: %w", err)
					continue
				}
				// Session started successfully

			case "Turn":
				var result StreamingResult
				if err := json.Unmarshal(message, &result); err != nil {
					s.errors <- fmt.Errorf("failed to parse Turn: %w", err)
					continue
				}
				result.Type = "Turn"
				s.transcripts <- result

			case "PartialTranscript":
				var result StreamingResult
				if err := json.Unmarshal(message, &result); err != nil {
					s.errors <- fmt.Errorf("failed to parse PartialTranscript: %w", err)
					continue
				}
				result.Type = "PartialTranscript"
				result.IsFinal = false
				s.transcripts <- result

			case "FinalTranscript":
				var result StreamingResult
				if err := json.Unmarshal(message, &result); err != nil {
					s.errors <- fmt.Errorf("failed to parse FinalTranscript: %w", err)
					continue
				}
				result.Type = "FinalTranscript"
				result.IsFinal = true
				s.transcripts <- result

			case "Termination":
				return

			default:
				s.errors <- fmt.Errorf("unknown message type: %s", msgType)
			}
		}
	}
}

// GetDefaultStreamingConfig returns default configuration for streaming
func GetDefaultStreamingConfig() StreamingConfig {
	return StreamingConfig{
		SampleRate:                       16000,
		Encoding:                         "pcm_s16le",
		FormatTurns:                      true,
		EndOfTurnConfidenceThreshold:     0.7,
		MinEndOfTurnSilenceWhenConfident: 1000,
		MaxTurnSilence:                   3000,
	}
}
