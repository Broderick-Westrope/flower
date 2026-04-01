# 🏵️ 🍃 🧘 Flower

A minimal CLI & TUI tool for the Flowtime Technique - a productivity method similar to Pomodoro but with an emphasis on following your natural rhythm of flow/break cycles. It was created by Zoe Read-Bivens as described [here](https://medium.com/@UrgentPigeon/the-flowtime-technique-7685101bd191).

![TUI Demo](./assets/tui-demo.gif)

## How It Works

1. **Start** a flow session by stating the task you're going to focus on.
2. Take a **Break** when you feel ready. Flower will suggest a break duration based on how long you worked.
3. Then **Resume** the same task, **Start** a new task, or **Stop** the session.
4. You can always **Cancel** a session if you made a mistake with the task description or don't want to track it.

All sessions are automatically tracked locally and can be viewed with `flower log`. Deleted sessions are soft-deleted (data is retained but hidden from display) so you can still access data after accidental deletion.

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

## License

GPL-3.0
