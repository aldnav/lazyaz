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

func AttachWorkItemExtensions(actionsPanel *tview.Flex, table *tview.Table, workItems *[]azuredevops.WorkItem) {
	for _, extension := range ExtRegistry.GetFor("workitems") {
		buttonWidth := len(extension.Name) + 6
		button := tview.NewButton(extension.Name)
		button.
			SetSelectedFunc(func() {
				// Get the selected work item
				row, _ := table.GetSelection()
				idx := row - 1
				if idx >= 0 && idx < len(*workItems) {
					selectedWorkItem := (*workItems)[idx]
					// Get the extension ID from the registry
					for id, ext := range ExtRegistry.Extensions {
						if ext.Name == extension.Name {
							// Call the extension's entry point with the work item
							// Ignore result for now
							_, err := extension.EntryPoint(id).(func(interface{}) (string, error))(selectedWorkItem)
							if err != nil {
								button.SetBorderColor(tcell.ColorRed)
								// log.Printf("Error executing extension %s: %v", extension.Name, err)
							} else {
								button.SetBorderColor(tcell.ColorGreen)
								// log.Printf("Extension %s executed successfully: %s", extension.Name, result)
							}
							break
						}
					}
				}
			})
		actionsPanel.AddItem(button, buttonWidth, 0, false)
	}
}
