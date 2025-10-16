package main

import (
	"fmt"
	"image"
	"image/color"
	"image/png"
	"os"
	"path/filepath"
	"strings"

	"golang.org/x/image/draw"
)

// Emoji codepoints to generate
// Auto-discovered from PAGE_TWEAKS.md + manually curated common emoji
var emojiCodepoints = []string{
	// Symbols (U+2xxx)
	"2139", // ℹ INFO
	"2197", // ↗ UP-RIGHT
	"2198", // ↘ DOWN-RIGHT
	"23f1", // ⏱ STOPWATCH
	"23f3", // ⏳ HOURGLASS
	// "25cb", // ○ WHITE CIRCLE (not in Twemoji)
	"2614", // ☔ UMBRELLA
	"26a0", // ⚠ WARNING
	"2705", // ✅ CHECK MARK
	"2714", // ✔ HEAVY CHECK
	"274c", // ❌ CROSS
	"27a1", // ➡ RIGHT ARROW
	"2b05", // ⬅ LEFT ARROW
	"2b06", // ⬆ UP ARROW
	"2b07", // ⬇ DOWN ARROW
	"2b50", // ⭐ STAR

	// Emoji (U+1Fxxx)
	"1f381", // 🎁 WRAPPED PRESENT
	"1f389", // 🎉 PARTY POPPER
	"1f393", // 🎓 GRADUATION CAP
	"1f3af", // 🎯 DIRECT HIT
	"1f3c6", // 🏆 TROPHY
	"1f44c", // 👌 OK HAND
	"1f44d", // 👍 THUMBS UP
	"1f44e", // 👎 THUMBS DOWN
	"1f465", // 👥 BUSTS IN SILHOUETTE
	"1f4a1", // 💡 LIGHT BULB
	"1f4aa", // 💪 FLEXED BICEPS
	"1f4ac", // 💬 SPEECH BALLOON
	"1f4b0", // 💰 MONEY BAG
	"1f4bc", // 💼 BRIEFCASE
	"1f4c5", // 📅 CALENDAR
	"1f4c8", // 📈 CHART INCREASING
	"1f4c9", // 📉 CHART DECREASING
	"1f4ca", // 📊 BAR CHART
	"1f4cc", // 📌 PUSHPIN
	"1f4dd", // 📝 MEMO
	"1f4de", // 📞 TELEPHONE
	"1f4e7", // 📧 E-MAIL
	"1f4f8", // 📸 CAMERA WITH FLASH
	"1f504", // 🔄 COUNTERCLOCKWISE
	"1f50d", // 🔍 MAGNIFYING GLASS
	"1f517", // 🔗 LINK
	"1f51c", // 🔜 SOON ARROW
	"1f527", // 🔧 WRENCH
	"1f52e", // 🔮 CRYSTAL BALL
	"1f600", // 😀 GRINNING FACE
	"1f622", // 😢 CRYING FACE
	"1f680", // 🚀 ROCKET
	"1f6a7", // 🚧 CONSTRUCTION
	"1f6a8", // 🚨 POLICE LIGHT
	"1f6d1", // 🛑 STOP SIGN
	"1f6e0", // 🛠 HAMMER AND WRENCH
	"1f916", // 🤖 ROBOT
	"1f91d", // 🤝 HANDSHAKE

	// Keycaps (multi-codepoint: digit + U+20E3)
	"30-20e3", // 0️⃣
	"31-20e3", // 1️⃣
	"32-20e3", // 2️⃣
	"33-20e3", // 3️⃣
	"34-20e3", // 4️⃣
	"35-20e3", // 5️⃣
	"36-20e3", // 6️⃣
	"37-20e3", // 7️⃣
	"38-20e3", // 8️⃣
	"39-20e3", // 9️⃣
	"23-20e3", // #️⃣
	"2a-20e3", // *️⃣
}

// convertToGrayscaleAlpha converts a color image to grayscale while preserving alpha channel
func convertToGrayscaleAlpha(img image.Image) *image.NRGBA {
	bounds := img.Bounds()
	gray := image.NewNRGBA(bounds)

	for y := bounds.Min.Y; y < bounds.Max.Y; y++ {
		for x := bounds.Min.X; x < bounds.Max.X; x++ {
			r, g, b, a := img.At(x, y).RGBA()

			// Convert to grayscale using luminosity method
			// Y = 0.299*R + 0.587*G + 0.114*B
			grayValue := (299*r + 587*g + 114*b) / 1000
			grayByte := uint8(grayValue >> 8)

			// Preserve alpha channel
			alphaByte := uint8(a >> 8)

			gray.Set(x, y, color.NRGBA{
				R: grayByte,
				G: grayByte,
				B: grayByte,
				A: alphaByte,
			})
		}
	}

	return gray
}

func main() {
	// Check if twemoji directory exists
	twemojiPath := "../twemoji/assets/72x72"
	if _, err := os.Stat(twemojiPath); os.IsNotExist(err) {
		fmt.Fprintf(os.Stderr, "ERROR: %s not found\n", twemojiPath)
		fmt.Fprintf(os.Stderr, "Clone twemoji repository:\n")
		fmt.Fprintf(os.Stderr, "  cd /Users/aleksander/Documents/projects4\n")
		fmt.Fprintf(os.Stderr, "  git clone https://github.com/jdecked/twemoji.git\n")
		os.Exit(1)
	}

	// Ensure output directory exists
	outputDir := "assets/emoji"
	if err := os.MkdirAll(outputDir, 0755); err != nil {
		fmt.Fprintf(os.Stderr, "ERROR: Failed to create output directory: %v\n", err)
		os.Exit(1)
	}

	successCount := 0
	failCount := 0

	fmt.Println("Generating grayscale emoji PNGs from Twemoji 72x72 rasters...")
	fmt.Println("=" + strings.Repeat("=", 60))

	for _, codepoint := range emojiCodepoints {
		sourcePNG := filepath.Join(twemojiPath, codepoint+".png")
		outputPNG := filepath.Join(outputDir, codepoint+".png")

		fmt.Printf("Processing %s... ", codepoint)

		// Read source PNG file
		srcFile, err := os.Open(sourcePNG)
		if err != nil {
			fmt.Printf("FAILED (read error: %v)\n", err)
			failCount++
			continue
		}

		// Decode PNG
		srcImg, err := png.Decode(srcFile)
		srcFile.Close()
		if err != nil {
			fmt.Printf("FAILED (decode error: %v)\n", err)
			failCount++
			continue
		}

		// Convert to grayscale (preserving alpha)
		grayImg := convertToGrayscaleAlpha(srcImg)

		// Scale to 128x128 using bilinear interpolation
		targetSize := 128
		scaledImg := image.NewNRGBA(image.Rect(0, 0, targetSize, targetSize))
		draw.BiLinear.Scale(scaledImg, scaledImg.Bounds(), grayImg, grayImg.Bounds(), draw.Over, nil)

		// Save as PNG
		outFile, err := os.Create(outputPNG)
		if err != nil {
			fmt.Printf("FAILED (create error: %v)\n", err)
			failCount++
			continue
		}

		if err := png.Encode(outFile, scaledImg); err != nil {
			outFile.Close()
			fmt.Printf("FAILED (encode error: %v)\n", err)
			failCount++
			continue
		}

		outFile.Close()
		fmt.Printf("OK\n")
		successCount++
	}

	fmt.Println("=" + strings.Repeat("=", 60))
	fmt.Printf("Results: %d succeeded, %d failed, %d total\n", successCount, failCount, len(emojiCodepoints))

	if failCount > 0 {
		os.Exit(1)
	}
}
