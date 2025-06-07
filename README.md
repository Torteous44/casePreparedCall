# Call Service

A Go-based call service application for handling audio processing, speech recognition, and orchestration.

## Overview

This service provides:
- Voice Activity Detection (VAD)
- Speech-to-Text (STT) processing using **AssemblyAI**
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

The STT service supports multiple transcription methods:

- **Transcribe from bytes**: `stt.Transcribe(audioData)`
- **Transcribe from file**: `stt.TranscribeFile(filePath)`
- **Transcribe from URL**: `stt.TranscribeFromURL(audioURL)`
- **Transcribe from stream**: `stt.TranscribeStream(reader)`
- **Transcribe with options**: `stt.TranscribeWithOptions(audioData, opts)`

### Text-to-Speech (OpenAI)

The TTS service supports various synthesis options:

- **Basic synthesis**: `tts.Synthesize(text)`
- **Custom voice**: `tts.SynthesizeWithVoice(text, voice)`
- **Advanced options**: `tts.SynthesizeWithOptions(text, opts)`
- **Save to file**: `tts.SynthesizeToFile(text, filePath)`

Available voices: `alloy`, `echo`, `fable`, `onyx`, `nova`, `shimmer`

## Configuration

Configuration is managed through YAML files in the `configs/` directory and environment variables in the `.env` file.

## Development

### Building

```bash
go build ./...
```

### Testing

```bash
go test ./...
```

## Deployment

Use the deployment script:

```bash
./scripts/deploy.sh
```

## Contributing

1. Fork the repository
2. Create your feature branch (`git checkout -b feature/amazing-feature`)
3. Commit your changes (`git commit -m 'Add some amazing feature'`)
4. Push to the branch (`git push origin feature/amazing-feature`)
5. Open a Pull Request

## License

This project is licensed under the MIT License - see the LICENSE file for details. 