package mdtopdf

import (
	"testing"

	"github.com/gomarkdown/markdown"
	"github.com/gomarkdown/markdown/ast"
	"github.com/gomarkdown/markdown/parser"
)

func firstListItem(markdownSrc string) *ast.ListItem {
	p := parser.NewWithExtensions(parser.CommonExtensions | parser.AutoHeadingIDs | parser.HardLineBreak)
	doc := markdown.Parse([]byte(markdownSrc), p)

	var item *ast.ListItem
	ast.WalkFunc(doc, func(n ast.Node, entering bool) ast.WalkStatus {
		if !entering || item != nil {
			return ast.GoToNext
		}
		if li, ok := n.(*ast.ListItem); ok {
			item = li
			return ast.Terminate
		}
		return ast.GoToNext
	})

	return item
}

func firstTextContent(item *ast.ListItem) string {
	var text string
	ast.WalkFunc(item, func(n ast.Node, entering bool) ast.WalkStatus {
		if !entering || text != "" {
			return ast.GoToNext
		}
		if tn, ok := n.(*ast.Text); ok {
			text = string(tn.Literal)
			return ast.Terminate
		}
		return ast.GoToNext
	})
	return text
}

func TestStripCheckboxMarker(t *testing.T) {
	cases := []struct {
		name     string
		markdown string
		expected string
		symbol   string
		matched  bool
	}{
		{
			name:     "unchecked",
			markdown: "- [ ] Task\n",
			expected: "Task",
			symbol:   "☐",
			matched:  true,
		},
		{
			name:     "checked lower",
			markdown: "- [x] Done\n",
			expected: "Done",
			symbol:   "☑",
			matched:  true,
		},
		{
			name:     "checked upper",
			markdown: "- [X] Done\n",
			expected: "Done",
			symbol:   "☑",
			matched:  true,
		},
		{
			name:     "plain",
			markdown: "- Plain item\n",
			expected: "Plain item",
			symbol:   "",
			matched:  false,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			item := firstListItem(tc.markdown)
			if item == nil {
				t.Fatalf("expected list item for %q", tc.markdown)
			}
			sym, matched := stripCheckboxMarker(item)
			if matched != tc.matched {
				t.Fatalf("expected matched=%v got %v", tc.matched, matched)
			}
			if sym != tc.symbol {
				t.Fatalf("expected symbol %q got %q", tc.symbol, sym)
			}
			if got := firstTextContent(item); got != tc.expected {
				t.Fatalf("expected text %q got %q", tc.expected, got)
			}
		})
	}
}

func TestEnsureCheckboxListSpacing(t *testing.T) {
	cases := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "checkbox after forced break",
			input:    "**Block**  \n- [ ] one\n- [ ] two\n",
			expected: "**Block**  \n\n- [ ] one\n- [ ] two\n",
		},
		{
			name:     "numbered list after forced break",
			input:    "Intro line  \n1. First\n2. Second\n",
			expected: "Intro line  \n\n1. First\n2. Second\n",
		},
		{
			name:     "nested list remains intact",
			input:    "1. Parent\n   - Child\n",
			expected: "1. Parent\n   - Child\n",
		},
	}

	for _, tc := range cases {
		result := ensureCheckboxListSpacing([]byte(tc.input))
		if string(result) != tc.expected {
			t.Fatalf("%s: expected %q got %q", tc.name, tc.expected, string(result))
		}
	}
}
