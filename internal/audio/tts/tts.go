package tts

import (
	"context"
	"fmt"
	"io"
	"os"

	"github.com/sashabaranov/go-openai"
)

// TTS represents a Text-to-Speech service using OpenAI
type TTS struct {
	client *openai.Client
}

// NewTTS creates a new Text-to-Speech service using OpenAI
func NewTTS() *TTS {
	apiKey := os.Getenv("OPENAI_API_KEY")
	if apiKey == "" {
		panic("OPENAI_API_KEY environment variable is not set")
	}

	client := openai.NewClient(apiKey)
	return &TTS{
		client: client,
	}
}

// Synthesize converts text to audio data using OpenAI TTS
func (t *TTS) Synthesize(text string) ([]byte, error) {
	return t.SynthesizeWithVoice(text, openai.VoiceAlloy)
}

// SynthesizeWithVoice converts text to audio with a specific voice
func (t *TTS) SynthesizeWithVoice(text string, voice openai.SpeechVoice) ([]byte, error) {
	req := openai.CreateSpeechRequest{
		Model: openai.TTSModel1,
		Input: text,
		Voice: voice,
	}

	response, err := t.client.CreateSpeech(context.Background(), req)
	if err != nil {
		return nil, fmt.Errorf("failed to create speech: %w", err)
	}
	defer response.Close()

	// Read the audio data
	audioData, err := io.ReadAll(response)
	if err != nil {
		return nil, fmt.Errorf("failed to read audio data: %w", err)
	}

	return audioData, nil
}

// SynthesizeWithOptions converts text to audio with custom options
func (t *TTS) SynthesizeWithOptions(text string, opts SynthesizeOptions) ([]byte, error) {
	req := openai.CreateSpeechRequest{
		Model:          opts.Model,
		Input:          text,
		Voice:          opts.Voice,
		ResponseFormat: opts.Format,
		Speed:          opts.Speed,
	}

	response, err := t.client.CreateSpeech(context.Background(), req)
	if err != nil {
		return nil, fmt.Errorf("failed to create speech with options: %w", err)
	}
	defer response.Close()

	// Read the audio data
	audioData, err := io.ReadAll(response)
	if err != nil {
		return nil, fmt.Errorf("failed to read audio data: %w", err)
	}

	return audioData, nil
}

// SynthesizeToFile converts text to speech and saves to a file
func (t *TTS) SynthesizeToFile(text, filePath string) error {
	audioData, err := t.Synthesize(text)
	if err != nil {
		return fmt.Errorf("failed to synthesize speech: %w", err)
	}

	file, err := os.Create(filePath)
	if err != nil {
		return fmt.Errorf("failed to create output file: %w", err)
	}
	defer file.Close()

	_, err = file.Write(audioData)
	if err != nil {
		return fmt.Errorf("failed to write audio data to file: %w", err)
	}

	return nil
}

// SynthesizeOptions holds options for text-to-speech synthesis
type SynthesizeOptions struct {
	Model  openai.SpeechModel          // TTS model to use
	Voice  openai.SpeechVoice          // Voice to use
	Format openai.SpeechResponseFormat // Audio format
	Speed  float64                     // Speed of speech (0.25 to 4.0)
}

// GetDefaultOptions returns default synthesis options
func GetDefaultOptions() SynthesizeOptions {
	return SynthesizeOptions{
		Model:  openai.TTSModel1,
		Voice:  openai.VoiceAlloy,
		Format: openai.SpeechResponseFormatMp3,
		Speed:  1.0,
	}
}

// Available voices as constants for convenience
const (
	VoiceAlloy   = openai.VoiceAlloy
	VoiceEcho    = openai.VoiceEcho
	VoiceFable   = openai.VoiceFable
	VoiceOnyx    = openai.VoiceOnyx
	VoiceNova    = openai.VoiceNova
	VoiceShimmer = openai.VoiceShimmer
)
