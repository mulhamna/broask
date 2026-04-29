package audio

import (
	"embed"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

//go:embed sounds/*.wav
var SoundFS embed.FS

type Config struct {
	Sound           string
	CustomSoundPath string
	Volume          float64
}

func userSoundsDir() string {
	home, _ := os.UserHomeDir()
	return filepath.Join(home, ".broask", "sounds")
}

func Play(cfg Config) error {
	var data []byte
	var err error

	if cfg.CustomSoundPath != "" {
		data, err = os.ReadFile(cfg.CustomSoundPath)
		if err != nil {
			return fmt.Errorf("read custom sound: %w", err)
		}
		return playPlatform(data, cfg.Volume)
	}

	name := cfg.Sound
	if name == "" {
		name = "default"
	}

	// check user sounds dir first
	userPath := filepath.Join(userSoundsDir(), name+".wav")
	if d, err := os.ReadFile(userPath); err == nil {
		return playPlatform(d, cfg.Volume)
	}

	data, err = SoundFS.ReadFile("sounds/" + name + ".wav")
	if err != nil {
		return fmt.Errorf("sound %q not found (check broask sounds)", name)
	}
	return playPlatform(data, cfg.Volume)
}

// AvailableSounds returns bundled sounds plus any .wav files in ~/.broask/sounds/.
func AvailableSounds() []string {
	seen := map[string]bool{}
	var names []string

	entries, _ := SoundFS.ReadDir("sounds")
	for _, e := range entries {
		n := strings.TrimSuffix(e.Name(), ".wav")
		seen[n] = true
		names = append(names, n)
	}

	dir := userSoundsDir()
	entries2, _ := os.ReadDir(dir)
	for _, e := range entries2 {
		if e.IsDir() {
			continue
		}
		n := strings.TrimSuffix(e.Name(), ".wav")
		if strings.HasSuffix(e.Name(), ".wav") && !seen[n] {
			names = append(names, n+" (custom)")
		}
	}

	return names
}
