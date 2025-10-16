package main

import (
	"encoding/xml"
	"fmt"
	"image"
	"image/color"
	"image/png"
	"os"
	"path/filepath"
	"strings"

	"github.com/srwiley/oksvg"
	"github.com/srwiley/rasterx"
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

// SVGElement represents an XML element in SVG
type SVGElement struct {
	XMLName xml.Name
	Attrs   []xml.Attr `xml:",any,attr"`
	Content []byte     `xml:",innerxml"`
}

// stripColors removes color attributes from SVG XML and replaces with grayscale
func stripColors(svgContent []byte) ([]byte, error) {
	// Replace fill and stroke colors with grayscale values
	content := string(svgContent)

	// Remove all fill and stroke attributes
	// This is a simple approach - just remove color-related attributes
	content = strings.ReplaceAll(content, `fill="#`, `fill="#666666" data-original-fill="#`)
	content = strings.ReplaceAll(content, `stroke="#`, `stroke="#666666" data-original-stroke="#`)

	// Also handle fill/stroke without quotes
	content = strings.ReplaceAll(content, `fill=`, `fill="#666666" data-original-fill=`)
	content = strings.ReplaceAll(content, `stroke=`, `stroke="#666666" data-original-stroke=`)

	return []byte(content), nil
}

// convertToGrayscale converts an RGBA image to grayscale
func convertToGrayscale(img *image.RGBA) *image.Gray {
	bounds := img.Bounds()
	gray := image.NewGray(bounds)

	for y := bounds.Min.Y; y < bounds.Max.Y; y++ {
		for x := bounds.Min.X; x < bounds.Max.X; x++ {
			r, g, b, a := img.At(x, y).RGBA()

			// If pixel is transparent, keep it transparent in grayscale
			if a == 0 {
				gray.Set(x, y, color.Gray{Y: 0})
				continue
			}

			// Convert to grayscale using luminosity method
			// Y = 0.299*R + 0.587*G + 0.114*B
			grayValue := (299*r + 587*g + 114*b) / 1000
			gray.Set(x, y, color.Gray{Y: uint8(grayValue >> 8)})
		}
	}

	return gray
}

func main() {
	// Check if twemoji directory exists
	twemojiPath := "../twemoji/assets/svg"
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

	fmt.Println("Generating grayscale emoji PNGs from Twemoji SVGs...")
	fmt.Println("=" + strings.Repeat("=", 60))

	for _, codepoint := range emojiCodepoints {
		svgPath := filepath.Join(twemojiPath, codepoint+".svg")
		pngPath := filepath.Join(outputDir, codepoint+".png")

		fmt.Printf("Processing %s... ", codepoint)

		// Read SVG file
		svgContent, err := os.ReadFile(svgPath)
		if err != nil {
			fmt.Printf("FAILED (read error: %v)\n", err)
			failCount++
			continue
		}

		// Strip colors from SVG
		graySVG, err := stripColors(svgContent)
		if err != nil {
			fmt.Printf("FAILED (color strip error: %v)\n", err)
			failCount++
			continue
		}

		// Parse SVG with oksvg
		icon, err := oksvg.ReadIconStream(strings.NewReader(string(graySVG)), oksvg.StrictErrorMode)
		if err != nil {
			fmt.Printf("FAILED (SVG parse error: %v)\n", err)
			failCount++
			continue
		}

		// Rasterize to 128x128 PNG
		size := 128
		icon.SetTarget(0, 0, float64(size), float64(size))
		rgba := image.NewRGBA(image.Rect(0, 0, size, size))

		// Fill with transparent background
		for y := 0; y < size; y++ {
			for x := 0; x < size; x++ {
				rgba.Set(x, y, color.RGBA{R: 255, G: 255, B: 255, A: 0})
			}
		}

		icon.Draw(rasterx.NewDasher(size, size, rasterx.NewScannerGV(size, size, rgba, rgba.Bounds())), 1.0)

		// Convert to grayscale
		grayImg := convertToGrayscale(rgba)

		// Save as PNG
		outFile, err := os.Create(pngPath)
		if err != nil {
			fmt.Printf("FAILED (create error: %v)\n", err)
			failCount++
			continue
		}

		if err := png.Encode(outFile, grayImg); err != nil {
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
