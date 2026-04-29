//go:build linux

package audio

import (
	"fmt"
	"os"
	"os/exec"
)

func playPlatform(data []byte, volume float64) error {
	f, err := os.CreateTemp("", "broask-*.wav")
	if err != nil {
		return err
	}
	defer os.Remove(f.Name())

	if _, err := f.Write(data); err != nil {
		f.Close()
		return err
	}
	f.Close()

	backends := [][]string{
		{"paplay", f.Name()},
		{"aplay", f.Name()},
		{"beep"},
	}

	for _, args := range backends {
		if path, err := exec.LookPath(args[0]); err == nil {
			cmd := exec.Command(path, args[1:]...)
			if err := cmd.Start(); err == nil {
				go cmd.Wait()
				return nil
			}
		}
	}

	// terminal bell fallback
	fmt.Print("\a")
	return nil
}
