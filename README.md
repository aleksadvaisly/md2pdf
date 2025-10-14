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

Clone and install:

```sh
git clone https://github.com/aleksadvaisly/md2pdf.git
cd md2pdf
make install
```

This installs the binary to `~/.local/bin`. Make sure this directory is in your `PATH`.

## Usage

```sh
# Basic usage
md2pdf input.md

# Specify output file
md2pdf input.md output.pdf

# With flags
md2pdf -i input.md -o output.pdf

# Don't treat --- as page break
md2pdf --no-new-page input.md output.pdf
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
md2pdf -i input.md -o output.pdf --font dejavu_sans
```

## License

Distributed under the MIT License. See [LICENSE](./LICENSE) for details.
