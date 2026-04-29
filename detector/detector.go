package detector

import (
	"bufio"
	"io"
	"regexp"
	"strings"
	"sync"
	"time"
)

var defaultPatterns = map[string]string{
	"yes_no":          `(?i)\(yes|no\)|\by/n\b|\[Y/n\]|\[y/N\]`,
	"always_never":    `(?i)\b(always|never|skip)\b`,
	"do_you_want":     `(?i)(do you want to|would you like to|are you sure)`,
	"press_enter":     `(?i)press enter to continue`,
	"overwrite":       `(?i)(overwrite|replace|delete|remove)\?`,
	"numbered_choice": `(?i)^\s*\[\d+\]`,
	"yn_bracket":      `\[y/n\]`,
	"confirm":         `(?i)confirm\?`,
}

type Detector struct {
	patterns []*regexp.Regexp
	mu       sync.Mutex
	lastPlay time.Time
	cooldown time.Duration
}

type Config struct {
	UseDefaults      bool
	Extra            []string
	DisabledDefaults []string
	CooldownMs       int
}

func New(cfg Config) (*Detector, error) {
	d := &Detector{
		cooldown: time.Duration(cfg.CooldownMs) * time.Millisecond,
	}

	disabled := make(map[string]bool)
	for _, key := range cfg.DisabledDefaults {
		disabled[key] = true
	}

	var patterns []string
	if cfg.UseDefaults {
		for key, pat := range defaultPatterns {
			if !disabled[key] {
				patterns = append(patterns, pat)
			}
		}
	}
	patterns = append(patterns, cfg.Extra...)

	for _, pat := range patterns {
		re, err := regexp.Compile(pat)
		if err != nil {
			return nil, err
		}
		d.patterns = append(d.patterns, re)
	}

	return d, nil
}

func (d *Detector) Matches(line string) bool {
	line = strings.TrimSpace(line)
	for _, re := range d.patterns {
		if re.MatchString(line) {
			return true
		}
	}
	return false
}

func (d *Detector) Watch(r io.Reader, onMatch func()) {
	scanner := bufio.NewScanner(r)
	for scanner.Scan() {
		line := scanner.Text()
		if d.Matches(line) {
			d.mu.Lock()
			elapsed := time.Since(d.lastPlay)
			canPlay := elapsed >= d.cooldown
			if canPlay {
				d.lastPlay = time.Now()
			}
			d.mu.Unlock()

			if canPlay {
				go onMatch()
			}
		}
	}
}
