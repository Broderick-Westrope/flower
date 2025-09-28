# Flower

A minimal CLI tool for the Flowtime Technique - a productivity method similar to Pomodoro but with an emphasis on following your natural rhythm of flow/break cycles. It was created by Zoë Read-Bivens as described [here](https://medium.com/@UrgentPigeon/the-flowtime-technique-7685101bd191).

## Features

- **Start flow sessions** with task tracking
- **Take breaks** with suggested durations
- **Resume work** from breaks or previous sessions  
- **View current status** and recent session history
- **Interactive TUI** or command-line interface
- **Persistent state** across sessions

## Installation

```bash
go install github.com/Broderick-Westrope/flower@latest
```

## Usage

### Interactive Mode
```bash
flower
```

### Command Mode
```bash
# Start a work session
flower start "Write documentation"

# Take a break
flower break

# Resume work  
flower resume

# Check current status
flower status

# View recent sessions
flower log

# Stop current session
flower stop

# Find state file location
flower locate
```

### Detached Mode
Add `-d` flag to run commands without the TUI:
```bash
flower start -d "Background task"
flower break -d
```

## How It Works

1. **Start** a flow session with a task description
2. Work until you naturally feel ready for a break
3. **Break** - the tool suggests break duration based on work time
4. **Resume** when ready to continue
5. **Stop** to end the session completely

All sessions are automatically tracked and can be viewed with `flower log`.

## License

GPL-3.0