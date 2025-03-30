package main

import (
	"bytes"
	"fmt"
	"strings"
	"text/tabwriter"

	"github.com/aldnav/lazyaz/pkg/azuredevops"
	"github.com/gdamore/tcell/v2"
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

func attachExtensionToPanel[T any](extension ExtensionConfig, actionsPanel *tview.Flex, table *tview.Table, items *[]T) {
	buttonWidth := len(extension.Name) + 6
	button := tview.NewButton(extension.Name)
	button.
		SetSelectedFunc(func() {
			// Get the selected item
			row, _ := table.GetSelection()
			idx := row - 1
			if idx >= 0 && idx < len(*items) {
				selectedItem := (*items)[idx]
				// Get the extension ID from the registry
				for id, ext := range ExtRegistry.Extensions {
					if ext.Name == extension.Name {
						// Call the extension's entry point with the selected item
						_, err := extension.EntryPoint(id).(func(interface{}) (string, error))(selectedItem)
						if err != nil {
							button.SetBorderColor(tcell.ColorRed)
						} else {
							button.SetBorderColor(tcell.ColorGreen)
						}
						break
					}
				}
			}
		})
	button.SetStyle(tcell.StyleDefault.Background(tcell.ColorWhite).Foreground(tcell.ColorBlack))
	button.SetActivatedStyle(tcell.StyleDefault.Background(tcell.ColorBlue).Foreground(tcell.ColorWhite))
	actionsPanel.AddItem(button, buttonWidth, 0, false)
	actionsPanel.AddItem(nil, 1, 1, false)
}

func AttachWorkItemExtensions(actionsPanel *tview.Flex, table *tview.Table, workItems *[]azuredevops.WorkItem) {
	for _, extension := range ExtRegistry.GetFor("workitems") {
		attachExtensionToPanel(extension, actionsPanel, table, workItems)
	}
}

func AttachPullRequestExtensions(actionsPanel *tview.Flex, table *tview.Table, pullRequests *[]azuredevops.PullRequestDetails) {
	for _, extension := range ExtRegistry.GetFor("pullrequests") {
		attachExtensionToPanel(extension, actionsPanel, table, pullRequests)
	}
}

func AttachPipelineRunExtensions(actionsPanel *tview.Flex, table *tview.Table, pipelines *[]azuredevops.PipelineRun) {
	for _, extension := range ExtRegistry.GetFor("pipelines") {
		attachExtensionToPanel(extension, actionsPanel, table, pipelines)
	}
}
