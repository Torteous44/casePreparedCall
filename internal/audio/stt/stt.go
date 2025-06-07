package stt

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"os"

	"github.com/sashabaranov/go-openai"
)

// STT represents a Speech-to-Text service using OpenAI Whisper
type STT struct {
	client *openai.Client
}

// NewSTT creates a new Speech-to-Text service
func NewSTT() *STT {
	apiKey := os.Getenv("OPENAI_API_KEY")
	if apiKey == "" {
		panic("OPENAI_API_KEY environment variable is not set")
	}

	client := openai.NewClient(apiKey)
	return &STT{
		client: client,
	}
}

// Transcribe converts audio data to text using OpenAI Whisper
func (s *STT) Transcribe(audioData []byte) (string, error) {
	// Create a reader from the audio data
	audioReader := bytes.NewReader(audioData)

	// Create transcription request
	req := openai.AudioRequest{
		Model:  openai.Whisper1,
		Reader: audioReader,
		Format: openai.AudioResponseFormatText,
	}

	// Call OpenAI Whisper API
	resp, err := s.client.CreateTranscription(context.Background(), req)
	if err != nil {
		return "", fmt.Errorf("failed to transcribe audio: %w", err)
	}

	return resp.Text, nil
}

// TranscribeFile transcribes audio from a file
func (s *STT) TranscribeFile(filePath string) (string, error) {
	file, err := os.Open(filePath)
	if err != nil {
		return "", fmt.Errorf("failed to open audio file: %w", err)
	}
	defer file.Close()

	req := openai.AudioRequest{
		Model:  openai.Whisper1,
		Reader: file,
		Format: openai.AudioResponseFormatText,
	}

	resp, err := s.client.CreateTranscription(context.Background(), req)
	if err != nil {
		return "", fmt.Errorf("failed to transcribe audio file: %w", err)
	}

	return resp.Text, nil
}

// TranscribeStream transcribes audio from an io.Reader
func (s *STT) TranscribeStream(reader io.Reader) (string, error) {
	req := openai.AudioRequest{
		Model:  openai.Whisper1,
		Reader: reader,
		Format: openai.AudioResponseFormatText,
	}

	resp, err := s.client.CreateTranscription(context.Background(), req)
	if err != nil {
		return "", fmt.Errorf("failed to transcribe audio stream: %w", err)
	}

	return resp.Text, nil
}
