# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project Overview

gobs-cli is a command-line interface for OBS Websocket v5, written in Go. It provides comprehensive control over OBS Studio through websocket commands, allowing users to manage scenes, inputs, recording, streaming, and other OBS features from the command line.

## Development Commands

### Building
- `task build` - Build the project (runs vet, fmt, then builds for Windows and Linux)
- `task` - Default task, same as `task build`
- `go build` - Build for current platform only

### Code Quality
- `task fmt` - Format code with `go fmt`
- `task vet` - Run `go vet` for static analysis
- `go fmt ./...` - Format all Go files
- `go vet ./...` - Vet all Go files

### Testing
- `task test` - Run all tests
- `go test ./...` - Run all tests
- `go test -run TestFilterList` - Run specific test
- `go test -v ./...` - Run tests with verbose output

### Cleaning
- `task clean` - Remove build artifacts from bin/ directory

## Architecture

### Core Structure
- **main.go**: Entry point with CLI configuration using Kong framework
- **Command modules**: Each OBS feature has its own Go file (scene.go, input.go, record.go, etc.)
- **Test files**: Comprehensive test coverage with *_test.go files for each module

### Key Components
- **CLI struct**: Main command structure embedding ObsConfig and StyleConfig
- **context struct**: Wraps goobs.Client, output writer, and styling configuration
- **Command pattern**: Each OBS feature (scenes, inputs, recording) is implemented as separate command structs

### Dependencies
- **Kong**: CLI argument parsing and command routing
- **goobs**: OBS WebSocket client library
- **lipgloss**: Terminal styling and formatting
- **Taskfile**: Build automation (replaces Make)

### Configuration
- Environment variables: OBS_HOST, OBS_PORT, OBS_PASSWORD, OBS_TIMEOUT
- Config files: `.env` in cwd or `$XDG_CONFIG_HOME/gobs-cli/config.env`
- Styling: GOBS_STYLE and GOBS_STYLE_NO_BORDER environment variables

### Command Structure
Commands are organized by OBS feature domains:
- Scene management (scene.go)
- Scene items (sceneitem.go) 
- Groups (group.go)
- Inputs (input.go)
- Text inputs (text.go)
- Recording (record.go)
- Streaming (stream.go)
- Profiles and collections (profile.go, scenecollection.go)
- Replay buffer (replaybuffer.go)
- Studio mode (studiomode.go)
- Virtual camera (virtualcam.go)
- Hotkeys (hotkey.go)
- Filters (filter.go)
- Projectors (projector.go)
- Screenshots (screenshot.go)

Each command file implements its own command structs that embed the main context for OBS client access.

## Code Conventions

### Command Naming
- Go struct names must match command names: if the command is `pass`, the struct should be `PassCmd`
- Function names should reflect the command: `func (cmd *PassCmd) Run()`
- Help text and comments should use the actual command name, not internal implementation names