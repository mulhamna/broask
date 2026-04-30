package main

import (
	"encoding/json"
	"fmt"
	"os"
	"strconv"
	"strings"

	"github.com/mulhamna/broask/audio"
	"github.com/mulhamna/broask/config"
)

var version = "dev"

func main() {
	if err := run(); err != nil {
		fmt.Fprintln(os.Stderr, "broask:", err)
		os.Exit(1)
	}
}

func run() error {
	args := os.Args[1:]

	if len(args) == 0 {
		printUsage()
		return nil
	}

	switch args[0] {
	case "version", "--version", "-v":
		fmt.Println("broask", version)
		return nil

	case "sounds":
		for _, s := range audio.AvailableSounds() {
			fmt.Println(s)
		}
		return nil

	case "test-sound":
		cfg, err := config.Load()
		if err != nil {
			return err
		}
		fmt.Println("playing sound:", cfg.Sound)
		return audio.Play(audio.Config{
			Sound:           cfg.Sound,
			CustomSoundPath: cfg.CustomSoundPath,
			Volume:          cfg.Volume,
		})

	case "config":
		return handleConfig(args[1:])

	case "init":
		shell := ""
		if len(args) > 1 {
			shell = args[1]
		}
		return handleInit(shell)

	case "--":
		if len(args) < 2 {
			return fmt.Errorf("usage: broask -- <command> [args...]")
		}
		return wrap(args[1:])

	default:
		if args[0][0] == '-' {
			printUsage()
			return nil
		}
		return wrap(args)
	}
}

func handleConfig(args []string) error {
	if len(args) == 0 {
		return fmt.Errorf("usage: broask config <show|set|add-pattern|reset>")
	}

	cfg, err := config.Load()
	if err != nil {
		return err
	}

	switch args[0] {
	case "show":
		data, _ := json.MarshalIndent(cfg, "", "  ")
		fmt.Println(string(data))

	case "reset":
		return config.Save(config.Default())

	case "set":
		if len(args) < 3 {
			return fmt.Errorf("usage: broask config set <key> <value>")
		}
		key, val := args[1], args[2]
		switch key {
		case "sound":
			cfg.Sound = val
		case "custom_sound_path":
			cfg.CustomSoundPath = val
		case "volume":
			v, err := strconv.ParseFloat(val, 64)
			if err != nil {
				return fmt.Errorf("volume must be 0.0–1.0")
			}
			cfg.Volume = v
		case "cooldown_ms":
			v, err := strconv.Atoi(val)
			if err != nil {
				return fmt.Errorf("cooldown_ms must be integer")
			}
			cfg.CooldownMs = v
		case "use_defaults":
			cfg.Patterns.UseDefaults = val == "true" || val == "1"
		default:
			return fmt.Errorf("unknown key: %s", key)
		}
		return config.Save(cfg)

	case "add-pattern":
		if len(args) < 2 {
			return fmt.Errorf("usage: broask config add-pattern <regex>")
		}
		pattern := strings.Join(args[1:], " ")
		cfg.Patterns.Extra = append(cfg.Patterns.Extra, pattern)
		return config.Save(cfg)

	default:
		return fmt.Errorf("unknown config command: %s", args[0])
	}
	return nil
}

// defaultWrappedCmds are the AI CLI tools broask wraps by default.
var defaultWrappedCmds = []string{"claude", "aider", "gemini", "sgpt", "llm"}

func handleInit(shell string) error {
	if shell == "" {
		shell = detectShell()
	}
	switch shell {
	case "fish":
		fmt.Println("# broask shell integration — add to ~/.config/fish/config.fish:")
		fmt.Println("# eval (broask init fish)")
		for _, cmd := range defaultWrappedCmds {
			fmt.Printf("function %s\n  command broask -- %s $argv\nend\n", cmd, cmd)
		}
	default: // bash / zsh
		fmt.Println("# broask shell integration — add to ~/.zshrc or ~/.bashrc:")
		fmt.Println("# eval \"$(broask init)\"")
		for _, cmd := range defaultWrappedCmds {
			fmt.Printf("%s() { command broask -- %s \"$@\"; }\n", cmd, cmd)
		}
	}
	return nil
}

func detectShell() string {
	shell := os.Getenv("SHELL")
	switch {
	case strings.HasSuffix(shell, "fish"):
		return "fish"
	case strings.HasSuffix(shell, "zsh"):
		return "zsh"
	default:
		return "bash"
	}
}

func printUsage() {
	fmt.Print(`broask — play a sound when CLI tools ask for confirmation

Usage:
  broask -- <command> [args...]   wrap a command
  broask config show              print config
  broask config set <key> <val>   update config
  broask config add-pattern <rx>  add custom pattern
  broask config reset             reset to defaults
  broask init [shell]             print shell integration (bash/zsh/fish)
  broask sounds                   list bundled sounds
  broask test-sound               test current sound
  broask version                  show version
`)
}
