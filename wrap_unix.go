//go:build darwin || linux

package main

import (
	"fmt"
	"io"
	"os"
	"os/exec"
	"os/signal"
	"syscall"

	"github.com/creack/pty"
	"github.com/mulhamna/broask/audio"
	"github.com/mulhamna/broask/config"
	"github.com/mulhamna/broask/detector"
	"golang.org/x/term"
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

	ptmx, err := pty.Start(cmd)
	if err != nil {
		return fmt.Errorf("pty: %w", err)
	}
	defer ptmx.Close()

	// Propagate terminal resize to child PTY.
	winch := make(chan os.Signal, 1)
	signal.Notify(winch, syscall.SIGWINCH)
	go func() {
		for range winch {
			pty.InheritSize(os.Stdin, ptmx)
		}
	}()
	winch <- syscall.SIGWINCH // set initial size

	// Raw mode so keystrokes pass through unmodified.
	if term.IsTerminal(int(os.Stdin.Fd())) {
		oldState, err := term.MakeRaw(int(os.Stdin.Fd()))
		if err != nil {
			return err
		}
		defer term.Restore(int(os.Stdin.Fd()), oldState)
	}

	// stdin → pty
	go io.Copy(ptmx, os.Stdin)

	aCfg := audio.Config{
		Sound:           cfg.Sound,
		CustomSoundPath: cfg.CustomSoundPath,
		Volume:          cfg.Volume,
	}

	pr, pw := io.Pipe()
	go det.Watch(pr, func() {
		audio.Play(aCfg) //nolint
	})

	buf := make([]byte, 4096)
	for {
		n, readErr := ptmx.Read(buf)
		if n > 0 {
			os.Stdout.Write(buf[:n])
			pw.Write(buf[:n])
		}
		if readErr != nil {
			break
		}
	}
	pw.Close()

	if err := cmd.Wait(); err != nil {
		if exitErr, ok := err.(*exec.ExitError); ok {
			os.Exit(exitErr.ExitCode())
		}
	}
	return nil
}
