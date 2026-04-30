package detector

import (
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

// matches ANSI/VT escape sequences
var ansiRe = regexp.MustCompile(`\x1b(?:\[[0-9;?]*[a-zA-Z]|[^[]|][^\x07]*\x07)`)

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
	line = strings.TrimSpace(ansiRe.ReplaceAllString(line, ""))
	for _, re := range d.patterns {
		if re.MatchString(line) {
			return true
		}
	}
	return false
}

func (d *Detector) tryPlay(onMatch func()) {
	d.mu.Lock()
	canPlay := time.Since(d.lastPlay) >= d.cooldown
	if canPlay {
		d.lastPlay = time.Now()
	}
	d.mu.Unlock()
	if canPlay {
		go onMatch()
	}
}

// Watch reads raw bytes from r, strips ANSI codes, and calls onMatch when a
// pattern is detected. Unlike a line scanner it does not wait for '\n', so
// prompts that never terminate with a newline are still caught.
func (d *Detector) Watch(r io.Reader, onMatch func()) {
	buf := make([]byte, 4096)
	var acc strings.Builder

	for {
		n, err := r.Read(buf)
		if n > 0 {
			clean := ansiRe.ReplaceAllString(string(buf[:n]), "")
			acc.WriteString(clean)

			text := acc.String()
			if d.matchesText(text) {
				d.tryPlay(onMatch)
				acc.Reset()
			} else {
				// Keep only content after the last newline to bound memory.
				if idx := strings.LastIndexByte(text, '\n'); idx >= 0 {
					tail := text[idx+1:]
					acc.Reset()
					acc.WriteString(tail)
				}
				// Hard cap: avoid unbounded growth on no-newline streams.
				if acc.Len() > 2048 {
					s := acc.String()
					acc.Reset()
					acc.WriteString(s[len(s)-512:])
				}
			}
		}
		if err != nil {
			break
		}
	}
}

func (d *Detector) matchesText(text string) bool {
	text = strings.TrimSpace(text)
	for _, re := range d.patterns {
		if re.MatchString(text) {
			return true
		}
	}
	return false
}
