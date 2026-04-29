//go:build windows

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

	script := fmt.Sprintf(
		`$p = New-Object System.Media.SoundPlayer '%s'; $p.Play()`,
		f.Name(),
	)
	cmd := exec.Command("powershell", "-NoProfile", "-Command", script)
	if err := cmd.Start(); err != nil {
		// fallback: console beep
		fmt.Print("\a")
		return nil
	}
	go cmd.Wait()
	return nil
}
