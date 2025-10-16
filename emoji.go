package mdtopdf

import (
	"embed"
	"fmt"
	"strings"

	"github.com/rivo/uniseg"
)

//go:embed assets/emoji/*.png
var emojiFS embed.FS

type TextSegment struct {
	IsEmoji bool
	Content string
	Runes   []rune
}

func isEmojiRune(r rune) bool {
	_, exists := iconBadges[r]
	if exists {
		return true
	}
	return r > 0x1F000
}

func isEmojiGrapheme(runes []rune) bool {
	if len(runes) == 0 {
		return false
	}

	// Check if grapheme contains combining enclosing keycap (U+20E3)
	// This handles keycap sequences like 1️⃣ = U+0031 + U+FE0F + U+20E3
	for _, r := range runes {
		if r == 0x20E3 {
			return true
		}
	}

	// Single rune: check if it's in iconBadges or in emoji range
	if len(runes) == 1 {
		return isEmojiRune(runes[0])
	}

	// Multi-rune: check if first rune is emoji
	// This handles flags, skin tones, ZWJ sequences, etc.
	return isEmojiRune(runes[0])
}

func segmentTextWithEmoji(text string) []TextSegment {
	var segments []TextSegment
	var currentText strings.Builder
	var currentRunes []rune

	gr := uniseg.NewGraphemes(text)
	for gr.Next() {
		grapheme := gr.Str()
		runes := gr.Runes()

		if isEmojiGrapheme(runes) {
			if currentText.Len() > 0 {
				segments = append(segments, TextSegment{
					IsEmoji: false,
					Content: currentText.String(),
					Runes:   currentRunes,
				})
				currentText.Reset()
				currentRunes = nil
			}

			segments = append(segments, TextSegment{
				IsEmoji: true,
				Content: grapheme,
				Runes:   runes,
			})
		} else {
			currentText.WriteString(grapheme)
			currentRunes = append(currentRunes, runes...)
		}
	}

	if currentText.Len() > 0 {
		segments = append(segments, TextSegment{
			IsEmoji: false,
			Content: currentText.String(),
			Runes:   currentRunes,
		})
	}

	return segments
}

func getEmojiPNGPath(runes []rune) string {
	var parts []string
	for _, r := range runes {
		if r >= 0xFE00 && r <= 0xFE0F {
			continue
		}
		parts = append(parts, fmt.Sprintf("%x", r))
	}

	if len(parts) == 0 {
		return ""
	}

	// Twemoji uses hyphen as separator for multi-codepoint emoji (e.g., 31-20e3.png for 1️⃣)
	filename := strings.Join(parts, "-") + ".png"
	return "assets/emoji/" + filename
}
