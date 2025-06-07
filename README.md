# Call Service

A Go-based call service application for handling audio processing, speech recognition, and orchestration.

## Overview

This service provides:
- Voice Activity Detection (VAD)
- Speech-to-Text (STT) processing using **AssemblyAI**
  - Batch transcription for files and audio data
  - **Real-time streaming transcription** via WebSocket
- Text-to-Speech (TTS) synthesis using **OpenAI**
- Real-time audio processing
- WebSocket-based communication
- Session state management
- Context brain integration

## Project Structure

```
.
├── cmd/callservice/          # Main application entry point
├── internal/                 # Internal application packages
│   ├── audio/               # Audio processing components
│   │   ├── vad/             # Voice Activity Detection
│   │   ├── stt/             # Speech-to-Text (AssemblyAI)
│   │   │   ├── stt.go       # Batch transcription
│   │   │   ├── streaming.go # Real-time streaming
│   │   │   └── example.go   # Usage examples
│   │   └── tts/             # Text-to-Speech (OpenAI)
│   ├── orchestrator/        # Call orchestration logic
│   ├── contextbrain/        # Context brain integration
│   └── sessionstate/        # Session management
├── pkg/transport/           # Transport layer (WebSocket)
├── configs/                 # Configuration files
├── docs/                    # Documentation
└── scripts/                 # Deployment and utility scripts
```

## Getting Started

### Prerequisites

- Go 1.19 or later
- AssemblyAI API key ([Get one here](https://assemblyai.com))
- OpenAI API key ([Get one here](https://platform.openai.com))

### Installation

```bash
# Clone the repository
git clone https://github.com/Torteous44/casePreparedCall.git
cd casePreparedCall

# Install dependencies
go mod download
```

### Configuration

1. Copy the example environment file:
   ```bash
   cp .env.example .env
   ```

2. Edit `.env` and add your API keys:
   ```bash
   OPENAI_API_KEY=your_openai_api_key_here
   ASSEMBLYAI_API_KEY=your_assemblyai_api_key_here
   ```

### Running the Service

```bash
go run cmd/callservice/main.go
```

## API Integration

### Speech-to-Text (AssemblyAI)

#### Batch Transcription

The STT service supports multiple transcription methods for pre-recorded audio:

- **Transcribe from bytes**: `stt.Transcribe(audioData)`
- **Transcribe from file**: `stt.TranscribeFile(filePath)`
- **Transcribe from URL**: `stt.TranscribeFromURL(audioURL)`
- **Transcribe from stream**: `stt.TranscribeStream(reader)`
- **Transcribe with options**: `stt.TranscribeWithOptions(audioData, opts)`

#### Real-time Streaming Transcription

For real-time audio processing, use the streaming API:

```go
// Create streaming STT with default config
config := stt.GetDefaultStreamingConfig()
streamSTT := stt.NewStreamingSTT(config)

// Connect to AssemblyAI streaming API
ctx := context.Background()
err := streamSTT.Connect(ctx)
if err != nil {
    log.Fatal(err)
}
defer streamSTT.Close()

// Listen for transcripts
go func() {
    for transcript := range streamSTT.GetTranscripts() {
        if transcript.IsFinal {
            fmt.Printf("Final: %s\n", transcript.Text)
        } else {
            fmt.Printf("Partial: %s\n", transcript.Text)
        }
    }
}()

// Send audio chunks (50ms to 1000ms each)
err = streamSTT.SendAudio(audioChunk)
```

**Streaming Features:**
- **Real-time transcription** with partial and final results
- **Turn-based formatting** for conversation analysis
- **Configurable parameters**: sample rate, encoding, confidence thresholds
- **Dynamic configuration updates** during active sessions
- **Manual endpoint forcing** for immediate results
- **Automatic session management** with graceful termination

**Supported Audio Formats:**
- **Sample Rates**: 8kHz, 16kHz, 22.05kHz, 44.1kHz, 48kHz
- **Encoding**: PCM 16-bit signed little-endian, PCM μ-law
- **Chunk Size**: 50ms to 1000ms per chunk

### Text-to-Speech (OpenAI)

The TTS service supports various synthesis options:

- **Basic synthesis**: `tts.Synthesize(text)`
- **Custom voice**: `tts.SynthesizeWithVoice(text, voice)`
- **Advanced options**: `tts.SynthesizeWithOptions(text, opts)`
- **Save to file**: `tts.SynthesizeToFile(text, filePath)`

Available voices: `alloy`, `echo`, `fable`, `onyx`, `nova`, `shimmer`

## Configuration

Configuration is managed through YAML files in the `configs/` directory and environment variables in the `.env` file.

### Streaming Configuration Options

```go
config := stt.StreamingConfig{
    SampleRate:                       16000,  // Audio sample rate
    Encoding:                         "pcm_s16le", // Audio encoding
    FormatTurns:                      true,   // Enable turn-based formatting
    EndOfTurnConfidenceThreshold:     0.7,    // Confidence threshold for turn detection
    MinEndOfTurnSilenceWhenConfident: 1000,   // Min silence (ms) when confident
    MaxTurnSilence:                   3000,   // Max silence (ms) before turn end
}
```

## Development

### Building

```bash
go build ./...
```

### Testing

```bash
go test ./...
```

### Running Streaming Examples

```bash
go run -c "package main; import \"your-module/internal/audio/stt\"; func main() { stt.StreamingExample() }"
```

## Deployment

Use the deployment script:

```bash
./scripts/deploy.sh
```

## Use Cases

- **Real-time call transcription** for customer service
- **Live meeting notes** with turn-based speaker detection
- **Voice assistants** with immediate response capability
- **Audio streaming applications** with live captions
- **Call analytics** with real-time sentiment analysis

## Contributing

1. Fork the repository
2. Create your feature branch (`git checkout -b feature/amazing-feature`)
3. Commit your changes (`git commit -m 'Add some amazing feature'`)
4. Push to the branch (`git push origin feature/amazing-feature`)
5. Open a Pull Request

## License

This project is licensed under the MIT License - see the LICENSE file for details. 