# Call Service

A Go-based call service application for handling audio processing, speech recognition, and orchestration.

## Overview

This service provides:
- Voice Activity Detection (VAD)
- Speech-to-Text (STT) processing
- Text-to-Speech (TTS) synthesis
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
- TODO: Add other dependencies

### Installation

```bash
go mod download
```

### Running the Service

```bash
go run cmd/callservice/main.go
```

## Configuration

Configuration is managed through YAML files in the `configs/` directory.

## TODO

- Add setup instructions
- Add API documentation
- Add testing instructions
- Add deployment guide 