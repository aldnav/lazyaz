package main

import (
	"bytes"
	"fmt"
	"strings"
	"text/tabwriter"

	"github.com/rivo/tview"
)

func GetHotkeyView() *tview.TextView {
	var helpTextBuilder strings.Builder
	helpTextBuilder.WriteString("\n\n[::b]Hotkeys[::-]\n\n")

	var buf bytes.Buffer
	w := tabwriter.NewWriter(&buf, 0, 0, 2, ' ', 0)
	fmt.Fprintln(w, "]\tMove to next page")
	fmt.Fprintln(w, "[\tMove to previous page")
	fmt.Fprintln(w, "\\ + ↑/↓\tView filters")
	fmt.Fprintln(w, "/\tBasic search")
	fmt.Fprintln(w, "Enter\tView/hide details")
	fmt.Fprintln(w, "D\tExpand detail panel")
	fmt.Fprintln(w, "Q\tClose details panel")
	fmt.Fprintln(w, " \tClose hotkeys")
	fmt.Fprintln(w, "R\tRefresh")
	fmt.Fprintln(w, "CTRL+K\tToggle hotkeys")
	fmt.Fprintln(w, "ESC\tExit application")
	w.Flush()

	helpTextBuilder.WriteString(buf.String())

	return tview.NewTextView().
		SetText(helpTextBuilder.String()).
		SetTextAlign(tview.AlignLeft).
		SetDynamicColors(true)
}
