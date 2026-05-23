package diff

import (
	"fmt"
	"io"
	"strings"
)

// RenderText writes a human-readable diff summary to w.
func RenderText(w io.Writer, service string, changes []Change) {
	if len(changes) == 0 {
		fmt.Fprintf(w, "[%s] no drift detected\n", service)
		return
	}
	fmt.Fprintf(w, "[%s] %d change(s) detected:\n", service, len(changes))
	for _, c := range changes {
		fmt.Fprintf(w, "  %-10s %s\n", strings.ToUpper(string(c.Kind)), c.String())
	}
}

// RenderMarkdown writes a Markdown-formatted diff block to w.
func RenderMarkdown(w io.Writer, service string, changes []Change) {
	if len(changes) == 0 {
		fmt.Fprintf(w, "**%s**: no drift detected\n", service)
		return
	}
	fmt.Fprintf(w, "### %s — %d change(s)\n\n", service, len(changes))
	fmt.Fprintln(w, "| Field | Kind | Old | New |")
	fmt.Fprintln(w, "|-------|------|-----|-----|")
	for _, c := range changes {
		fmt.Fprintf(w, "| %s | %s | %v | %v |\n",
			c.Field, c.Kind, c.OldValue, c.NewValue)
	}
}
