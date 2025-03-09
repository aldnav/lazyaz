package main

import (
	"bytes"
	"fmt"
	"log"
	"os"
	"strings"
	"text/tabwriter"

	"github.com/aldnav/lazyaz/pkg/azuredevops"
	"github.com/gdamore/tcell/v2"
	strip "github.com/grokify/html-strip-tags-go"
	"github.com/microcosm-cc/bluemonday"
	"github.com/rivo/tview"
)

const tableData = `ID|Work Item Type|Created On|Assigned To|State|Title
0|Fetching work items...|-|-|-|-`

var p = bluemonday.UGCPolicy()

func workItemsToTableData(workItems []azuredevops.WorkItem) string {
	tableData := "ID|Work Item Type|Created On|Assigned To|State|Title\n"
	for i, workItem := range workItems {
		tableData += fmt.Sprintf("%d|%s|%s|%s|%s|%s", workItem.ID, workItem.WorkItemType, workItem.CreatedDate, workItem.AssignedTo, workItem.State, workItem.Title)
		// Add newline only if it's not the last item
		if i < len(workItems)-1 {
			tableData += "\n"
		}
	}
	return tableData
}

func redrawTable(table *tview.Table, workItems []azuredevops.WorkItem) {
	table.Clear()
	tableData := workItemsToTableData(workItems)
	for row, line := range strings.Split(tableData, "\n") {
		for column, cell := range strings.Split(line, "|") {
			color := tcell.ColorGray
			// Row 0 is the header
			if row > 0 {
				color = tcell.ColorWhite
			}
			if column == 0 && row > 0 {
				color = tcell.ColorPink
			}
			align := tview.AlignLeft
			tableCell := tview.NewTableCell(cell).
				SetTextColor(color).
				SetAlign(align).
				SetSelectable(row != 0)
			// Select expandable columns
			if column == 5 {
				tableCell.SetExpansion(1)
			}
			table.SetCell(row, column, tableCell)
		}
	}
}

func workItemToDetailsData(workItem *azuredevops.WorkItem) string {
	var buf bytes.Buffer
	w := tabwriter.NewWriter(&buf, 0, 0, 2, ' ', tabwriter.TabIndent)

	// Build header section
	fmt.Fprintf(w, "Title\t%s\n", workItem.Title)
	fmt.Fprintf(w, "ID\t%d\n", workItem.ID)
	fmt.Fprintf(w, "Work Item Type\t%s\n", workItem.WorkItemType)
	fmt.Fprintf(w, "Assigned To\t%s\n", workItem.AssignedTo)
	fmt.Fprintf(w, "State\t%s\n", workItem.State)

	if workItem.Details != nil {
		fmt.Fprintf(w, "Area Path\t%s\n", workItem.Details.SystemAreaPath)
	} else {
		fmt.Fprintf(w, "Area Path\tLoading...\n")
	}

	fmt.Fprintf(w, "Iteration Path\t%s\n", workItem.IterationPath)

	if workItem.WorkItemType == "Bug" {
		if workItem.Details != nil {
			fmt.Fprintf(w, "Description\t\n%s\n\n", p.Sanitize(strip.StripTags(workItem.Details.ReproSteps)))
		} else {
			fmt.Fprintf(w, "Description\tLoading...\n")
		}
	} else {
		fmt.Fprintf(w, "Description\t\n%s\n\n", p.Sanitize(strip.StripTags(workItem.Description)))
	}

	// Add tags section
	fmt.Fprintf(w, "Tags\t")
	if len(workItem.Tags) > 0 {
		fmt.Fprintf(w, "%s\n", workItem.Tags)
	} else {
		fmt.Fprintf(w, "[]\n")
	}

	fmt.Fprintf(w, "Created On\t%s\n", workItem.CreatedDate)
	fmt.Fprintf(w, "Created By\t%s\n", workItem.CreatedBy)
	fmt.Fprintf(w, "Changed On\t%s\n", workItem.ChangedDate)
	fmt.Fprintf(w, "Changed By\t%s\n", workItem.ChangedBy)

	// TODO DO not commit
	if workItem.ID == 279547 {
		log.Printf("Work item %d: %+v", workItem.ID, workItem)
	}

	w.Flush()
	return buf.String()
}

// Cover returns the cover page.
func WorkItemsPage(nextSlide func()) (title string, content tview.Primitive) {
	log.SetOutput(os.Stderr)
	log.SetPrefix("[lazyaz] ")
	var workItems []azuredevops.WorkItem
	var currentIndex int
	var loadingWorkItemID int = -1

	table := tview.NewTable().
		SetFixed(1, 1).
		SetBorders(false).
		SetSelectable(true, false).
		SetSeparator(' ')

	// Set custom selection style - red text on black background
	table.SetSelectedStyle(tcell.StyleDefault.
		Foreground(tcell.ColorBlack).
		Background(tcell.ColorLimeGreen))

	// Set initial table data
	for row, line := range strings.Split(tableData, "\n") {
		for column, cell := range strings.Split(line, "|") {
			color := tcell.ColorGray
			// Row 0 is the header
			if row > 0 {
				color = tcell.ColorWhite
			}
			if column == 0 && row > 0 {
				color = tcell.ColorPink
			}
			align := tview.AlignLeft
			tableCell := tview.NewTableCell(cell).
				SetTextColor(color).
				SetAlign(align).
				SetSelectable(row != 0)
			// Select expandable columns
			if column == 5 {
				tableCell.SetExpansion(1)
			}
			table.SetCell(row, column, tableCell)
		}
	}

	// Create a details panel
	detailsPanel := tview.NewTextView().
		SetScrollable(true).
		SetWordWrap(true)
	detailsPanel.
		SetDynamicColors(true).
		SetBorder(true).
		SetTitle(" Work Item ")

	// Create a flex container
	mainFlex := tview.NewFlex().
		SetDirection(tview.FlexColumn).
		AddItem(table, 0, 1, true)
		// Details panel not added initially

	// Variable to track if details are visible
	detailsVisible := false

	// displayDetails := func() {
	// 	fmt.Fprint(detailsPanel, detailsData)
	// }

	displayWorkItemDetails := func(workItems []azuredevops.WorkItem, index int) {
		if index >= 0 && index < len(workItems) {
			// Display available data immediately
			currentWorkItem := workItems[index]
			details := workItemToDetailsData(&currentWorkItem)
			detailsPanel.SetText(details)

			if currentWorkItem.Details == nil {
				loadingWorkItemID = currentWorkItem.ID
				// Fetch details in a goroutine
				go func() {
					// Capture the ID for comparison later
					requestedID := currentWorkItem.ID
					workItems[index].GetMoreWorkItemDetails()

					// Update UI on the main thread when done
					app.QueueUpdateDraw(func() {
						// Refresh with complete details
						details := workItemToDetailsData(&workItems[index])
						if requestedID == loadingWorkItemID {
							detailsPanel.SetText(details)
						}
					})
				}()
			}
		}
	}
	displayCurrentWorkItemDetails := func() {
		displayWorkItemDetails(workItems, currentIndex)
	}

	// When the table highlight is changed
	table.SetSelectionChangedFunc(func(row, column int) {
		currentIndex = row - 1
		if detailsVisible && currentIndex >= 0 {
			loadingWorkItemID = workItems[currentIndex].ID
			detailsPanel.SetText("")
			displayWorkItemDetails(workItems, currentIndex)
		}
	})

	// Add input capture for toggling details panel
	mainFlex.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		// Handle tab key to toggle details panel
		if event.Key() == tcell.KeyTab && len(workItems) > 0 {
			if detailsVisible {
				// Hide details
				mainFlex.RemoveItem(detailsPanel)
				detailsVisible = false
				detailsPanel.SetText("")
			} else {
				// Show details
				mainFlex.AddItem(detailsPanel, 0, 1, false)
				detailsVisible = true
				displayCurrentWorkItemDetails()
			}
			return nil
		}
		return event
	})

	// Integrate with Azure DevOps
	go func() {
		config, err := azuredevops.NewConfig()
		if err != nil {
			log.Printf("Configuration error: %v", err)
		}
		client := azuredevops.NewClient(config)
		workItems, err = client.GetWorkItemsAssignedToUser()
		if err != nil {
			log.Printf("Error fetching work items: %v", err)
		}
		if len(workItems) > 0 {
			app.QueueUpdateDraw(func() {
				redrawTable(table, workItems)
			})
		} else {
			app.QueueUpdateDraw(func() {
				table.SetCell(0, 0, tview.NewTableCell("No work items found").
					SetTextColor(tcell.ColorRed).
					SetAlign(tview.AlignCenter))
			})
		}
	}()

	return "Work Items", mainFlex
}
