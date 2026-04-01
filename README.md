# Flower

A minimal CLI & TUI tool for the Flowtime Technique - a productivity method similar to Pomodoro but with an emphasis on following your natural rhythm of flow/break cycles. It was created by Zoe Read-Bivens as described [here](https://medium.com/@UrgentPigeon/the-flowtime-technique-7685101bd191).

## Features

- **Start flow sessions** with task tracking
- **Take breaks** with suggested durations based on work time
- **Resume work** from breaks or previous sessions
- **Cancel sessions** without recording them
- **Soft-delete** individual or all completed sessions
- **View current status** and paginated session history
- **Interactive TUI** with keyboard shortcuts or command-line interface
- **Persistent state** saved as JSON

## Installation

```bash
go install github.com/Broderick-Westrope/flower@latest
```

## Usage

### Interactive Mode

Run `flower` with no arguments to launch the TUI:

```bash
flower
```

The TUI has four views:

| View      | Key     | Action                                      |
| --------- | ------- | ------------------------------------------- |
| **Idle**  | `enter` | Start a session (type a task name first)    |
|           | `l`     | View session log                            |
|           | `q`     | Quit                                        |
| **Flow**  | `space` | Take a break                                |
|           | `s`     | Stop and record session                     |
|           | `c`     | Cancel session (with confirmation)          |
|           | `l`     | View session log                            |
|           | `q`     | Quit                                        |
| **Break** | `space` | Resume working                              |
|           | `s`     | Stop and record session                     |
|           | `c`     | Cancel session (with confirmation)          |
|           | `l`     | View session log                            |
|           | `q`     | Quit                                        |
| **Log**   | `j/k`   | Navigate rows                               |
|           | `d`     | Delete selected session (with confirmation) |
|           | `D`     | Delete all sessions (with confirmation)     |
|           | `esc`   | Back                                        |
|           | `q`     | Quit                                        |

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

# Cancel current session without recording it
flower cancel

# Delete a specific session (1 = most recent)
flower delete 3

# Delete all completed sessions
flower clear

# Find state file location
flower locate
```

### Detached Mode

Add `-d` (for "detach") to run start/break/resume without launching the TUI:

```bash
flower start -d "Background task"
flower break -d
flower resume -d
```

### Skipping Confirmation

The `cancel`, `delete`, and `clear` commands prompt for confirmation by default. Use `-y` to skip:

```bash
flower cancel -y
flower delete 1 -y
flower clear -y
```

## How It Works

1. **Start** a flow session with a task description
2. Work until you naturally feel ready for a break
3. **Break** - the tool suggests a break duration based on how long you worked
4. **Resume** when ready to continue
5. **Stop** to end the session, or **Cancel** to discard it

All sessions are automatically tracked and can be viewed with `flower log`. Deleted sessions are soft-deleted (data is retained but hidden from display).

## License

GPL-3.0
