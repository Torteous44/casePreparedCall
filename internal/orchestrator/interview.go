package orchestrator

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/gorilla/websocket"
	"github.com/torteous44/callservice/internal/audio/stt"
	"github.com/torteous44/callservice/internal/audio/vad"
	"github.com/torteous44/callservice/internal/sessionstate"
)

// InterviewSession represents an active interview session
type InterviewSession struct {
	ID              string                     `json:"id"`
	StartTime       time.Time                  `json:"start_time"`
	Status          string                     `json:"status"`
	StreamingSTT    *stt.StreamingSTT          `json:"-"`
	VAD             *vad.VAD                   `json:"-"`
	SessionState    *sessionstate.SessionState `json:"-"`
	WebSocketConn   *websocket.Conn            `json:"-"`
	AudioBuffer     []byte                     `json:"-"`
	TranscriptCount int                        `json:"transcript_count"`
	UtteranceCount  int                        `json:"utterance_count"`
	AssemblyAIID    string                     `json:"assemblyai_id,omitempty"` // Track AssemblyAI session ID
	mu              sync.RWMutex               `json:"-"`
	ctx             context.Context            `json:"-"`
	cancel          context.CancelFunc         `json:"-"`
}

// InterviewManager manages interview sessions
type InterviewManager struct {
	sessions map[string]*InterviewSession
	store    *sessionstate.Store
	mu       sync.RWMutex
	upgrader websocket.Upgrader
}

// NewInterviewManager creates a new interview manager
func NewInterviewManager() *InterviewManager {
	return &InterviewManager{
		sessions: make(map[string]*InterviewSession),
		store:    sessionstate.NewStore(),
		upgrader: websocket.Upgrader{
			CheckOrigin: func(r *http.Request) bool {
				// Allow all origins for testing - restrict in production
				return true
			},
		},
	}
}

// CreateSessionRequest represents the request to create a new session
type CreateSessionRequest struct {
	SampleRate int    `json:"sample_rate,omitempty"`
	Encoding   string `json:"encoding,omitempty"`
}

// CreateSessionResponse represents the response when creating a session
type CreateSessionResponse struct {
	SessionID    string `json:"session_id"`
	WebSocketURL string `json:"websocket_url"`
	Status       string `json:"status"`
}

// SessionStatusResponse represents session status information
type SessionStatusResponse struct {
	SessionID       string    `json:"session_id"`
	Status          string    `json:"status"`
	StartTime       time.Time `json:"start_time"`
	TranscriptCount int       `json:"transcript_count"`
	UtteranceCount  int       `json:"utterance_count"`
}

// InitializeSession creates a new interview session
func (im *InterviewManager) InitializeSession(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Methods", "POST, OPTIONS")
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type")

	if r.Method == "OPTIONS" {
		w.WriteHeader(http.StatusOK)
		return
	}

	if r.Method != "POST" {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req CreateSessionRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		// Use defaults if no body provided
		req.SampleRate = 16000
		req.Encoding = "pcm_s16le"
	}

	// Generate session ID
	sessionID := uuid.New().String()

	// Check if session already exists
	im.mu.Lock()
	if _, exists := im.sessions[sessionID]; exists {
		im.mu.Unlock()
		http.Error(w, "Session already exists", http.StatusConflict)
		return
	}

	// Create streaming configuration
	config := stt.StreamingConfig{
		SampleRate:                       req.SampleRate,
		Encoding:                         req.Encoding,
		FormatTurns:                      true,
		EndOfTurnConfidenceThreshold:     0.7,
		MinEndOfTurnSilenceWhenConfident: 1000,
		MaxTurnSilence:                   3000,
	}

	// Create context for the session
	ctx, cancel := context.WithCancel(context.Background())

	// Create session
	session := &InterviewSession{
		ID:           sessionID,
		StartTime:    time.Now(),
		Status:       "initialized",
		StreamingSTT: stt.NewStreamingSTT(config),
		VAD:          vad.NewVAD(),
		SessionState: sessionstate.NewSessionState(sessionID),
		ctx:          ctx,
		cancel:       cancel,
	}

	// Store session
	im.sessions[sessionID] = session
	im.mu.Unlock()

	log.Printf("[INFO] Interview session initialized: %s", sessionID)

	response := CreateSessionResponse{
		SessionID:    sessionID,
		WebSocketURL: fmt.Sprintf("ws://localhost:8080/ws/interview/%s", sessionID),
		Status:       "initialized",
	}

	json.NewEncoder(w).Encode(response)
}

// GetSessionStatus returns the current status of a session
func (im *InterviewManager) GetSessionStatus(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Access-Control-Allow-Origin", "*")

	sessionID := r.URL.Query().Get("session_id")
	if sessionID == "" {
		http.Error(w, "session_id required", http.StatusBadRequest)
		return
	}

	im.mu.RLock()
	session, exists := im.sessions[sessionID]
	im.mu.RUnlock()

	if !exists {
		http.Error(w, "Session not found", http.StatusNotFound)
		return
	}

	session.mu.RLock()
	response := SessionStatusResponse{
		SessionID:       session.ID,
		Status:          session.Status,
		StartTime:       session.StartTime,
		TranscriptCount: session.TranscriptCount,
		UtteranceCount:  session.UtteranceCount,
	}
	session.mu.RUnlock()

	json.NewEncoder(w).Encode(response)
}

// HandleWebSocket handles WebSocket connections for audio streaming
func (im *InterviewManager) HandleWebSocket(w http.ResponseWriter, r *http.Request) {
	// Extract session ID from URL path
	sessionID := r.URL.Path[len("/ws/interview/"):]

	im.mu.RLock()
	session, exists := im.sessions[sessionID]
	im.mu.RUnlock()

	if !exists {
		http.Error(w, "Session not found", http.StatusNotFound)
		return
	}

	// Check if session is already connected
	session.mu.Lock()
	if session.Status == "connected" {
		session.mu.Unlock()
		log.Printf("[WARN] Rejecting duplicate WebSocket connection for session: %s", sessionID)
		http.Error(w, "Session already connected", http.StatusConflict)
		return
	}

	// Upgrade connection to WebSocket
	conn, err := im.upgrader.Upgrade(w, r, nil)
	if err != nil {
		session.mu.Unlock()
		log.Printf("[ERROR] WebSocket upgrade failed: %v", err)
		return
	}

	// Close any existing connection
	if session.WebSocketConn != nil {
		log.Printf("[WARN] Closing existing WebSocket connection for session: %s", sessionID)
		session.WebSocketConn.Close()
	}

	session.WebSocketConn = conn
	session.Status = "connected"
	session.mu.Unlock()

	log.Printf("[INFO] WebSocket connected for session: %s", sessionID)

	// Start the session processing in a separate goroutine
	processingStarted := make(chan struct{})
	go func() {
		im.processSession(session)
		close(processingStarted)
	}()

	// Wait for processing to start
	<-processingStarted

	// Handle incoming audio data
	im.handleAudioStream(session)
}

// processSession handles the streaming STT and VAD processing
func (im *InterviewManager) processSession(session *InterviewSession) {
	// Connect to AssemblyAI streaming API
	err := session.StreamingSTT.Connect(session.ctx)
	if err != nil {
		log.Printf("[ERROR] Failed to connect streaming STT for session %s: %v", session.ID, err)
		return
	}

	log.Printf("[INFO] Streaming STT connected for session: %s", session.ID)

	// Listen for transcripts
	go func() {
		for {
			select {
			case transcript, ok := <-session.StreamingSTT.GetTranscripts():
				if !ok {
					return
				}
				im.handleTranscript(session, transcript)
			case err, ok := <-session.StreamingSTT.GetErrors():
				if !ok {
					return
				}
				if err != nil && err.Error() == "connection lost" {
					// Ignore expected connection lost errors during reconnection
					continue
				}
				log.Printf("[ERROR] Streaming STT error for session %s: %v", session.ID, err)
			case <-session.ctx.Done():
				return
			}
		}
	}()
}

// handleTranscript processes transcription results and detects utterances
func (im *InterviewManager) handleTranscript(session *InterviewSession, result stt.StreamingResult) {
	session.mu.Lock()
	defer session.mu.Unlock()

	// Track AssemblyAI session ID if we receive it
	if result.SessionID != "" && session.AssemblyAIID == "" {
		session.AssemblyAIID = result.SessionID
		log.Printf("[INFO] Tracking AssemblyAI session ID: %s for session: %s",
			result.SessionID, session.ID)
	}

	session.TranscriptCount++

	// Terminal output with clear formatting
	switch result.MessageType {
	case "SessionBegins":
		if result.SessionID != "" {
			session.AssemblyAIID = result.SessionID
			log.Printf("[INFO] New AssemblyAI session established: %s for session: %s",
				result.SessionID, session.ID)
		}
	case "PartialTranscript":
		if result.Text != "" {
			fmt.Printf("[%s] [PARTIAL] %s (conf: %.2f)\n",
				session.ID[:8], result.Text, result.Confidence)
		}
	case "FinalTranscript":
		if result.Text != "" {
			fmt.Printf("[%s] [FINAL] %s (conf: %.2f)\n",
				session.ID[:8], result.Text, result.Confidence)

			// Update session state with transcript
			session.SessionState.UpdateState("last_transcript", result.Text)
			session.SessionState.UpdateState("last_confidence", result.Confidence)
			session.SessionState.UpdateState("last_timestamp", time.Now())
		}
	case "Turn":
		if result.Text != "" {
			session.UtteranceCount++
			fmt.Printf("\n[%s] [UTTERANCE #%d] ================================\n",
				session.ID[:8], session.UtteranceCount)
			fmt.Printf("[%s] [UTTERANCE #%d] %s\n",
				session.ID[:8], session.UtteranceCount, result.Text)
			fmt.Printf("[%s] [UTTERANCE #%d] Confidence: %.2f\n",
				session.ID[:8], session.UtteranceCount, result.Confidence)
			fmt.Printf("[%s] [UTTERANCE #%d] ================================\n\n",
				session.ID[:8], session.UtteranceCount)

			// Update session state with complete utterance
			session.SessionState.UpdateState("utterance_count", session.UtteranceCount)
			session.SessionState.UpdateState("last_utterance", result.Text)
			session.SessionState.UpdateState("last_utterance_confidence", result.Confidence)
		}
	}

	// Send transcript back to frontend via WebSocket
	if session.WebSocketConn != nil {
		transcriptMsg := map[string]interface{}{
			"type":         "transcript",
			"message_type": result.MessageType,
			"text":         result.Text,
			"confidence":   result.Confidence,
			"is_final":     result.IsFinal,
			"timestamp":    time.Now().Unix(),
			"session_id":   session.AssemblyAIID,
		}

		if err := session.WebSocketConn.WriteJSON(transcriptMsg); err != nil {
			log.Printf("[ERROR] Failed to send transcript to client: %v", err)
		}
	}
}

// handleAudioStream processes incoming audio data from WebSocket
func (im *InterviewManager) handleAudioStream(session *InterviewSession) {
	defer func() {
		session.mu.Lock()
		if session.WebSocketConn != nil {
			session.WebSocketConn.Close()
			session.WebSocketConn = nil
		}
		session.Status = "disconnected"
		session.cancel()
		session.mu.Unlock()

		log.Printf("[INFO] WebSocket disconnected for session: %s", session.ID)
	}()

	var lastVoiceTime time.Time
	var silenceDuration time.Duration
	var continuousSilenceCount int
	const (
		endOfUtteranceThreshold = 1200 * time.Millisecond // 1.2 seconds of silence for end of utterance
		maxUtteranceDuration    = 30 * time.Second        // Maximum duration for a single utterance
		minUtteranceDuration    = 500 * time.Millisecond  // Minimum duration to consider as valid utterance
		silenceCheckInterval    = 100 * time.Millisecond  // How often to check silence duration
		maxSilenceCount         = 12                      // Number of silence intervals before ending utterance
		reconnectDelay          = 500 * time.Millisecond  // Delay before reconnecting
	)

	inUtterance := false
	utteranceStartTime := time.Time{}
	reconnectAttempts := 0
	const maxReconnectAttempts = 3

	reconnectSTT := func() error {
		if reconnectAttempts >= maxReconnectAttempts {
			return fmt.Errorf("exceeded maximum reconnection attempts")
		}
		reconnectAttempts++

		// Wait before reconnecting to allow final transcripts
		time.Sleep(reconnectDelay)

		// Close existing connection gracefully
		if err := session.StreamingSTT.Close(); err != nil {
			log.Printf("[WARN] Error closing STT connection: %v", err)
		}

		// Create new STT instance with same config
		config := session.StreamingSTT.GetConfig()
		session.StreamingSTT = stt.NewStreamingSTT(config)

		// Reconnect
		err := session.StreamingSTT.Connect(session.ctx)
		if err != nil {
			log.Printf("[ERROR] Failed to reconnect STT for session %s (attempt %d): %v",
				session.ID, reconnectAttempts, err)
			return err
		}

		log.Printf("[INFO] Successfully reconnected STT for session %s (attempt %d)",
			session.ID, reconnectAttempts)
		return nil
	}

	silenceTimer := time.NewTicker(silenceCheckInterval)
	defer silenceTimer.Stop()

	for {
		select {
		case <-session.ctx.Done():
			return
		case <-silenceTimer.C:
			if inUtterance {
				now := time.Now()
				silenceDuration = now.Sub(lastVoiceTime)

				// Only count silence after minimum utterance duration
				if now.Sub(utteranceStartTime) > minUtteranceDuration {
					if silenceDuration >= silenceCheckInterval {
						continuousSilenceCount++
						if continuousSilenceCount >= maxSilenceCount {
							inUtterance = false
							fmt.Printf("\n[%s] [UTTERANCE-END] ⸻⸻⸻⸻⸻⸻⸻⸻⸻⸻⸻⸻⸻⸻⸻⸻\n", session.ID[:8])
							fmt.Printf("[%s] [RESPONSE-TRIGGER] Preparing to generate response after %.1f seconds of silence\n",
								session.ID[:8], silenceDuration.Seconds())
							fmt.Printf("[%s] [UTTERANCE-END] ⸻⸻⸻⸻⸻⸻⸻⸻⸻⸻⸻⸻⸻⸻⸻⸻\n\n", session.ID[:8])

							// Wait for final transcript before reconnecting
							time.Sleep(500 * time.Millisecond)

							// Reset STT connection for new utterance
							if err := reconnectSTT(); err != nil {
								log.Printf("[ERROR] Failed to reset STT connection: %v", err)
								return
							}
						} else {
							fmt.Printf("[%s] [SILENCE] Waiting for more speech... (%.1fs) [%d/%d]\n",
								session.ID[:8], silenceDuration.Seconds(), continuousSilenceCount, maxSilenceCount)
						}
					}
				}
			}
		default:
			// Read audio data from WebSocket
			_, audioData, err := session.WebSocketConn.ReadMessage()
			if err != nil {
				if websocket.IsCloseError(err, websocket.CloseNormalClosure, websocket.CloseGoingAway) {
					log.Printf("[INFO] WebSocket closed normally for session %s", session.ID)
					return
				}
				log.Printf("[ERROR] Error reading WebSocket message: %v", err)
				return
			}

			// Process with VAD
			hasVoice, err := session.VAD.DetectActivity(audioData)
			if err != nil {
				log.Printf("[ERROR] VAD error: %v", err)
				continue
			}

			now := time.Now()

			// Check for maximum utterance duration
			if inUtterance && now.Sub(utteranceStartTime) > maxUtteranceDuration {
				inUtterance = false
				fmt.Printf("\n[%s] [UTTERANCE-END] ⸻⸻⸻⸻⸻⸻⸻⸻⸻⸻⸻⸻⸻⸻⸻⸻\n", session.ID[:8])
				fmt.Printf("[%s] [MAX-DURATION] Maximum utterance duration reached (30s)\n", session.ID[:8])
				fmt.Printf("[%s] [UTTERANCE-END] ⸻⸻⸻⸻⸻⸻⸻⸻⸻⸻⸻⸻⸻⸻⸻⸻\n\n", session.ID[:8])

				// Wait for final transcript before reconnecting
				time.Sleep(500 * time.Millisecond)

				// Reset STT connection for new utterance
				if err := reconnectSTT(); err != nil {
					log.Printf("[ERROR] Failed to reset STT connection: %v", err)
					return
				}
				continue
			}

			// Send to streaming STT if voice detected
			if hasVoice {
				err = session.StreamingSTT.SendAudio(audioData)
				if err != nil {
					log.Printf("[ERROR] Error sending audio to STT: %v", err)

					// Attempt to reconnect on error
					if err := reconnectSTT(); err != nil {
						log.Printf("[ERROR] Failed to recover STT connection: %v", err)
						return
					}
					continue
				}

				lastVoiceTime = now
				continuousSilenceCount = 0 // Reset silence counter when voice is detected

				if !inUtterance {
					inUtterance = true
					utteranceStartTime = now
					reconnectAttempts = 0 // Reset reconnect counter for new utterance
					fmt.Printf("\n[%s] [UTTERANCE-START] User started speaking\n", session.ID[:8])
				}

				fmt.Printf("[%s] [VOICE] Audio chunk sent to STT (%d bytes)\n",
					session.ID[:8], len(audioData))
			}
		}
	}
}

// CloseSession terminates a session
func (im *InterviewManager) CloseSession(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Access-Control-Allow-Origin", "*")

	sessionID := r.URL.Query().Get("session_id")
	if sessionID == "" {
		http.Error(w, "session_id required", http.StatusBadRequest)
		return
	}

	im.mu.Lock()
	session, exists := im.sessions[sessionID]
	if exists {
		// Only close if session exists and is connected
		if session.Status == "connected" {
			session.mu.Lock()
			if session.WebSocketConn != nil {
				session.WebSocketConn.Close()
				session.WebSocketConn = nil
			}
			session.Status = "disconnected"
			session.cancel()
			session.mu.Unlock()

			// Close streaming STT
			session.StreamingSTT.Close()
		}
		delete(im.sessions, sessionID)
	}
	im.mu.Unlock()

	if !exists {
		http.Error(w, "Session not found", http.StatusNotFound)
		return
	}

	log.Printf("[INFO] Interview session closed: %s", sessionID)

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{
		"status":     "closed",
		"session_id": sessionID,
	})
}
