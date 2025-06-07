package main

import (
	"fmt"
	"log"
	"net/http"

	"github.com/joho/godotenv"
	"github.com/torteous44/callservice/internal/orchestrator"
)

func main() {
	// Load environment variables from .env file
	err := godotenv.Load()
	if err != nil {
		log.Printf("Warning: Error loading .env file: %v", err)
	}

	fmt.Println("🚀 Call Service starting...")
	log.Println("Call Service initialized")

	// Create interview manager
	interviewManager := orchestrator.NewInterviewManager()

	// Set up HTTP routes
	http.HandleFunc("/api/interview/init", interviewManager.InitializeSession)
	http.HandleFunc("/api/interview/init-with-lesson", interviewManager.InitializeSessionWithLesson)
	http.HandleFunc("/api/interview/status", interviewManager.GetSessionStatus)
	http.HandleFunc("/api/interview/close", interviewManager.CloseSession)

	// WebSocket endpoint for audio streaming
	http.HandleFunc("/ws/interview/", interviewManager.HandleWebSocket)

	// Health check endpoint
	http.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"status": "healthy", "service": "call-service"}`))
	})

	// Serve static files (optional, for serving a simple test page)
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/" {
			w.Header().Set("Content-Type", "text/html")
			w.Write([]byte(`
<!DOCTYPE html>
<html>
<head>
    <title>Call Service API</title>
</head>
<body>
    <h1> Call Service API</h1>
    <h2>Available Endpoints:</h2>
    <ul>
        <li><strong>POST /api/interview/init</strong> - Initialize a new interview session</li>
        <li><strong>GET /api/interview/status?session_id=xxx</strong> - Get session status</li>
        <li><strong>DELETE /api/interview/close?session_id=xxx</strong> - Close session</li>
        <li><strong>WebSocket /ws/interview/{session_id}</strong> - Audio streaming</li>
        <li><strong>GET /health</strong> - Health check</li>
    </ul>
    
    <h2>Example Usage:</h2>
    <pre>
// Initialize session
fetch('/api/interview/init', { method: 'POST' })
  .then(r => r.json())
  .then(data => {
    console.log('Session ID:', data.session_id);
    console.log('WebSocket URL:', data.websocket_url);
  });
    </pre>
</body>
</html>
			`))
		} else {
			http.NotFound(w, r)
		}
	})

	port := "8080"
	log.Printf("🌐 Server starting on http://localhost:%s", port)
	log.Printf("WebSocket endpoint: ws://localhost:%s/ws/interview/{session_id}", port)
	log.Printf("Interview API: http://localhost:%s/api/interview/", port)

	fmt.Println("\nLegend:")
	fmt.Println("   Voice detected - when VAD detects speech")
	fmt.Println("   Silence detected - when VAD detects silence")
	fmt.Println("   🔄 Partial transcripts - real-time speech recognition")
	fmt.Println("   ✅ Final transcripts - completed phrases")
	fmt.Println("   🗣️  End of utterance - when speaker stops talking")
	fmt.Println()

	if err := http.ListenAndServe(":"+port, nil); err != nil {
		log.Fatal("❌ Server failed to start:", err)
	}
}
