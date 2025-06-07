# Call Service Frontend Implementation Guide

## Overview
This guide explains how to implement a frontend client for the Call Service API, which provides real-time audio streaming and transcription capabilities using WebSocket connections.

## API Endpoints

### Base URL
```
http://localhost:8080
```

### Available Endpoints

1. **Initialize Interview Session**
   - Method: `POST`
   - Endpoint: `/api/interview/init`
   - Request Body:
     ```json
     {
       "sample_rate": 16000,
       "encoding": "pcm_s16le"
     }
     ```
   - Response:
     ```json
     {
       "session_id": "uuid-string",
       "websocket_url": "ws://localhost:8080/ws/interview/uuid-string",
       "status": "initialized"
     }
     ```

2. **Get Session Status**
   - Method: `GET`
   - Endpoint: `/api/interview/status?session_id={session_id}`
   - Response:
     ```json
     {
       "session_id": "uuid-string",
       "status": "connected",
       "start_time": "2024-01-01T00:00:00Z",
       "transcript_count": 10,
       "utterance_count": 5
     }
     ```

3. **Close Session**
   - Method: `DELETE`
   - Endpoint: `/api/interview/close?session_id={session_id}`
   - Response:
     ```json
     {
       "status": "closed",
       "session_id": "uuid-string"
     }
     ```

4. **WebSocket Audio Streaming**
   - Endpoint: `ws://localhost:8080/ws/interview/{session_id}`

## Implementation Guide

### 1. Setting Up Audio Recording

```javascript
let mediaRecorder;
let websocket;
const sampleRate = 16000;

async function setupAudioRecording() {
  try {
    const stream = await navigator.mediaDevices.getUserMedia({ audio: true });
    
    // Configure audio context for correct sample rate
    const audioContext = new AudioContext({ sampleRate });
    const source = audioContext.createMediaStreamSource(stream);
    const processor = audioContext.createScriptProcessor(4096, 1, 1);
    
    source.connect(processor);
    processor.connect(audioContext.destination);
    
    processor.onaudioprocess = (e) => {
      if (websocket && websocket.readyState === WebSocket.OPEN) {
        // Convert Float32Array to Int16Array
        const float32Array = e.inputBuffer.getChannelData(0);
        const int16Array = new Int16Array(float32Array.length);
        for (let i = 0; i < float32Array.length; i++) {
          int16Array[i] = float32Array[i] * 32767;
        }
        
        // Send audio data through WebSocket
        websocket.send(int16Array.buffer);
      }
    };
  } catch (error) {
    console.error('Error accessing microphone:', error);
  }
}
```

### 2. Session Management

```javascript
async function initializeSession() {
  try {
    const response = await fetch('http://localhost:8080/api/interview/init', {
      method: 'POST',
      headers: {
        'Content-Type': 'application/json',
      },
      body: JSON.stringify({
        sample_rate: 16000,
        encoding: 'pcm_s16le'
      })
    });
    
    const data = await response.json();
    return data;
  } catch (error) {
    console.error('Error initializing session:', error);
  }
}

async function closeSession(sessionId) {
  try {
    const response = await fetch(`http://localhost:8080/api/interview/close?session_id=${sessionId}`, {
      method: 'DELETE'
    });
    
    const data = await response.json();
    return data;
  } catch (error) {
    console.error('Error closing session:', error);
  }
}
```

### 3. WebSocket Connection

```javascript
// Add reconnection configuration
const WEBSOCKET_RECONNECT_DELAY = 2000; // 2 seconds
const MAX_RECONNECT_ATTEMPTS = 5;
let reconnectAttempts = 0;
let isReconnecting = false;

function connectWebSocket(websocketUrl) {
  websocket = new WebSocket(websocketUrl);
  
  websocket.onopen = () => {
    console.log('WebSocket connected');
    reconnectAttempts = 0;
    isReconnecting = false;
    setupAudioRecording();
  };
  
  websocket.onmessage = (event) => {
    const data = JSON.parse(event.data);
    
    switch (data.type) {
      case 'transcript':
        handleTranscript(data);
        break;
      case 'error':
        handleError(data);
        break;
    }
  };
  
  websocket.onerror = (error) => {
    console.error('WebSocket error:', error);
    handleWebSocketError(error);
  };
  
  websocket.onclose = (event) => {
    console.log('WebSocket closed:', event.code, event.reason);
    handleWebSocketClose(event, websocketUrl);
  };
}

function handleWebSocketError(error) {
  // Log the error details
  console.error('WebSocket error details:', {
    timestamp: new Date().toISOString(),
    error: error.message || 'Unknown error'
  });
  
  // Notify the user
  const statusDiv = document.getElementById('status');
  if (statusDiv) {
    statusDiv.innerHTML = `<div class="error">Connection error. Attempting to reconnect...</div>`;
  }
}

function handleWebSocketClose(event, websocketUrl) {
  if (!isReconnecting && reconnectAttempts < MAX_RECONNECT_ATTEMPTS) {
    isReconnecting = true;
    reconnectAttempts++;
    
    console.log(`Attempting to reconnect (${reconnectAttempts}/${MAX_RECONNECT_ATTEMPTS})...`);
    
    // Update UI
    const statusDiv = document.getElementById('status');
    if (statusDiv) {
      statusDiv.innerHTML = `<div class="warning">Connection lost. Reconnecting... (Attempt ${reconnectAttempts}/${MAX_RECONNECT_ATTEMPTS})</div>`;
    }
    
    // Attempt to reconnect after delay
    setTimeout(() => {
      if (reconnectAttempts < MAX_RECONNECT_ATTEMPTS) {
        connectWebSocket(websocketUrl);
      } else {
        console.error('Max reconnection attempts reached');
        const statusDiv = document.getElementById('status');
        if (statusDiv) {
          statusDiv.innerHTML = `<div class="error">Connection failed after ${MAX_RECONNECT_ATTEMPTS} attempts. Please refresh the page.</div>`;
        }
      }
    }, WEBSOCKET_RECONNECT_DELAY);
  }
}

function handleError(data) {
  console.error('Server error:', data);
  // Handle specific error types
  switch (data.error_type) {
    case 'STT_ERROR':
      handleSTTError(data);
      break;
    case 'AUDIO_ERROR':
      handleAudioError(data);
      break;
    default:
      console.error('Unknown error type:', data);
  }
}

function handleSTTError(data) {
  const statusDiv = document.getElementById('status');
  if (statusDiv) {
    statusDiv.innerHTML = `<div class="error">Speech-to-text error: ${data.message}</div>`;
  }
}

function handleAudioError(data) {
  const statusDiv = document.getElementById('status');
  if (statusDiv) {
    statusDiv.innerHTML = `<div class="error">Audio error: ${data.message}</div>`;
  }
}
```

### 4. Complete Implementation Example

```javascript
async function startInterview() {
  // Initialize session
  const sessionData = await initializeSession();
  
  if (sessionData) {
    // Connect WebSocket
    connectWebSocket(sessionData.websocket_url);
    
    // Start periodic status checks
    const statusInterval = setInterval(async () => {
      const status = await fetch(`http://localhost:8080/api/interview/status?session_id=${sessionData.session_id}`);
      const statusData = await status.json();
      console.log('Session status:', statusData);
    }, 5000);
    
    // Cleanup function
    return async () => {
      clearInterval(statusInterval);
      if (websocket) {
        websocket.close();
      }
      await closeSession(sessionData.session_id);
    };
  }
}
```

## Important Notes

1. **Audio Format Requirements**
   - Sample Rate: 16000 Hz
   - Encoding: PCM 16-bit Little Endian
   - Single Channel (Mono)

2. **WebSocket Messages**
   The server sends different types of transcript messages:
   - `PartialTranscript`: Real-time updates as speech is detected
   - `FinalTranscript`: Completed phrases with higher confidence
   - `Turn`: Complete utterances when the speaker stops talking

3. **Error Handling**
   - Implement proper error handling for all API calls
   - Handle WebSocket disconnections and reconnection logic
   - Monitor session status periodically

4. **Browser Compatibility**
   - Ensure the browser supports:
     - WebSocket API
     - getUserMedia API
     - AudioContext API
     - Int16Array and Float32Array

## Example UI Implementation

```html
<!DOCTYPE html>
<html>
<head>
    <title>Call Service Client</title>
    <style>
        .transcript {
            margin: 20px;
            padding: 10px;
            border: 1px solid #ccc;
        }
        .partial { color: gray; }
        .final { color: black; }
        .utterance { color: blue; }
        .error {
            color: #721c24;
            background-color: #f8d7da;
            border: 1px solid #f5c6cb;
            padding: 10px;
            margin: 10px 0;
            border-radius: 4px;
        }
        .warning {
            color: #856404;
            background-color: #fff3cd;
            border: 1px solid #ffeeba;
            padding: 10px;
            margin: 10px 0;
            border-radius: 4px;
        }
        #status {
            margin: 20px 0;
        }
    </style>
</head>
<body>
    <div>
        <button id="startBtn">Start Interview</button>
        <button id="stopBtn" disabled>Stop Interview</button>
    </div>
    <div id="status"></div>
    <div id="transcripts"></div>
    
    <script>
        // Include the JavaScript implementation from above
        
        document.getElementById('startBtn').onclick = async () => {
            try {
                const cleanup = await startInterview();
                document.getElementById('startBtn').disabled = true;
                document.getElementById('stopBtn').disabled = false;
                
                document.getElementById('stopBtn').onclick = async () => {
                    try {
                        await cleanup();
                        document.getElementById('startBtn').disabled = false;
                        document.getElementById('stopBtn').disabled = true;
                        document.getElementById('status').innerHTML = '<div>Session ended</div>';
                    } catch (error) {
                        console.error('Error during cleanup:', error);
                        document.getElementById('status').innerHTML = '<div class="error">Error ending session</div>';
                    }
                };
            } catch (error) {
                console.error('Error starting interview:', error);
                document.getElementById('status').innerHTML = '<div class="error">Failed to start interview</div>';
            }
        };
    </script>
</body>
</html>
```

## Security Considerations

1. In production:
   - Use HTTPS for API endpoints
   - Use WSS (WebSocket Secure) for WebSocket connections
   - Implement proper authentication and authorization
   - Add CORS restrictions
   - Rate limit API endpoints

2. Handle sensitive data appropriately:
   - Don't log sensitive transcripts
   - Implement data retention policies
   - Follow relevant privacy regulations

## Troubleshooting

Common issues and solutions:

1. **Audio not streaming**
   - Check microphone permissions
   - Verify audio format settings
   - Check WebSocket connection status

2. **High latency**
   - Reduce audio buffer size
   - Check network connection
   - Monitor server load

3. **WebSocket disconnections**
   - Implement reconnection logic
   - Check network stability
   - Monitor server logs

4. **Poor transcription quality**
   - Verify audio sample rate
   - Check microphone quality
   - Reduce background noise

## Best Practices for Connection Handling

1. **WebSocket Reconnection Strategy**
   - Implement exponential backoff for reconnection attempts
   - Set a maximum number of reconnection attempts
   - Provide clear feedback to users during reconnection
   - Reset reconnection counters on successful connection

2. **Error Recovery**
   - Cache unsent audio data during connection interruptions
   - Resume from last known good state after reconnection
   - Implement session recovery mechanism
   - Provide manual refresh option when automatic recovery fails

3. **User Feedback**
   - Display connection status clearly
   - Show meaningful error messages
   - Indicate reconnection progress
   - Provide actionable instructions when intervention is needed

4. **Session Management**
   - Track session state during disconnections
   - Implement session timeout handling
   - Clean up resources properly on session end
   - Handle multiple connection scenarios 