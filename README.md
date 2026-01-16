# pmux

Plugin-based tmux session manager with web interface.

```
pmux = tmux + tmuxinator + ttyd + web UI
```

## Features

- **Session Management**: Create, attach, and manage tmux sessions
- **Web Terminal**: Browser-based terminal via ttyd
- **Web Dashboard**: Manage sessions from a web UI
- **Multi-user**: Share sessions for pair programming

## Installation

### Prerequisites

```bash
# macOS (Homebrew)
brew install tmux tmuxinator ttyd

# Linux
apt install tmux
gem install tmuxinator
# ttyd: https://github.com/tsl0922/ttyd#installation
```

### Install pmux

```bash
go install github.com/es6kr/pmux/cmd/pmux@latest
```

Or build from source:

```bash
git clone https://github.com/es6kr/pmux.git
cd pmux
make build
```

## Usage

### Check dependencies

```bash
pmux doctor
```

### Session management

```bash
# Start a new session
pmux start dev

# List sessions
pmux list

# Attach to a session
pmux attach dev

# Stop a session
pmux stop dev
```

### Web terminal (ttyd)

```bash
# Start web terminal for a session
pmux ttyd start dev --port 7681
# Open http://localhost:7681

# List running ttyd instances
pmux ttyd list

# Stop web terminal
pmux ttyd stop dev
```

### Web dashboard

```bash
pmux web --port 8080
# Open http://localhost:8080
```

## Architecture

```
pmux (Go)
   │
   ├── CLI commands
   │   ├── start/stop/attach/list (tmux)
   │   ├── ttyd start/stop/list
   │   └── web (dashboard)
   │
   └── Wraps existing tools
       ├── tmux (session multiplexer)
       ├── tmuxinator (session templates)
       └── ttyd (web terminal)
```

## Development

```bash
# Install dependencies
make deps

# Generate templ files
make templ

# Build
make build

# Run with live reload
make dev
```

## License

MIT License - see [LICENSE](LICENSE)
