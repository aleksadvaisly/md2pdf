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
	input := "**Block**  \n- [ ] one\n- [ ] two\n"
	expected := "**Block**  \n\n- [ ] one\n- [ ] two\n"
	result := ensureCheckboxListSpacing([]byte(input))
	if string(result) != expected {
		t.Fatalf("expected %q got %q", expected, string(result))
	}

	alreadySeparated := "**Block**\n\n- [ ] one\n"
	out := ensureCheckboxListSpacing([]byte(alreadySeparated))
	if string(out) != alreadySeparated {
		t.Fatalf("expected unchanged content")
	}
}
