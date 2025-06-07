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

// Lesson object structures based on system design

// LessonObject represents the complete case interview definition
type LessonObject struct {
	LessonID                 string   `json:"lesson_id"`
	CaseID                   string   `json:"case_id"`
	CaseImage                string   `json:"case_image"`
	CaseType                 string   `json:"case_type"`
	CaseLevel                string   `json:"case_level"`
	CaseCompany              string   `json:"case_company"`
	CaseDescription          string   `json:"case_description"`
	CasePrompt               string   `json:"case_prompt"`
	CasePromptAdditionalInfo string   `json:"case_prompt_additional_information"`
	Questions                []string `json:"questions"`
	CaseIntroduction         string   `json:"case_introduction"` // Links to IntroductionObject
	CaseConclusion           string   `json:"case_conclusion"`   // Links to ConclusionObject
}

// IntroductionObject represents the case introduction phase
type IntroductionObject struct {
	IntroductionID             string `json:"introduction_id"`
	IntroductionCasePrompt     string `json:"introduction_case_prompt"`
	IntroductionAdditionalInfo string `json:"introduction_additional_information"`
	IntroductionGuideSteps     string `json:"introduction_guide_steps"`
	IntroductionQuestionPrompt string `json:"introduction_question_prompt"`
}

// QuestionObject represents a single interview question
type QuestionObject struct {
	QuestionID         string              `json:"question_id"`
	QuestionPrompt     string              `json:"question_prompt"`
	ExpectedComponents []ExpectedComponent `json:"expected_components"`
	GuideSteps         string              `json:"guide_steps"`
	Hints              []string            `json:"hints"`
	FollowUps          []string            `json:"follow_ups"`
	Clarifiers         []string            `json:"clarifiers"`
}

// ExpectedComponent represents a key area the candidate should mention
type ExpectedComponent struct {
	Title       string `json:"title"`
	Description string `json:"description"`
}

// GuideStepsObject represents the expected structural format for answers
type GuideStepsObject struct {
	GuideSteps []GuideStep `json:"guide_steps"`
}

// GuideStep represents a single step in the guide
type GuideStep struct {
	StepID          int    `json:"step_id"`
	Label           string `json:"label"`
	Description     string `json:"description"`
	ClarifierPrompt string `json:"clarifier_prompt"`
}

// ConclusionObject represents the case conclusion phase
type ConclusionObject struct {
	ConclusionID             string `json:"conclusion_id"`
	FarewellScript           string `json:"farewell_script"`
	NextStepsScript          string `json:"next_steps_script"`
	PostCaseQuestionResponse string `json:"post_case_question_response"`
}

// PersonaObject represents the AI interviewer's behavior and tone
type PersonaObject struct {
	CaseInterviewCompany string `json:"case_interview_company"`
	InterviewerTone      string `json:"interviewer_tone"`
	GreetingStyle        string `json:"greeting_style"`
	GeneralPersona       string `json:"general_persona"`
}

// SessionStateObject tracks the runtime context of the interview
type SessionStateObject struct {
	CurrentQuestion   int      `json:"current_question"`
	SilenceTimer      int      `json:"silence_timer"`
	HintsUsed         int      `json:"hints_used"`
	ComponentsHit     []string `json:"components_hit"`
	StepsHit          []string `json:"steps_hit"`
	FollowUpsUsed     []int    `json:"follow_ups_used"`
	UserReady         bool     `json:"user_ready"`
	UserReadyQuestion string   `json:"user_ready_question"`
	Completed         bool     `json:"completed"`
}

// TranscriptEntry represents a single transcript entry
type TranscriptEntry struct {
	Timestamp  time.Time `json:"timestamp"`
	Type       string    `json:"type"` // "partial", "final", "utterance"
	Text       string    `json:"text"`
	Confidence float64   `json:"confidence"`
	SessionID  string    `json:"session_id"`
}

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

	// Lesson and Context Data
	Lesson                *LessonObject                `json:"lesson"`
	Introduction          *IntroductionObject          `json:"introduction"`
	Questions             []*QuestionObject            `json:"questions"`
	GuideStepsMap         map[string]*GuideStepsObject `json:"guide_steps_map"`
	Conclusion            *ConclusionObject            `json:"conclusion"`
	Persona               *PersonaObject               `json:"persona"`
	InterviewSessionState *SessionStateObject          `json:"interview_session_state"`

	// Ephemeral transcript storage
	Transcript []TranscriptEntry `json:"transcript"`
}

// SessionInitializationRequest represents the request to initialize with lesson data
type SessionInitializationRequest struct {
	Lesson       LessonObject                `json:"lesson"`
	Introduction IntroductionObject          `json:"introduction"`
	Questions    []QuestionObject            `json:"questions"`
	GuideSteps   map[string]GuideStepsObject `json:"guide_steps"`
	Conclusion   ConclusionObject            `json:"conclusion"`
	Persona      PersonaObject               `json:"persona"`
	SampleRate   int                         `json:"sample_rate,omitempty"`
	Encoding     string                      `json:"encoding,omitempty"`
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
		Transcript:   make([]TranscriptEntry, 0),
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

// InitializeSessionWithLesson creates a new interview session with lesson data
func (im *InterviewManager) InitializeSessionWithLesson(w http.ResponseWriter, r *http.Request) {
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

	var req SessionInitializationRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		log.Printf("[ERROR] Failed to decode session initialization request: %v", err)
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	// Validate required lesson data
	if req.Lesson.LessonID == "" {
		http.Error(w, "lesson_id is required", http.StatusBadRequest)
		return
	}

	// Use defaults for audio config if not provided
	if req.SampleRate == 0 {
		req.SampleRate = 16000
	}
	if req.Encoding == "" {
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

	// Initialize session state with defaults
	sessionState := &SessionStateObject{
		CurrentQuestion:   0,
		SilenceTimer:      0,
		HintsUsed:         0,
		ComponentsHit:     make([]string, 0),
		StepsHit:          make([]string, 0),
		FollowUpsUsed:     make([]int, 0),
		UserReady:         false,
		UserReadyQuestion: "Are you ready to move on to the next question?",
		Completed:         false,
	}

	// Convert questions to pointers
	questions := make([]*QuestionObject, len(req.Questions))
	for i := range req.Questions {
		questions[i] = &req.Questions[i]
	}

	// Convert guide steps to pointer map
	guideStepsMap := make(map[string]*GuideStepsObject)
	for key, steps := range req.GuideSteps {
		stepsCopy := steps
		guideStepsMap[key] = &stepsCopy
	}

	// Create session with lesson data
	session := &InterviewSession{
		ID:           sessionID,
		StartTime:    time.Now(),
		Status:       "initialized",
		StreamingSTT: stt.NewStreamingSTT(config),
		VAD:          vad.NewVAD(),
		SessionState: sessionstate.NewSessionState(sessionID),
		ctx:          ctx,
		cancel:       cancel,

		// Lesson data
		Lesson:                &req.Lesson,
		Introduction:          &req.Introduction,
		Questions:             questions,
		GuideStepsMap:         guideStepsMap,
		Conclusion:            &req.Conclusion,
		Persona:               &req.Persona,
		InterviewSessionState: sessionState,

		// Initialize transcript
		Transcript: make([]TranscriptEntry, 0),
	}

	// Store session
	im.sessions[sessionID] = session
	im.mu.Unlock()

	log.Printf("[INFO] Interview session with lesson initialized: %s (lesson: %s)", sessionID, req.Lesson.LessonID)

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
		log.Printf("[ERROR] Session not found: %s", sessionID)
		http.Error(w, "Session not found", http.StatusNotFound)
		return
	}

	// Acquire session lock for state check and update
	session.mu.Lock()

	// Check session state and handle accordingly
	switch session.Status {
	case "disconnected":
		// Allow reconnection of disconnected sessions
		log.Printf("[INFO] Reconnecting disconnected session: %s", sessionID)
	case "connected":
		// Reject if already connected
		session.mu.Unlock()
		log.Printf("[WARN] Rejecting duplicate connection for session: %s", sessionID)
		http.Error(w, "Session already connected", http.StatusConflict)
		return
	case "initialized":
		// First connection, proceed normally
		log.Printf("[INFO] First connection for session: %s", sessionID)
	default:
		// Invalid state
		session.mu.Unlock()
		log.Printf("[ERROR] Invalid session state '%s' for session: %s", session.Status, sessionID)
		http.Error(w, "Invalid session state", http.StatusBadRequest)
		return
	}

	// Close any existing connection before establishing new one
	if session.WebSocketConn != nil {
		log.Printf("[INFO] Closing existing connection for session: %s", sessionID)
		session.WebSocketConn.Close()
		session.WebSocketConn = nil
	}

	// Reset session state if needed
	if session.Status == "disconnected" {
		if session.StreamingSTT != nil {
			session.StreamingSTT.Close()
		}
		// Create new context for the session
		session.ctx, session.cancel = context.WithCancel(context.Background())
		session.StreamingSTT = stt.NewStreamingSTT(session.StreamingSTT.GetConfig())
	}

	// Upgrade connection to WebSocket
	conn, err := im.upgrader.Upgrade(w, r, nil)
	if err != nil {
		session.mu.Unlock()
		log.Printf("[ERROR] WebSocket upgrade failed for session %s: %v", sessionID, err)
		return
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

	// Start listening for transcripts - this will restart automatically after reconnections
	go im.listenForTranscripts(session)
}

// listenForTranscripts handles transcript listening with automatic restart capability
func (im *InterviewManager) listenForTranscripts(session *InterviewSession) {
	for {
		select {
		case <-session.ctx.Done():
			return
		default:
			// Check if StreamingSTT is available
			if session.StreamingSTT == nil {
				time.Sleep(100 * time.Millisecond)
				continue
			}

			// Listen for transcripts
			select {
			case transcript, ok := <-session.StreamingSTT.GetTranscripts():
				if !ok {
					// Channel closed, wait for reconnection
					time.Sleep(100 * time.Millisecond)
					continue
				}
				im.handleTranscript(session, transcript)
			case err, ok := <-session.StreamingSTT.GetErrors():
				if !ok {
					// Channel closed, wait for reconnection
					time.Sleep(100 * time.Millisecond)
					continue
				}
				if err != nil {
					// Check if it's a connection lost error during reconnection
					if err.Error() == "connection lost" {
						log.Printf("[DEBUG] STT connection lost for session %s (expected during reconnection)", session.ID)
						continue
					}
					log.Printf("[ERROR] Streaming STT error for session %s: %v", session.ID, err)
				}
			case <-session.ctx.Done():
				return
			}
		}
	}
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

	// Store transcript entry in ephemeral storage
	transcriptEntry := TranscriptEntry{
		Timestamp:  time.Now(),
		Type:       getTranscriptType(result.MessageType),
		Text:       result.Text,
		Confidence: result.Confidence,
		SessionID:  result.SessionID,
	}
	session.Transcript = append(session.Transcript, transcriptEntry)

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

			// TODO: Trigger context brain analysis for utterance
			im.analyzeUtterance(session, result.Text)
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

// getTranscriptType converts STT message type to transcript type
func getTranscriptType(messageType string) string {
	switch messageType {
	case "PartialTranscript":
		return "partial"
	case "FinalTranscript":
		return "final"
	case "Turn":
		return "utterance"
	default:
		return "unknown"
	}
}

// analyzeUtterance analyzes the utterance against lesson context (placeholder for context brain)
func (im *InterviewManager) analyzeUtterance(session *InterviewSession, utteranceText string) {
	// TODO: Implement context brain analysis
	// This will analyze the utterance against:
	// - Current question's expected components
	// - Guide steps for structural analysis
	// - Session state for progress tracking
	// - Persona for appropriate response generation

	log.Printf("[INFO] Context brain analysis needed for session %s, utterance: %s",
		session.ID, utteranceText)
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
		if session.StreamingSTT != nil {
			session.StreamingSTT.Close()
		}
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
		reconnectDelay          = 1000 * time.Millisecond // Delay before reconnecting (for errors only)
	)

	inUtterance := false
	utteranceStartTime := time.Time{}
	reconnectAttempts := 0
	const maxReconnectAttempts = 3

	// Only reconnect on actual connection errors, not for end of utterance
	reconnectSTTOnError := func() error {
		if reconnectAttempts >= maxReconnectAttempts {
			return fmt.Errorf("exceeded maximum reconnection attempts")
		}
		reconnectAttempts++

		log.Printf("[ERROR] STT connection error, initiating reconnection for session %s (attempt %d)",
			session.ID, reconnectAttempts)

		// Wait before reconnecting to allow cleanup
		time.Sleep(reconnectDelay)

		// Close existing connection gracefully and wait for it to fully close
		if session.StreamingSTT != nil {
			log.Printf("[INFO] Closing existing STT connection for session %s", session.ID)
			if err := session.StreamingSTT.Close(); err != nil {
				log.Printf("[WARN] Error closing STT connection: %v", err)
			}
			// Give time for the connection to fully close
			time.Sleep(500 * time.Millisecond)
		}

		// Create completely new STT instance with fresh configuration
		config := stt.StreamingConfig{
			SampleRate:                       16000,
			Encoding:                         "pcm_s16le",
			FormatTurns:                      true,
			EndOfTurnConfidenceThreshold:     0.7,
			MinEndOfTurnSilenceWhenConfident: 1000,
			MaxTurnSilence:                   3000,
		}
		session.StreamingSTT = stt.NewStreamingSTT(config)

		// Connect with retry logic
		var connectErr error
		for i := 0; i < 3; i++ {
			connectErr = session.StreamingSTT.Connect(session.ctx)
			if connectErr == nil {
				break
			}
			log.Printf("[WARN] STT connection attempt %d failed for session %s: %v",
				i+1, session.ID, connectErr)
			if i < 2 {
				time.Sleep(500 * time.Millisecond)
			}
		}

		if connectErr != nil {
			log.Printf("[ERROR] Failed to reconnect STT for session %s after %d attempts: %v",
				session.ID, reconnectAttempts, connectErr)
			return connectErr
		}

		log.Printf("[INFO] Successfully reconnected STT for session %s (attempt %d)",
			session.ID, reconnectAttempts)

		// Reset AssemblyAI session ID since we have a new connection
		session.mu.Lock()
		session.AssemblyAIID = ""
		session.mu.Unlock()

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

							// TODO: Trigger response generation here
							// Keep STT connection alive for continued listening
							log.Printf("[INFO] End of utterance detected for session %s - ready for response generation", session.ID)
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

				// TODO: Trigger response generation here
				// Keep STT connection alive for continued listening
				log.Printf("[INFO] Max duration utterance ended for session %s - ready for response generation", session.ID)
				continue
			}

			// Send to streaming STT if voice detected
			if hasVoice {
				if session.StreamingSTT == nil {
					log.Printf("[WARN] StreamingSTT is nil, skipping audio data")
					continue
				}

				err = session.StreamingSTT.SendAudio(audioData)
				if err != nil {
					log.Printf("[ERROR] Error sending audio to STT: %v", err)

					// Attempt to reconnect only on actual connection errors
					if err := reconnectSTTOnError(); err != nil {
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
