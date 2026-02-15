package chunk

import (
	"strings"
	"unicode/utf8"
)

type Config struct {
	MaxChars  int
	Overlap   int
	MinChars  int
	HardLimit int
}

// Split is a deterministic, low-surprise chunker.
// v1 uses a paragraph-aware split with overlap, then falls back to hard splits.
func Split(cfg Config, content string) []string {
	content = strings.TrimSpace(content)
	if content == "" {
		return nil
	}

	paras := splitParas(content)
	var chunks []string
	var cur strings.Builder

	flush := func() {
		c := strings.TrimSpace(cur.String())
		cur.Reset()
		if utf8.RuneCountInString(c) >= cfg.MinChars {
			chunks = append(chunks, c)
		}
	}

	for _, p := range paras {
		p = strings.TrimSpace(p)
		if p == "" {
			continue
		}
		if runeLen(cur.String())+runeLen(p)+2 <= cfg.MaxChars {
			if cur.Len() > 0 {
				cur.WriteString("\n\n")
			}
			cur.WriteString(p)
			continue
		}

		// Current chunk is full; flush then start new with overlap.
		prev := strings.TrimSpace(cur.String())
		flush()

		over := takeTailRunes(prev, cfg.Overlap)
		if over != "" {
			cur.WriteString(over)
			cur.WriteString("\n\n")
		}

		// If paragraph is huge, hard-split it.
		if runeLen(p) > cfg.MaxChars {
			for _, h := range hardSplitRunes(p, cfg.MaxChars) {
				if cur.Len() > 0 {
					flush()
				}
				cur.WriteString(h)
				flush()
			}
			continue
		}

		cur.WriteString(p)
	}

	if cur.Len() > 0 {
		flush()
	}

	if cfg.HardLimit > 0 && len(chunks) > cfg.HardLimit {
		return chunks[:cfg.HardLimit]
	}
	return chunks
}

func splitParas(s string) []string {
	// normalize CRLF
	s = strings.ReplaceAll(s, "\r\n", "\n")
	return strings.Split(s, "\n\n")
}

func runeLen(s string) int { return utf8.RuneCountInString(s) }

func takeTailRunes(s string, n int) string {
	if n <= 0 || s == "" {
		return ""
	}
	r := []rune(s)
	if len(r) <= n {
		return string(r)
	}
	return string(r[len(r)-n:])
}

func hardSplitRunes(s string, max int) []string {
	if max <= 0 {
		return []string{s}
	}
	r := []rune(s)
	if len(r) <= max {
		return []string{s}
	}
	var out []string
	for len(r) > 0 {
		end := max
		if end > len(r) {
			end = len(r)
		}
		out = append(out, string(r[:end]))
		r = r[end:]
	}
	return out
}
