# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project Overview

VRChat Print Upload CLI - A Go command-line tool for uploading images to VRChat's print feature with 2FA authentication support.

## Development Commands

### Build
```bash
go build -o vrc-print cmd/vrc-print/main.go
```

### Test
```bash
go test ./...
```

### Install globally
```bash
go install github.com/yoshiken/vrc-print-upload/cmd/vrc-print@latest
```

### Linting
No linting configuration found. Consider using:
```bash
go fmt ./...
go vet ./...
```

## Architecture

### Command Structure (using Cobra framework)
- `login` - Authenticate with VRChat API, supports 2FA
- `upload` - Upload images to VRChat prints
- `auth status` - Check authentication status
- `auth logout` - Clear saved credentials
- `config` - Display current configuration

### Package Organization
- `cmd/vrc-print/main.go` - Entry point with Cobra command definitions
- `internal/auth/` - Authentication logic, cookie management, 2FA handling
- `internal/client/` - HTTP client wrapper for VRChat API
- `internal/config/` - Configuration management (loads from ~/.vrc-print/config.yaml)
- `internal/upload/` - Image processing (resize to 1080p) and upload logic

### Key Dependencies
- `github.com/spf13/cobra` - Command-line interface
- `github.com/spf13/viper` - Configuration management
- `github.com/go-resty/resty/v2` - HTTP client
- `github.com/disintegration/imaging` - Image processing
- `golang.org/x/term` - Terminal input handling (password masking)

### Configuration
- Config directory: `~/.vrc-print/`
- Cookie storage: `~/.vrc-print/cookies.json` (permissions 0700)
- Config file: `~/.vrc-print/config.yaml`
- Environment variable: `VRC_PRINT_API_BASE_URL`

### Image Processing
- Supported formats: PNG, JPEG, GIF (auto-converts to PNG)
- Max resolution: 2048×2048 pixels
- Upload resolution: 1080p (1920×1080 or 1080×1920)
- Max file size: 32MB

## Development Notes

- Go 1.21+ required
- The tool is written with Japanese documentation (README.md)
- No existing tests - consider adding tests for auth, upload, and client packages
- Authentication uses persistent cookies with proper security (0700 permissions)
- Rate limiting consideration: Keep requests under 60 seconds apart