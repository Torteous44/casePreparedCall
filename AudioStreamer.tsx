import React, { useEffect, useState, useCallback, useRef } from "react";

// Configuration
const API_BASE_URL = "http://localhost:8080";
const SAMPLE_RATE = 16000;

interface TranscriptMessage {
  type: string;
  message_type: "PartialTranscript" | "FinalTranscript" | "Turn";
  text: string;
}

interface SessionResponse {
  session_id: string;
  websocket_url: string;
  status: string;
}

interface AudioStreamerProps {
  onStatusChange?: (status: string) => void;
  onTranscript?: (text: string, type: string) => void;
}

const AudioStreamer: React.FC<AudioStreamerProps> = ({
  onStatusChange,
  onTranscript,
}) => {
  const [status, setStatus] = useState<string>("Initializing...");
  const [error, setError] = useState<string | null>(null);
  const [sessionId, setSessionId] = useState<string | null>(null);

  // Refs to track connection state and cleanup
  const websocketRef = useRef<WebSocket | null>(null);
  const cleanupRef = useRef<(() => void) | null>(null);
  const statusIntervalRef = useRef<NodeJS.Timeout | null>(null);
  const isConnectingRef = useRef<boolean>(false);

  const updateStatus = useCallback(
    (newStatus: string) => {
      setStatus(newStatus);
      onStatusChange?.(newStatus);
    },
    [onStatusChange]
  );

  const cleanupSession = useCallback(async () => {
    // Clear all refs and intervals
    if (cleanupRef.current) {
      cleanupRef.current();
      cleanupRef.current = null;
    }

    if (statusIntervalRef.current) {
      clearInterval(statusIntervalRef.current);
      statusIntervalRef.current = null;
    }

    if (websocketRef.current) {
      websocketRef.current.close();
      websocketRef.current = null;
    }

    // Close session with backend
    if (sessionId) {
      try {
        await fetch(
          `${API_BASE_URL}/api/interview/close?session_id=${sessionId}`,
          {
            method: "DELETE",
          }
        );
        setSessionId(null);
      } catch (err) {
        console.error("Error closing session:", err);
      }
    }
  }, [sessionId]);

  const initializeSession = async (): Promise<SessionResponse | null> => {
    if (isConnectingRef.current) {
      console.warn("Session initialization already in progress");
      return null;
    }

    try {
      isConnectingRef.current = true;
      const response = await fetch(`${API_BASE_URL}/api/interview/init`, {
        method: "POST",
        headers: {
          "Content-Type": "application/json",
        },
        body: JSON.stringify({
          sample_rate: SAMPLE_RATE,
          encoding: "pcm_s16le",
        }),
      });

      if (!response.ok) {
        throw new Error(`Failed to initialize session: ${response.statusText}`);
      }

      const data = await response.json();
      setSessionId(data.session_id);
      return data;
    } catch (err) {
      console.error("Error initializing session:", err);
      setError(
        err instanceof Error ? err.message : "Failed to initialize session"
      );
      return null;
    } finally {
      isConnectingRef.current = false;
    }
  };

  const setupAudioProcessing = async (websocket: WebSocket) => {
    try {
      const stream = await navigator.mediaDevices.getUserMedia({ audio: true });
      updateStatus("Microphone access granted");

      // Configure audio context for correct sample rate
      const audioContext = new AudioContext({ sampleRate: SAMPLE_RATE });
      const source = audioContext.createMediaStreamSource(stream);
      const processor = audioContext.createScriptProcessor(4096, 1, 1);

      source.connect(processor);
      processor.connect(audioContext.destination);

      processor.onaudioprocess = (e) => {
        if (websocket.readyState === WebSocket.OPEN) {
          const float32Array = e.inputBuffer.getChannelData(0);
          const int16Array = new Int16Array(float32Array.length);

          for (let i = 0; i < float32Array.length; i++) {
            const s = Math.max(-1, Math.min(1, float32Array[i]));
            int16Array[i] = s < 0 ? s * 0x8000 : s * 0x7fff;
          }

          websocket.send(int16Array.buffer);
        }
      };

      return () => {
        processor.disconnect();
        source.disconnect();
        audioContext.close();
        stream.getTracks().forEach((track) => track.stop());
      };
    } catch (err) {
      console.error("Error setting up audio:", err);
      setError(
        err instanceof Error ? err.message : "Failed to setup audio stream"
      );
      return () => {};
    }
  };

  const handleTranscriptMessage = (data: TranscriptMessage) => {
    switch (data.message_type) {
      case "PartialTranscript":
        onTranscript?.(data.text, "partial");
        break;
      case "FinalTranscript":
        onTranscript?.(data.text, "final");
        break;
      case "Turn":
        onTranscript?.(data.text, "utterance");
        break;
    }
  };

  useEffect(() => {
    const setupConnection = async () => {
      // Clean up any existing session first
      await cleanupSession();

      const sessionData = await initializeSession();
      if (!sessionData) return;

      const websocket = new WebSocket(sessionData.websocket_url);
      websocketRef.current = websocket;

      websocket.onopen = async () => {
        updateStatus("Connected");
        cleanupRef.current = await setupAudioProcessing(websocket);
      };

      websocket.onmessage = (event) => {
        try {
          const data = JSON.parse(event.data);
          if (data.type === "transcript") {
            handleTranscriptMessage(data);
          }
        } catch (err) {
          console.error("Error processing message:", err);
        }
      };

      websocket.onerror = (error) => {
        console.error("WebSocket error:", error);
        updateStatus("Error");
        setError("WebSocket connection error");
      };

      websocket.onclose = () => {
        updateStatus("Disconnected");
      };

      // Start periodic status checks
      statusIntervalRef.current = setInterval(async () => {
        if (!sessionData.session_id) return;

        try {
          const response = await fetch(
            `${API_BASE_URL}/api/interview/status?session_id=${sessionData.session_id}`
          );
          if (!response.ok) {
            throw new Error(`Status check failed: ${response.statusText}`);
          }
          const statusData = await response.json();
          console.log("Session status:", statusData);
        } catch (err) {
          console.error("Error checking status:", err);
        }
      }, 5000);
    };

    setupConnection();

    // Cleanup function
    return () => {
      cleanupSession();
    };
  }, []); // Empty dependency array - only run once on mount

  return (
    <div className="audio-streamer">
      <div className="status">Status: {status}</div>
      {error && <div className="error">Error: {error}</div>}
    </div>
  );
};

export default AudioStreamer;
