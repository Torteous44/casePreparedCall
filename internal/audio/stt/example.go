package stt

import (
	"context"
	"fmt"
	"log"
	"time"
)

// StreamingExample demonstrates how to use the streaming STT functionality
func StreamingExample() {
	// Create a streaming STT instance with default configuration
	config := GetDefaultStreamingConfig()
	streamSTT := NewStreamingSTT(config)

	// Create context with timeout
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
	defer cancel()

	// Connect to AssemblyAI streaming API
	err := streamSTT.Connect(ctx)
	if err != nil {
		log.Fatalf("Failed to connect: %v", err)
	}
	defer streamSTT.Close()

	// Start listening for transcripts and errors
	go func() {
		for {
			select {
			case transcript, ok := <-streamSTT.GetTranscripts():
				if !ok {
					return
				}
				handleTranscript(transcript)
			case err, ok := <-streamSTT.GetErrors():
				if !ok {
					return
				}
				log.Printf("Streaming error: %v", err)
			case <-ctx.Done():
				return
			}
		}
	}()

	// Simulate sending audio chunks
	// In a real application, you would get audio from a microphone or audio stream
	audioChunks := [][]byte{
		// These would be actual audio data chunks (50ms to 1000ms each)
		// For demonstration, we're using empty byte slices
		make([]byte, 1600), // ~100ms of 16kHz 16-bit audio
		make([]byte, 1600),
		make([]byte, 1600),
	}

	for _, chunk := range audioChunks {
		err := streamSTT.SendAudio(chunk)
		if err != nil {
			log.Printf("Failed to send audio: %v", err)
			break
		}
		time.Sleep(100 * time.Millisecond) // Simulate real-time audio
	}

	// Force an endpoint to complete any remaining transcription
	streamSTT.ForceEndpoint()

	// Wait a bit for final results
	time.Sleep(2 * time.Second)
}

// handleTranscript processes received transcription results
func handleTranscript(result StreamingResult) {
	switch result.MessageType {
	case "PartialTranscript":
		fmt.Printf("Partial: %s (confidence: %.2f)\n", result.Text, result.Confidence)
	case "FinalTranscript":
		fmt.Printf("Final: %s (confidence: %.2f)\n", result.Text, result.Confidence)
	case "Turn":
		fmt.Printf("Turn [%s]: %s (%.2f)\n", result.TurnID, result.Text, result.Confidence)
	default:
		fmt.Printf("Unknown type %s: %s\n", result.MessageType, result.Text)
	}
}

// StreamingWithCustomConfig demonstrates custom configuration
func StreamingWithCustomConfig() {
	// Create custom configuration
	config := StreamingConfig{
		SampleRate:                       48000, // Higher sample rate
		Encoding:                         "pcm_s16le",
		FormatTurns:                      true,
		EndOfTurnConfidenceThreshold:     0.8,  // Higher confidence threshold
		MinEndOfTurnSilenceWhenConfident: 500,  // Shorter silence detection
		MaxTurnSilence:                   2000, // Shorter max silence
	}

	streamSTT := NewStreamingSTT(config)
	ctx := context.Background()

	err := streamSTT.Connect(ctx)
	if err != nil {
		log.Fatalf("Failed to connect: %v", err)
	}
	defer streamSTT.Close()

	// Update configuration during session
	newConfig := config
	newConfig.EndOfTurnConfidenceThreshold = 0.9
	err = streamSTT.UpdateConfig(newConfig)
	if err != nil {
		log.Printf("Failed to update config: %v", err)
	}

	// Continue with audio processing...
}
