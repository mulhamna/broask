# broask 🔔

Play a sound whenever a CLI tool asks for your confirmation — so you can tab away and come back when needed.

Works with Claude Code, Cursor, Aider, Gemini CLI, or any interactive terminal tool.

Built in Go — single binary, zero runtime dependency, cross-platform.

---

## Install

### Via Go
```bash
go install github.com/mulhamna/broask@latest
```

### Manual (GitHub Releases)
Download the binary for your platform from [Releases](https://github.com/mulhamna/broask/releases) and add to your `$PATH`.

### Via Homebrew (macOS/Linux)
```bash
brew install mulhamna/tap/broask
```

---

## Usage

```bash
# Wrap any CLI tool
broask -- claude
broask -- aider --model gpt-4o
broask -- gemini
```

That's it. broask pipes stdin/stdout transparently — the tool behaves exactly the same, except you'll hear a 🔔 whenever it asks you something.

---

## Config

Config lives at `~/.broask.json`, auto-created on first run.

```bash
broask config show                            # print current config
broask config set sound chime                 # change bundled sound
broask config set volume 0.5                  # set volume (0.0–1.0)
broask config set cooldown_ms 2000            # min ms between sounds
broask config set custom_sound_path ~/my.wav  # use your own sound file
broask config add-pattern "confirm\?"         # add custom regex pattern
broask config reset                           # reset to defaults
```

### Config schema

```json
{
  "sound": "default",
  "custom_sound_path": "",
  "volume": 0.8,
  "cooldown_ms": 1500,
  "patterns": {
    "use_defaults": true,
    "extra": [],
    "disabled_defaults": []
  }
}
```

---

## Sounds

### Bundled sounds

```bash
broask sounds        # list all available sounds
broask test-sound    # play current sound
```

Bundled options: `default`, `chime`, `ping`, `boop`

### Add your own sounds

Drop any `.wav` file into `~/.broask/sounds/`:

```bash
mkdir -p ~/.broask/sounds
cp ~/Downloads/mybell.wav ~/.broask/sounds/mybell.wav

broask config set sound mybell
broask test-sound
```

Your custom sounds show up in `broask sounds` and work exactly like bundled ones.

---

## Pattern Detection

broask triggers on output lines matching these patterns (case-insensitive):

| Pattern | Examples |
|---|---|
| yes/no prompts | `(yes/no)`, `[Y/n]`, `y/n` |
| always/never/skip | permission prompts |
| confirmation phrases | `do you want to`, `are you sure` |
| destructive actions | `overwrite?`, `delete?`, `remove?` |
| numbered menus | `[1]`, `[2]` choice lists |
| press enter | `press enter to continue` |

Add custom patterns:
```bash
broask config add-pattern "my custom prompt\?"
```

Disable a built-in pattern by adding its key to `disabled_defaults` in `~/.broask.json`:
```json
"disabled_defaults": ["numbered_choice"]
```

---

## Platform Support

| Platform | Audio backend |
|---|---|
| macOS | `afplay` (built-in) |
| Linux | `paplay` → `aplay` → `beep` → terminal bell |
| Windows | PowerShell `System.Media.SoundPlayer` → console beep |

---

## License

MIT
