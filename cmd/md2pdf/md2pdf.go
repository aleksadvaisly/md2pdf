package main

import (
	"errors"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"regexp"
	"runtime"
	"strings"
	"time"

	"github.com/gomarkdown/markdown/parser"
	"github.com/solworktech/md2pdf/v2"
	flag "github.com/spf13/pflag"
	"golang.org/x/exp/slices"
)

var input = flag.StringP("input", "i", "", "Input filename, dir consisting of .md|.markdown files or HTTP(s) URL; default is os.Stdin")
var output = flag.StringP("output", "o", "", "Output PDF filename; required")
var pathToSyntaxFiles = flag.StringP("syntax-files", "s", "", "Path to github.com/jessp01/gohighlight/syntax_files")
var title = flag.String("title", "", "Presentation title")
var author = flag.String("author", "", "Author's name; used if -footer is passed")
var fontFamily = flag.String("font-family", "", "System font family [Times | Helvetica | Courier]")
var presetFont = flag.String("font", "", "Predefined Unicode font [dejavu_sans | dejavu_serif | noto_sans | roboto | eb_garamond | merriweather | source_serif] (default: source_serif)")
var themeArg = flag.String("theme", "light", "[light | dark | /path/to/custom/theme.json]")
var noNewPage = flag.Bool("no-new-page", false, "Don't interpret HR (---) as page break")
var keepNumbering = flag.Bool("keep-numbering", false, "Preserve continuous list numbering across headers (default: reset to 1)")
var printFooter = flag.Bool("with-footer", false, "Print doc footer (<author>  <title>  <page number>)")
var generateTOC = flag.Bool("generate-toc", false, "Auto Generate Table of Contents (TOC)")
var pageSize = flag.String("page-size", "A4", "[A3 | A4 | A5]")
var orientation = flag.String("orientation", "portrait", "[portrait | landscape]")
var logFile = flag.String("log-file", "", "Path to log file")
var help = flag.Bool("help", false, "Show usage message")
var ver = flag.Bool("version", false, "Print version and build info")
var version = "dev"
var commit = "none"
var date = "unknown"
var _, fileName, fileLine, ok = runtime.Caller(0)

var opts []mdtopdf.RenderOption

func processRemoteInputFile(url string) ([]byte, error) {
	client := &http.Client{
		Timeout: 30 * time.Second,
	}
	resp, err := client.Get(url)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode != 200 {
		return nil, errors.New("Received non 200 response code: " + fmt.Sprintf("HTTP %d", resp.StatusCode))
	}
	content, rerr := io.ReadAll(resp.Body)
	return content, rerr
}

func glob(dir string, validExts []string) ([]string, error) {
	files := []string{}
	err := filepath.Walk(dir, func(path string, f os.FileInfo, err error) error {
		if slices.Contains(validExts, filepath.Ext(path)) {
			files = append(files, path)
		}
		return nil
	})

	return files, err
}

func loadPresetFont(fontName string) error {
	validFonts := map[string]bool{
		"dejavu_sans":  true,
		"dejavu_serif": true,
		"noto_sans":    true,
		"roboto":       true,
		"eb_garamond":  true,
		"merriweather": true,
		"source_serif": true,
	}

	if _, exists := validFonts[fontName]; !exists {
		return fmt.Errorf("unknown preset font: %s (available: dejavu_sans, dejavu_serif, noto_sans, roboto, eb_garamond, merriweather, source_serif)", fontName)
	}

	return nil
}

func main() {
	log.SetFlags(log.LstdFlags | log.Lshortfile)
	flag.Parse()

	if *input == "" && len(flag.Args()) > 0 {
		*input = flag.Args()[0]
	}

	if *help {
		usage("")
		return
	}

	if *ver {
		fmt.Printf("md2pdf %s\n", version)
		return
	}

	if *noNewPage {
		opts = append(opts, mdtopdf.IsHorizontalRuleNewPage(false))
	} else {
		opts = append(opts, mdtopdf.IsHorizontalRuleNewPage(true))
	}

	if *pathToSyntaxFiles != "" {
		opts = append(opts, mdtopdf.SetSyntaxHighlightBaseDir(*pathToSyntaxFiles))
	} else {
		if _, err := os.Stat("../../highlight/syntax_files"); err == nil {
			opts = append(opts, mdtopdf.SetSyntaxHighlightBaseDir("../../highlight/syntax_files"))
		} else if _, err := os.Stat("/usr/share/mdtopdf/syntax_files"); err == nil {
			opts = append(opts, mdtopdf.SetSyntaxHighlightBaseDir("/usr/share/mdtopdf/syntax_files"))
		}
	}

	// get text for PDF
	var content []byte
	var err error
	var inputBaseURL string
	if *input == "" {
		content, err = io.ReadAll(os.Stdin)
		if err != nil {
			log.Fatal(err)
		}
	} else {
		httpRegex := regexp.MustCompile("^http(s)?://")
		if httpRegex.Match([]byte(*input)) {
			content, err = processRemoteInputFile(*input)
			if err != nil {
				log.Fatal(err)
			}
			// get the base URL so we can adjust relative links and images
			inputBaseURL = strings.Replace(filepath.Dir(*input), ":/", "://", 1)
		} else {
			fileInfo, err := os.Stat(*input)
			if err != nil {
				log.Fatal(err)
			}

			if fileInfo.IsDir() {
				validExts := []string{".md", ".markdown"}
				files, err := glob(*input, validExts)
				if err != nil {
					log.Fatal(err)
				}
				for i, filePath := range files {
					fileContents, err := os.ReadFile(filePath)
					if err != nil {
						log.Fatal(err)
					}
					content = append(content, fileContents...)
					if i < len(files)-1 {
						content = append(content, []byte("---\n")...)
					}
				}
			} else {
				content, err = os.ReadFile(*input)
				if err != nil {
					log.Fatal(err)
				}
			}
		}
	}

	// Auto-generate output filename if not provided
	if *output == "" {
		if *input == "" {
			usage("Output PDF filename is required when reading from stdin")
		} else {
			httpRegex := regexp.MustCompile("^http(s)?://")
			if httpRegex.Match([]byte(*input)) {
				// For URLs, use the base filename from URL
				baseName := filepath.Base(*input)
				*output = strings.TrimSuffix(baseName, filepath.Ext(baseName)) + ".pdf"
			} else {
				fileInfo, err := os.Stat(*input)
				if err == nil && fileInfo.IsDir() {
					// For directories, use directory name
					*output = filepath.Base(*input) + ".pdf"
				} else {
					// For files, replace .md or .markdown extension with .pdf
					baseName := *input
					if strings.HasSuffix(baseName, ".md") {
						*output = strings.TrimSuffix(baseName, ".md") + ".pdf"
					} else if strings.HasSuffix(baseName, ".markdown") {
						*output = strings.TrimSuffix(baseName, ".markdown") + ".pdf"
					} else {
						*output = baseName + ".pdf"
					}
				}
			}
		}
	}

	theme := mdtopdf.LIGHT
	themeFile := ""
	if *themeArg == "dark" {
		theme = mdtopdf.DARK
	} else if _, err := os.Stat(*themeArg); err == nil {
		theme = mdtopdf.CUSTOM
		themeFile = *themeArg
	}

	if *presetFont == "" && *fontFamily == "" {
		*presetFont = "source_serif"
	}

	if *presetFont != "" {
		if *fontFamily != "" {
			log.Printf("Warning: Both --font and --font-family specified. --font takes priority.")
		}

		err := loadPresetFont(*presetFont)
		if err != nil {
			log.Fatalf("Failed to load preset font: %v", err)
		}
	}

	params := mdtopdf.PdfRendererParams{
		Orientation:     *orientation,
		Papersz:         *pageSize,
		PdfFile:         *output,
		TracerFile:      *logFile,
		Opts:            opts,
		Theme:           theme,
		CustomThemeFile: themeFile,
		DefaultFont:     *fontFamily,
		PresetFont:      *presetFont,
		KeepNumbering:   *keepNumbering,
	}

	pf := mdtopdf.NewPdfRenderer(params)

	if *generateTOC == true {
		headers, err := mdtopdf.GetTOCEntries(content)
		if err != nil {
			log.Fatal(err)
		}
		headerLinks := make(map[string]*int)
		for _, header := range headers {
			linkID := pf.Pdf.AddLink()
			headerLinks[header.Title] = &linkID

			// debug
			// log.Printf("Header: '%s' (Level %d) -> Link ID: %d\n",
			// header.Title, header.Level, linkID)
		}

		pf.SetTOCLinks(headerLinks)
		pf.Pdf.SetFont("Arial", "B", 24)

		// Add a table of contents with clickable links
		pf.Pdf.Cell(40, 10, "Table of Contents")
		pf.Pdf.Ln(30)

		for _, header := range headers {
			if linkPtr, exists := headerLinks[header.Title]; exists {
				link := *linkPtr
				pf.Pdf.SetFont("Arial", "", 12)
				pf.Pdf.SetTextColor(100, 149, 237)
				tr := pf.Pdf.UnicodeTranslatorFromDescriptor("")
				bulletChar := tr("•")
				indent := strings.Repeat("  ", header.Level-1)
				pf.Pdf.WriteLinkID(8, fmt.Sprintf("%s %s %s", indent, bulletChar, header.Title), link)
				pf.Pdf.Ln(15)
			}
		}
		pf.Pdf.AddPage()
	}

	if inputBaseURL != "" {
		pf.InputBaseURL = inputBaseURL
	}
	pf.Pdf.SetSubject(*title, true)
	pf.Pdf.SetTitle(*title, true)
	pf.Extensions = parser.NoIntraEmphasis | parser.Tables | parser.FencedCode | parser.Autolink | parser.Strikethrough | parser.SpaceHeadings | parser.HeadingIDs | parser.BackslashLineBreak | parser.DefinitionLists | parser.HardLineBreak | parser.OrderedListStart

	if *printFooter {
		pf.Pdf.SetFooterFunc(func() {
			pf.Pdf.SetFillColor(pf.BackgroundColor.Red, pf.BackgroundColor.Green, pf.BackgroundColor.Blue)
			// Position at 1.5 cm from bottom
			pf.Pdf.SetY(-15)
			// Arial italic 8
			pf.Pdf.SetFont("Arial", "I", 8)
			// Text color in gray
			pf.Pdf.SetTextColor(128, 128, 128)
			w, h, _ := pf.Pdf.PageSize(pf.Pdf.PageNo())
			// fmt.Printf("Width: %f, height: %f, unit: %s\n", w, h, u)
			pf.Pdf.SetX(4)
			pf.Pdf.CellFormat(0, 10, fmt.Sprintf("%s", *author), "", 0, "", true, 0, "")
			middle := w / 2
			if *orientation == "landscape" {
				middle = h / 2
			}
			pf.Pdf.SetX(middle - float64(len(*title)))
			pf.Pdf.CellFormat(0, 10, fmt.Sprintf("%s", *title), "", 0, "", true, 0, "")
			pf.Pdf.SetX(-40)
			pf.Pdf.CellFormat(0, 10, fmt.Sprintf("Page %d", pf.Pdf.PageNo()), "", 0, "", true, 0, "")
		})
	}

	err = pf.Process(content)
	if err != nil {
		fmt.Printf("error: %v\n", err)
	}
}

func usage(msg string) {
	fmt.Println(msg + "\n")
	fmt.Printf("Usage: %s (%s) [options]\n", filepath.Base(fileName), version)
	flag.PrintDefaults()
	os.Exit(0)
}
