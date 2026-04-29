# broask 🔔

A universal CLI tool that plays a sound whenever a terminal prompt is waiting for user confirmation (yes/no, always, etc). Works with Claude Code, Cursor, Aider, Gemini CLI, or any interactive CLI tool.

Built in Go — single binary, zero runtime dependency, cross-platform.

---

## How It Works

`broask` acts as a **stdin/stdout pipe wrapper**. It wraps around any command, monitors its output stream in real-time, and triggers a sound whenever it detects a confirmation-style prompt pattern.

```
your terminal → broask → [target CLI tool]
                    ↓
              (pattern match)
                    ↓
              🔔 plays sound
```

---

## Project Structure

```
broask/
├── main.go
├── detector/
│   └── detector.go        # Pattern matching engine
├── audio/
│   ├── audio.go           # Platform-agnostic audio interface
│   ├── audio_darwin.go    # macOS: afplay
│   ├── audio_linux.go     # Linux: paplay / aplay / beep fallback
│   └── audio_windows.go   # Windows: PowerShell Media.SoundPlayer
├── config/
│   └── config.go          # Load/save user config (~/.broask.json)
├── sounds/
│   ├── default.wav         # Default bundled sound (embed via go:embed)
│   ├── chime.wav
│   ├── ping.wav
│   └── boop.wav
├── go.mod
├── go.sum
├── README.md
└── .github/
    └── workflows/
        └── release.yml    # Goreleaser cross-compile CI
```

---

## Core Features

### 1. Universal Pipe Wrapper
- Wrap any CLI: `broask -- claude`, `broask -- aider`, `broask -- gemini`
- Passthrough stdin/stdout/stderr transparently — user experience is identical
- Real-time stream monitoring with zero buffering delay

### 2. Pattern Detection (`detector/detector.go`)
- Regex-based pattern matching on output lines
- Default patterns to detect (case-insensitive):
  - `(yes|no)` or `y/n` or `[Y/n]` or `[y/N]`
  - `(always|never|skip)`
  - `do you want to`, `would you like to`, `are you sure`
  - `press enter to continue`
  - `overwrite?`, `replace?`, `delete?`, `remove?`
  - `[1/2/3...]` numbered choice menus
- User can **add custom patterns** via config
- User can **disable default patterns** individually

### 3. Audio Engine (`audio/`)
- **Bundled default sounds** (embedded via `//go:embed sounds/*.wav`):
  - `default` — a soft notification chime
  - `chime` — classic bell
  - `ping` — short ping
  - `boop` — playful boop
- **User sounds dir**: `~/.broask/sounds/` — drop any `.wav` here, available by filename without extension
  - Resolution order: user sounds dir → bundled → error
  - `broask sounds` lists both bundled + user sounds
- **Custom sound path**: `custom_sound_path` config — absolute path to any `.wav` file (overrides sound name entirely)
- Platform audio backends:
  - **macOS**: `afplay` (built-in, zero deps) — blocking call
  - **Linux**: try `paplay` → `aplay` → `beep` → fallback to terminal bell `\a`
  - **Windows**: PowerShell `[System.Media.SoundPlayer]` or `[System.Console]::Beep()`
- Volume control (0.0–1.0, passed to backend where supported)
- Playback is blocking in `playPlatform`; callers run it in goroutines as needed

### 4. Config System (`config/config.go`)
- Config file: `~/.broask.json`
- Generated on first run with defaults
- Editable manually or via `broask config set <key> <value>`

---

## Config Schema (`~/.broask.json`)

```json
{
  "sound": "default",
  "custom_sound_path": "",
  "volume": 0.8,
  "cooldown_ms": 1500,
  "patterns": {
    "use_defaults": true,
    "extra": [
      "custom pattern here"
    ],
    "disabled_defaults": []
  }
}
```

| Key                          | Type     | Description                                              |
| ---------------------------- | -------- | -------------------------------------------------------- |
| `sound`                      | string   | Bundled sound name: `default`, `chime`, `ping`, `boop`   |
| `custom_sound_path`          | string   | Absolute path to a custom sound file (overrides `sound`) |
| `volume`                     | float    | Playback volume 0.0–1.0                                  |
| `cooldown_ms`                | int      | Minimum ms between sounds (prevents rapid-fire spam)     |
| `patterns.use_defaults`      | bool     | Whether to use built-in detection patterns               |
| `patterns.extra`             | []string | Additional regex patterns to detect                      |
| `patterns.disabled_defaults` | []string | Built-in pattern keys to disable                         |

---

## CLI Usage

```bash
# Basic usage — wrap any CLI tool
broask -- claude
broask -- aider --model gpt-4o
broask -- gemini
broask -- npx claude-dev

# Config management
broask config show                          # Print current config
broask config set sound chime              # Change bundled sound
broask config set volume 0.5               # Set volume
broask config set custom_sound_path /path/to/mybell.wav  # Custom sound
broask config add-pattern "confirm\?"      # Add custom regex pattern
broask config reset                        # Reset to defaults

# List available bundled sounds
broask sounds

# Test your current sound config
broask test-sound

# Show version
broask version
```

---

## Installation

### Via Go
```bash
go install github.com/mulhamna/broask@latest
```

### Via Homebrew (macOS/Linux)
```bash
brew install mulhamna/tap/broask
```

### Via Scoop (Windows)
```bash
scoop bucket add mulhamna https://github.com/mulhamna/scoop-bucket
scoop install broask
```

### Manual (GitHub Releases)
Download the binary for your platform from [Releases](https://github.com/mulhamna/broask/releases) and add to your `$PATH`.

---

## Implementation Notes for Claude

### Pipe Architecture
Use `os/exec` to spawn the target process with stdin/stdout/stderr piped. Use `io.TeeReader` or a custom writer to intercept output while still forwarding it to the real stdout.

```go
// Pseudocode — implement properly in main.go
cmd := exec.Command(args[0], args[1:]...)
cmd.Stdin = os.Stdin

pr, pw := io.Pipe()
cmd.Stdout = io.MultiWriter(os.Stdout, pw)
cmd.Stderr = os.Stderr

go detector.Watch(pr, func() {
    audio.Play(cfg)
})

cmd.Run()
```

### Embedding Sounds
```go
//go:embed sounds/*.wav
var soundFS embed.FS
```

### Cross-Platform Audio
Each `audio_<platform>.go` file implements the same interface:
```go
type Player interface {
    Play(soundData []byte, volume float64) error
}
```
Use build tags: `//go:build darwin`, `//go:build linux`, `//go:build windows`

### Pattern Detection
Compile all regex patterns once at startup, not per-line. Use `regexp.MustCompile` in an `init()` or constructor. Scan line-by-line using `bufio.Scanner`.

### Cooldown
Track `lastPlayedAt time.Time` and skip playback if `time.Since(lastPlayedAt) < cooldown`.

---

## Release Pipeline (`.github/workflows/release.yml`)

Use **GoReleaser** for cross-compilation and GitHub Releases:
- Targets: `linux/amd64`, `linux/arm64`, `darwin/amd64`, `darwin/arm64`, `windows/amd64`
- Auto-generate Homebrew formula and Scoop manifest
- Triggered on `git tag v*`

---

## Out of Scope (for now)
- GUI configuration app
- Plugin system for custom audio backends
- Websocket/remote trigger mode
- tmux / screen session detection

---

## Vibe

Keep it **simple, fast, and zero-friction**. The user should forget it's even running — until they hear the 🔔.
