# GoAct Stack

A full-stack web application boilerplate using Go, React, and modern tooling.

## Tech Stack

### Backend
- **Go** - Backend language
- **Echo** - Web framework
- **Cobra** - CLI framework
- **Viper** - Configuration management
- **Air** - Hot-reload for development

### Frontend
- **React 19** - UI library
- **TypeScript** - Type-safe JavaScript
- **Vite** - Build tool and dev server
- **Tailwind CSS v4** - Utility-first CSS
- **HeroUI v3** - Component library (beta)

### Tooling
- **Bun** - JavaScript package manager
- **Dagger** - CI/CD pipelines
- **Docker Compose** - Development environment
- **Taskfile** - Task runner

## Quick Start

### Prerequisites
- Docker and Docker Compose
- [Task](https://taskfile.dev/) (optional, for task commands)

### Development

Start the development environment:

```bash
# With Task
task dev

# Or directly with Docker Compose
docker compose up --build
```

This starts:
- **Go backend** on `http://localhost:8080` (with Air hot-reload)
- **Vite dev server** on `http://localhost:3000` (with HMR)

The Vite dev server proxies `/api/*` requests to the Go backend.

### Available Tasks

```bash
task              # List all available tasks

# Development
task dev          # Start dev environment
task dev-stop     # Stop dev environment
task dev-logs     # View logs
task dev-shell    # Open shell in container

# Build & Release
task build        # Build production container via Dagger
task build-local  # Build Go binary locally

# Testing & Linting
task test         # Run Go tests
task lint         # Run Go linter
task lint-frontend # Run ESLint
task typecheck    # Run TypeScript type-check
task check        # Run all checks

# Formatting
task fmt          # Format Go code
task fmt-frontend # Format frontend code

# Cleanup
task clean        # Remove build artifacts
task clean-docker # Remove Docker volumes
```

## Project Structure

```
goact-stack/
├── cmd/goact-stack/     # Application entry point
│   ├── main.go          # Main function
│   ├── root.go          # Cobra root command + Viper config
│   ├── server.go        # Echo server setup
│   └── version.go       # Version command
├── internal/
│   └── static/          # Embedded frontend assets
│       └── static.go    # Go embed directive
├── frontend/            # React frontend
│   ├── src/
│   │   ├── main.tsx     # React entry point
│   │   ├── App.tsx      # Main component
│   │   └── index.css    # Tailwind + HeroUI styles
│   ├── package.json
│   ├── vite.config.ts
│   └── tsconfig.json
├── version/             # Version package for ldflags
├── scripts/             # Development scripts
├── .dagger/             # Dagger CI/CD pipeline
├── docker-compose.yml   # Dev environment
├── Dockerfile           # Multi-stage build
├── Taskfile.yml         # Task definitions
└── .air.toml            # Air hot-reload config
```

## Configuration

Configuration can be provided via:
1. Config file (`goact-stack.yaml`)
2. Environment variables (prefixed with `GOACT_`)
3. CLI flags

| Setting | Flag | Env Var | Default |
|---------|------|---------|---------|
| Port | `--port` | `GOACT_PORT` | `:8080` |
| Log Level | `--log-level` | `GOACT_LOG_LEVEL` | `info` |

## Production Build

The production build embeds the frontend assets into the Go binary for single-file deployment:

```bash
# Build via Dagger (recommended)
task build

# Or build locally
cd frontend && bun install && bun run build && cd ..
go build -ldflags "-X goact-stack/version.Tag=$(git describe --tags)" \
  -o ./bin/goact-stack ./cmd/goact-stack/
```

## CI/CD with Dagger

The Dagger pipeline provides:
- `lint` - Go linting with golangci-lint
- `lint-frontend` - ESLint for TypeScript/React
- `typecheck` - TypeScript type checking
- `test` - Go tests
- `build` - Full build (lint → test → build assets → compile)
- `release` - Production container image

```bash
# Run individual steps
dagger call lint --source=.
dagger call test --source=.
dagger call typecheck --source=.

# Full build
dagger call build --source=.

# Production release
dagger call release --source=.
```

## API Endpoints

| Method | Path | Description |
|--------|------|-------------|
| GET | `/api/health` | Health check endpoint |

## Adding New Features

### Adding API Endpoints

Edit `cmd/goact-stack/server.go`:

```go
api := e.Group("/api")
api.GET("/health", healthHandler)
api.GET("/users", usersHandler)  // Add new endpoint
```

### Adding React Components

Use HeroUI v3 components:

```tsx
import { Button, Card } from "@heroui/react";

// HeroUI v3 uses compound components
<Card>
  <Card.Header>
    <Card.Title>Title</Card.Title>
  </Card.Header>
  <Card.Content>Content</Card.Content>
</Card>
```

## License

MIT
