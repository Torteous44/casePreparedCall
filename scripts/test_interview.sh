#!/bin/bash

# Test script for interview API endpoints
echo "🧪 Testing Call Service Interview API"
echo "======================================"

BASE_URL="http://localhost:8080"

# Test health endpoint
echo "1. Testing health endpoint..."
curl -s "$BASE_URL/health" | jq . || echo "❌ Health check failed"
echo ""

# Test session initialization
echo "2. Creating new interview session..."
SESSION_RESPONSE=$(curl -s -X POST "$BASE_URL/api/interview/init" \
  -H "Content-Type: application/json" \
  -d '{"sample_rate": 16000, "encoding": "pcm_s16le"}')

echo "Response: $SESSION_RESPONSE"

# Extract session ID
SESSION_ID=$(echo "$SESSION_RESPONSE" | jq -r '.session_id')
echo "Session ID: $SESSION_ID"
echo ""

# Test session status
echo "3. Checking session status..."
curl -s "$BASE_URL/api/interview/status?session_id=$SESSION_ID" | jq . || echo "❌ Status check failed"
echo ""

# Test WebSocket connection
echo "4. Testing WebSocket connection..."
WS_URL="ws://localhost:8080/ws/interview/$SESSION_ID"
echo "Connecting to: $WS_URL"

# Use websocat if available, otherwise provide instructions
if command -v websocat &> /dev/null; then
    echo "Sending test audio data..."
    
    # Create a small test audio file (1 second of silence)
    dd if=/dev/zero of=test_audio.raw bs=32000 count=1 2>/dev/null
    
    # Create a temporary file for the JSON message
    echo "{\"type\":\"audio_data\",\"audio_data\":\"$(base64 test_audio.raw)\"}" > test_message.json
    
    # Send the JSON message through WebSocket
    cat test_message.json | websocat "$WS_URL" &
    WEBSOCAT_PID=$!
    
    # Wait for a few seconds to receive any responses
    echo "Waiting for responses..."
    sleep 3
    
    # Kill the WebSocket connection
    kill $WEBSOCAT_PID 2>/dev/null
    
    # Clean up test files
    rm test_audio.raw test_message.json
else
    echo "⚠️  websocat not installed. To test WebSocket manually:"
    echo "1. Install websocat: brew install websocat"
    echo "2. Connect using: websocat $WS_URL"
    echo "3. Send audio data as JSON: {\"type\":\"audio_data\",\"audio_data\":\"<base64-audio>\"}"
fi

# Wait a moment
echo "5. Waiting 2 seconds..."
sleep 2

