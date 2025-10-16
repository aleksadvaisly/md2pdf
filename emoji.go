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

func segmentTextWithEmoji(text string) []TextSegment {
	var segments []TextSegment
	var currentText strings.Builder
	var currentRunes []rune

	gr := uniseg.NewGraphemes(text)
	for gr.Next() {
		grapheme := gr.Str()
		runes := gr.Runes()

		if len(runes) == 1 && isEmojiRune(runes[0]) {
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
		} else if len(runes) > 1 && isEmojiRune(runes[0]) {
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

	filename := strings.Join(parts, "_") + ".png"
	return "assets/emoji/" + filename
}
