package stt

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"log"
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
	MessageType string  `json:"message_type"`
	Text        string  `json:"text,omitempty"`
	Confidence  float64 `json:"confidence,omitempty"`
	IsFinal     bool    `json:"is_final,omitempty"`
	TurnID      string  `json:"turn_id,omitempty"`
	StartTime   int64   `json:"start_time,omitempty"`
	EndTime     int64   `json:"end_time,omitempty"`
	SessionID   string  `json:"session_id,omitempty"`
}

// SessionBegins represents the session start message
type SessionBegins struct {
	MessageType string `json:"message_type"`
	SessionID   string `json:"session_id"`
	ExpiresAt   string `json:"expires_at"`
}

// AudioMessage represents an audio data message
type AudioMessage struct {
	MessageType string `json:"message_type"`
	AudioData   string `json:"audio_data"` // base64 encoded audio data
}

// ConfigUpdateMessage represents a configuration update message
type ConfigUpdateMessage struct {
	MessageType string          `json:"message_type"`
	Config      StreamingConfig `json:"config"`
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
	u, err := url.Parse("wss://api.assemblyai.com/v2/realtime/ws")
	if err != nil {
		return fmt.Errorf("failed to parse WebSocket URL: %w", err)
	}

	q := u.Query()
	q.Set("sample_rate", fmt.Sprintf("%d", s.config.SampleRate))
	u.RawQuery = q.Encode()

	// Set up headers
	headers := http.Header{}
	headers.Set("Authorization", s.apiKey)

	log.Printf("Connecting to AssemblyAI at %s", u.String())

	// Connect to WebSocket with retry logic
	var retryCount int
	maxRetries := 3
	retryDelay := time.Second

	for retryCount < maxRetries {
		conn, _, err := websocket.Dial(ctx, u.String(), &websocket.DialOptions{
			HTTPHeader: headers,
		})

		if err == nil {
			s.conn = conn
			s.isConnected = true

			// Start message handler
			go s.handleMessages(ctx)
			return nil
		}

		log.Printf("Connection attempt %d failed: %v", retryCount+1, err)
		retryCount++
		if retryCount < maxRetries {
			time.Sleep(retryDelay)
			retryDelay *= 2 // Exponential backoff
		}
	}

	return fmt.Errorf("failed to connect after %d retries: %w", maxRetries, err)
}

// SendAudio sends audio data to the streaming API
func (s *StreamingSTT) SendAudio(audioData []byte) error {
	s.mu.RLock()
	defer s.mu.RUnlock()

	if !s.isConnected || s.conn == nil {
		return fmt.Errorf("not connected")
	}

	// Create a context with timeout for the write operation
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// Create audio message
	msg := map[string]interface{}{
		"message_type": "AudioData",
		"audio_data":   base64.StdEncoding.EncodeToString(audioData),
	}

	data, err := json.Marshal(msg)
	if err != nil {
		return fmt.Errorf("failed to marshal audio message: %w", err)
	}

	// Send audio message
	if err := s.conn.Write(ctx, websocket.MessageText, data); err != nil {
		s.sendError(fmt.Errorf("failed to send audio: %w", err))
		return err
	}

	return nil
}

// UpdateConfig sends a configuration update during the session
func (s *StreamingSTT) UpdateConfig(config StreamingConfig) error {
	s.mu.RLock()
	defer s.mu.RUnlock()

	if !s.isConnected || s.conn == nil {
		return fmt.Errorf("not connected")
	}

	msg := ConfigUpdateMessage{
		MessageType: "UpdateConfiguration",
		Config:      config,
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
		"message_type": "ForceEndpoint",
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
	conn := s.conn
	s.isConnected = false
	s.conn = nil
	s.mu.Unlock()

	if conn == nil {
		return nil
	}

	// Send session termination message
	msg := map[string]string{
		"message_type": "SessionTermination",
	}

	data, err := json.Marshal(msg)
	if err == nil {
		if err := conn.Write(context.Background(), websocket.MessageText, data); err != nil {
			log.Printf("Failed to send SessionTermination: %v", err)
		}
	}

	// Close WebSocket connection
	err = conn.Close(websocket.StatusNormalClosure, "")

	// Close channels after connection is closed
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

	var currentSessionID string

	for {
		select {
		case <-ctx.Done():
			return
		default:
			s.mu.RLock()
			conn := s.conn
			if conn == nil {
				s.mu.RUnlock()
				s.errors <- fmt.Errorf("connection lost")
				s.reconnect(ctx)
				return
			}

			_, message, err := conn.Read(ctx)
			s.mu.RUnlock()

			if err != nil {
				if websocket.CloseStatus(err) == websocket.StatusNormalClosure {
					return
				}
				select {
				case s.errors <- fmt.Errorf("failed to read message: %w", err):
				default:
					log.Printf("[WARN] Dropping error: failed to read message: %v", err)
				}
				s.reconnect(ctx)
				return
			}

			// Parse message
			var baseMsg map[string]interface{}
			if err := json.Unmarshal(message, &baseMsg); err != nil {
				s.sendError(fmt.Errorf("failed to parse message: %w", err))
				continue
			}

			// Log raw message for debugging
			log.Printf("[DEBUG] Raw message: %s", string(message))

			msgType, ok := baseMsg["message_type"].(string)
			if !ok {
				s.sendError(fmt.Errorf("invalid message type"))
				continue
			}

			switch msgType {
			case "SessionBegins":
				var sessionBegins SessionBegins
				if err := json.Unmarshal(message, &sessionBegins); err != nil {
					s.sendError(fmt.Errorf("failed to parse SessionBegins: %w", err))
					continue
				}
				currentSessionID = sessionBegins.SessionID
				log.Printf("[INFO] Session established, ID: %s", sessionBegins.SessionID)

			case "Connected":
				log.Printf("[INFO] Successfully connected to AssemblyAI streaming service")

			case "PartialTranscript":
				var result StreamingResult
				if err := json.Unmarshal(message, &result); err != nil {
					s.sendError(fmt.Errorf("failed to parse PartialTranscript: %w", err))
					continue
				}
				result.MessageType = "PartialTranscript"
				result.IsFinal = false
				result.SessionID = currentSessionID
				if result.Text != "" {
					log.Printf("[DEBUG] Partial: %s", result.Text)
					s.transcripts <- result
				}

			case "FinalTranscript":
				var result StreamingResult
				if err := json.Unmarshal(message, &result); err != nil {
					s.sendError(fmt.Errorf("failed to parse FinalTranscript: %w", err))
					continue
				}
				result.MessageType = "FinalTranscript"
				result.IsFinal = true
				result.SessionID = currentSessionID
				if result.Text != "" {
					log.Printf("[INFO] Final: %s", result.Text)
					s.transcripts <- result
				}

			case "Error":
				var errorMsg struct {
					Type    string `json:"message_type"`
					Message string `json:"message"`
					Code    string `json:"error"`
				}
				if err := json.Unmarshal(message, &errorMsg); err != nil {
					s.sendError(fmt.Errorf("failed to parse error message: %w", err))
					continue
				}
				log.Printf("[ERROR] AssemblyAI error: %s (code: %s)", errorMsg.Message, errorMsg.Code)
				s.sendError(fmt.Errorf("server error: %s (code: %s)", errorMsg.Message, errorMsg.Code))

			case "SessionTerminated":
				log.Printf("[INFO] Session terminated by server")
				currentSessionID = ""
				return

			default:
				if msgType != "" {
					log.Printf("[DEBUG] Received message type: %s", msgType)
					log.Printf("[DEBUG] Raw message: %s", string(message))
				}
			}
		}
	}
}

// reconnect attempts to reestablish the WebSocket connection
func (s *StreamingSTT) reconnect(ctx context.Context) {
	s.mu.Lock()
	defer s.mu.Unlock()

	if s.conn != nil {
		s.conn.Close(websocket.StatusGoingAway, "reconnecting")
		s.conn = nil
	}
	s.isConnected = false

	// Try to reconnect
	err := s.Connect(ctx)
	if err != nil {
		s.errors <- fmt.Errorf("reconnection failed: %w", err)
	}
}

// GetDefaultStreamingConfig returns default configuration for streaming
func GetDefaultStreamingConfig() StreamingConfig {
	return StreamingConfig{
		SampleRate: 16000,
		Encoding:   "pcm_s16le",
	}
}

// Helper function for sending errors
func (s *StreamingSTT) sendError(err error) {
	select {
	case s.errors <- err:
	default:
		log.Printf("Dropping error: %v", err)
	}
}

// GetConfig returns the current streaming configuration
func (s *StreamingSTT) GetConfig() StreamingConfig {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.config
}
