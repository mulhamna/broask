//go:build darwin

package audio

import (
	"fmt"
	"os"
	"os/exec"
	"strconv"
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

	vol := strconv.FormatFloat(volume, 'f', 2, 64)
	cmd := exec.Command("afplay", "-v", vol, f.Name())
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("afplay: %w", err)
	}
	return nil
}
