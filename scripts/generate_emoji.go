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

// Emoji codepoints from processor.go iconBadges map
var emojiCodepoints = []string{
	// Status
	"2705", // ✅
	"274c", // ❌
	"26a0", // ⚠
	"2139", // ℹ
	"1f6d1", // 🛑
	"2714", // ✔

	// Actions
	"1f680", // 🚀
	"23f1", // ⏱
	"1f4ca", // 📊
	"1f4c8", // 📈
	"1f4c9", // 📉
	"1f50d", // 🔍
	"1f527", // 🔧
	"1f6e0", // 🛠
	"1f504", // 🔄

	// Objects
	"1f4b0", // 💰
	"1f4a1", // 💡
	"1f3af", // 🎯
	"1f381", // 🎁
	"1f3c6", // 🏆
	"1f4e7", // 📧
	"1f4de", // 📞
	"1f4c5", // 📅
	"1f4dd", // 📝
	"1f4cc", // 📌
	"1f517", // 🔗

	// Arrows/Direction
	"27a1", // ➡
	"2b05", // ⬅
	"2b06", // ⬆
	"2b07", // ⬇
	"2197", // ↗
	"2198", // ↘

	// Emotions
	"1f389", // 🎉
	"1f44d", // 👍
	"1f44e", // 👎
	"1f600", // 😀
	"1f622", // 😢
	"1f4aa", // 💪
	"1f44c", // 👌
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
