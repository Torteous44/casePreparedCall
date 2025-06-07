package stt

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"os"

	"github.com/AssemblyAI/assemblyai-go-sdk"
)

// STT represents a Speech-to-Text service using AssemblyAI
type STT struct {
	client *assemblyai.Client
}

// NewSTT creates a new Speech-to-Text service using AssemblyAI
func NewSTT() *STT {
	apiKey := os.Getenv("ASSEMBLYAI_API_KEY")
	if apiKey == "" {
		panic("ASSEMBLYAI_API_KEY environment variable is not set")
	}

	client := assemblyai.NewClient(apiKey)
	return &STT{
		client: client,
	}
}

// Transcribe converts audio data to text using AssemblyAI
func (s *STT) Transcribe(audioData []byte) (string, error) {
	// Create a reader from the audio data
	audioReader := bytes.NewReader(audioData)

	// Use TranscribeFromReader method
	transcript, err := s.client.Transcripts.TranscribeFromReader(context.Background(), audioReader, nil)
	if err != nil {
		return "", fmt.Errorf("failed to transcribe audio: %w", err)
	}

	if transcript.Text == nil {
		return "", fmt.Errorf("transcription completed but no text was returned")
	}

	return *transcript.Text, nil
}

// TranscribeFile transcribes audio from a file using AssemblyAI
func (s *STT) TranscribeFile(filePath string) (string, error) {
	file, err := os.Open(filePath)
	if err != nil {
		return "", fmt.Errorf("failed to open audio file: %w", err)
	}
	defer file.Close()

	// Use TranscribeFromReader method
	transcript, err := s.client.Transcripts.TranscribeFromReader(context.Background(), file, nil)
	if err != nil {
		return "", fmt.Errorf("failed to transcribe audio file: %w", err)
	}

	if transcript.Text == nil {
		return "", fmt.Errorf("transcription completed but no text was returned")
	}

	return *transcript.Text, nil
}

// TranscribeFromURL transcribes audio from a URL using AssemblyAI
func (s *STT) TranscribeFromURL(audioURL string) (string, error) {
	// Use TranscribeFromURL method
	transcript, err := s.client.Transcripts.TranscribeFromURL(context.Background(), audioURL, nil)
	if err != nil {
		return "", fmt.Errorf("failed to transcribe audio from URL: %w", err)
	}

	if transcript.Text == nil {
		return "", fmt.Errorf("transcription completed but no text was returned")
	}

	return *transcript.Text, nil
}

// TranscribeStream transcribes audio from an io.Reader using AssemblyAI
func (s *STT) TranscribeStream(reader io.Reader) (string, error) {
	// Use TranscribeFromReader method
	transcript, err := s.client.Transcripts.TranscribeFromReader(context.Background(), reader, nil)
	if err != nil {
		return "", fmt.Errorf("failed to transcribe audio stream: %w", err)
	}

	if transcript.Text == nil {
		return "", fmt.Errorf("transcription completed but no text was returned")
	}

	return *transcript.Text, nil
}

// TranscribeWithOptions transcribes audio with custom options
func (s *STT) TranscribeWithOptions(audioData []byte, opts *assemblyai.TranscriptOptionalParams) (string, error) {
	audioReader := bytes.NewReader(audioData)

	transcript, err := s.client.Transcripts.TranscribeFromReader(context.Background(), audioReader, opts)
	if err != nil {
		return "", fmt.Errorf("failed to transcribe audio with options: %w", err)
	}

	if transcript.Text == nil {
		return "", fmt.Errorf("transcription completed but no text was returned")
	}

	return *transcript.Text, nil
}

// NewStreamingSTTWithDefaults creates a new streaming STT instance with default configuration
func (s *STT) NewStreamingSTTWithDefaults() *StreamingSTT {
	return NewStreamingSTT(GetDefaultStreamingConfig())
}

// NewStreamingSTTWithConfig creates a new streaming STT instance with custom configuration
func (s *STT) NewStreamingSTTWithConfig(config StreamingConfig) *StreamingSTT {
	return NewStreamingSTT(config)
}
