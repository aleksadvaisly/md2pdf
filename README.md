# Markdown to PDF

<img src="docs/small_create_with_ai.png" style="float: left; margin: 0 15px 15px 0;" width="150">

A command-line tool for converting Markdown files to PDF. Written in pure Go with no external dependencies.

This package uses:
- [gomarkdown](https://github.com/gomarkdown/markdown) for parsing Markdown
- [fpdf](https://codeberg.org/go-pdf/fpdf) for PDF generation

## Features

- Syntax highlighting for code blocks
- Unicode support with multiple fonts
- Tables, lists, links, and images
- Page control with horizontal rules

## Supported Markdown Elements

- Emphasised and strong text
- Headings 1-6
- Ordered and unordered lists
- Nested lists
- Images
- Tables
- Links
- Code blocks and inline code

## Installation

Download pre-built binaries from [releases](https://github.com/solworktech/md2pdf/releases), or install with Go:

```sh
go install github.com/solworktech/md2pdf/v2/cmd/md2pdf@latest
```

## Usage

```sh
# Basic usage
md2pdf input.md

# Specify output file
md2pdf input.md output.pdf

# With flags
md2pdf -i input.md -o output.pdf
```

Convert multiple files:
```sh
md2pdf -i /path/to/markdown/directory -o combined.pdf
```

## Building

With Make:
```sh
make build    # Build binary to bin/md2pdf
make install  # Install to ~/.local/bin
make test     # Run tests
make clean    # Remove build artifacts
```

With Go:
```sh
go build -o bin/md2pdf ./cmd/md2pdf
```

## Syntax Highlighting

Code blocks are automatically highlighted. To specify a language, annotate the code block:

    ```go
    func main() {
        fmt.Println("Hello")
    }
    ```

The annotation name must match the syntax file basename. See [testdata/syntax_highlighting.md](./testdata/syntax_highlighting.md) for examples.

## Table of Contents

Generate a clickable table of contents:

```sh
md2pdf --generate-toc input.md
```

## Fonts

Several Unicode fonts are included:

- `dejavu_sans` - sans-serif with excellent Unicode coverage
- `dejavu_serif` - serif with excellent Unicode coverage
- `noto_sans` - comprehensive Unicode support
- `roboto` - modern sans-serif
- `eb_garamond` - elegant serif (default)
- `merriweather` - readable serif
- `source_serif` - Adobe serif

Usage:
```sh
md2pdf --font dejavu_sans input.md
```

## Options

```
  -author string
        Author name (used in footer)
  -font string
        Font preset [dejavu_sans | dejavu_serif | noto_sans | roboto |
        eb_garamond | merriweather | source_serif] (default: eb_garamond)
  -generate-toc
        Generate table of contents
  -i string
        Input file, directory, or URL
  -o string
        Output PDF file (auto-generated if omitted)
  -orientation string
        Page orientation [portrait | landscape] (default: portrait)
  -page-size string
        Paper size [A3 | A4 | A5] (default: A4)
  -title string
        Document title
  -with-footer
        Print footer with author, title, and page number
  --debug
        Enable debug logging
```

## Example

```sh
md2pdf -i report.md \
       -o report.pdf \
       --title "Project Report" \
       --author "Your Name" \
       --font dejavu_serif \
       --with-footer
```

## Tests

Tests are in the `testdata` folder. Run with:

```sh
make test
```

Visual inspection of generated PDFs is recommended to verify output quality.

## Limitations

- HTML in Markdown is treated as code blocks and not rendered
- Strikethrough is not supported
- Definition lists are not supported
- Link titles (hover text) are not rendered

## License

Distributed under the MIT License. See [LICENSE](./LICENSE) for details.
