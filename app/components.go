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
	fmt.Fprintln(w, "CTRL+K\tToggle hotkeys")
	fmt.Fprintln(w, "ESC\tExit application")
	w.Flush()

	helpTextBuilder.WriteString(buf.String())

	return tview.NewTextView().
		SetText(helpTextBuilder.String()).
		SetTextAlign(tview.AlignLeft).
		SetDynamicColors(true)
}

func QuickActionsPane() tview.Primitive {
	form := tview.NewForm().
		AddInputField("Search", "", 20, nil, nil)
	form.SetBorder(false).
		SetTitleAlign(tview.AlignCenter)
	form.SetButtonBackgroundColor(tview.Styles.ContrastBackgroundColor)
	form.SetHorizontal(true)
	modal := NewFormModal(form).
		SetText("").
		SetBackgroundColor(tview.Styles.PrimitiveBackgroundColor).
		AddButtons([]string{"Search"}).
		SetDoneFunc(func(buttonIndex int, buttonLabel string) {
			// if buttonLabel == "Quit" {
			// 	app.Stop()
			// }
			if buttonLabel == "Search" {
				// TODO Delegate from the calling function
			}
		})
	return modal
}
