package tui

import (
	"strings"

	"github.com/alecthomas/chroma/v2"
	"github.com/alecthomas/chroma/v2/formatters"
	"github.com/alecthomas/chroma/v2/lexers"
	"github.com/alecthomas/chroma/v2/styles"
	"github.com/charmbracelet/glamour"
)

var chromaStyle *chroma.Style

func init() {
	chromaStyle = styles.Get("dracula")
	if chromaStyle == nil {
		chromaStyle = styles.Get("monokai")
	}
}

func highlightShell(code string) string {
	if strings.TrimSpace(code) == "" {
		return code
	}

	lexer := lexers.Get("bash")
	if lexer == nil {
		lexer = lexers.Fallback
	}
	lexer = chroma.Coalesce(lexer)

	iterator, err := lexer.Tokenise(nil, code)
	if err != nil {
		return code
	}

	formatter := formatters.TTY16m
	var buf strings.Builder
	if err := formatter.Format(&buf, chromaStyle, iterator); err != nil {
		return code
	}

	return buf.String()
}

func renderMarkdown(text string) string {
	if !glamourEnabled() {
		return text
	}
	r, err := glamour.NewTermRenderer(
		glamour.WithAutoStyle(),
		glamour.WithWordWrap(80),
	)
	if err != nil {
		return text
	}
	out, err := r.Render(text)
	if err != nil {
		return text
	}
	return strings.TrimSpace(out)
}

func glamourEnabled() bool {
	return true
}
