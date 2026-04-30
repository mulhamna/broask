//go:build windows

package main

import (
	"fmt"
	"io"
	"os"
	"os/exec"
	"os/signal"
	"syscall"

	"github.com/mulhamna/broask/audio"
	"github.com/mulhamna/broask/config"
	"github.com/mulhamna/broask/detector"
)

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
