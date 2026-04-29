package main

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
	"os/exec"
	"os/signal"
	"strconv"
	"strings"
	"syscall"

	"github.com/mulhamna/broask/audio"
	"github.com/mulhamna/broask/config"
	"github.com/mulhamna/broask/detector"
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

func wrap(args []string) error {
	cfg, err := config.Load()
	if err != nil {
		return err
	}

	det, err := detector.New(detector.Config{
		UseDefaults:      cfg.Patterns.UseDefaults,
		Extra:            cfg.Patterns.Extra,
		DisabledDefaults: cfg.Patterns.DisabledDefaults,
		CooldownMs:       cfg.CooldownMs,
	})
	if err != nil {
		return fmt.Errorf("detector: %w", err)
	}

	cmd := exec.Command(args[0], args[1:]...)
	cmd.Stdin = os.Stdin
	cmd.Stderr = os.Stderr

	pr, pw := io.Pipe()
	cmd.Stdout = io.MultiWriter(os.Stdout, pw)

	aCfg := audio.Config{
		Sound:           cfg.Sound,
		CustomSoundPath: cfg.CustomSoundPath,
		Volume:          cfg.Volume,
	}

	go det.Watch(pr, func() {
		audio.Play(aCfg) //nolint
	})

	sig := make(chan os.Signal, 1)
	signal.Notify(sig, syscall.SIGINT, syscall.SIGTERM)
	go func() {
		s := <-sig
		if cmd.Process != nil {
			cmd.Process.Signal(s)
		}
	}()

	err = cmd.Run()
	pw.Close()

	if err != nil {
		if exitErr, ok := err.(*exec.ExitError); ok {
			os.Exit(exitErr.ExitCode())
		}
		return err
	}
	return nil
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

func printUsage() {
	fmt.Print(`broask — play a sound when CLI tools ask for confirmation

Usage:
  broask -- <command> [args...]   wrap a command
  broask config show              print config
  broask config set <key> <val>   update config
  broask config add-pattern <rx>  add custom pattern
  broask config reset             reset to defaults
  broask sounds                   list bundled sounds
  broask test-sound               test current sound
  broask version                  show version
`)
}
