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
var _project string
var _organization string

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
			color := tcell.ColorWhite
			// color := tcell.ColorGray
			// Row 0 is the header
			// if row > 0 {
			// 	color = tcell.ColorWhite
			// }
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

// Replaces common html tags, then strip other tags, and sanitize
func normalizeDataString(data string) string {
	data = strings.ReplaceAll(data, "<br>", "\n")
	data = strings.ReplaceAll(data, "<br />", "\n")
	data = strings.ReplaceAll(data, "<br/>", "\n")
	data = strings.ReplaceAll(data, "<br>", "\n")
	data = strings.ReplaceAll(data, "<br />", "\n")
	data = strings.ReplaceAll(data, "<br/>", "\n")
	data = strings.ReplaceAll(data, "&#39;", "'")
	data = strings.ReplaceAll(data, "&#34;", "\"")
	data = strings.ReplaceAll(data, "&quot;", "\"")
	data = strings.ReplaceAll(data, "&apos;", "'")
	data = strings.ReplaceAll(data, "&amp;", "&")
	data = strings.ReplaceAll(data, "\n\n", "\n")
	return p.Sanitize(strip.StripTags(data))
}

func workItemToDetailsData(workItem *azuredevops.WorkItem) string {
	var buf bytes.Buffer
	w := tabwriter.NewWriter(&buf, 0, 0, 2, ' ', tabwriter.TabIndent)

	// Build header section
	fmt.Fprintf(w, "Title\t%s\n", workItem.Title)
	fmt.Fprintf(w, "ID\t%d\n", workItem.ID)
	fmt.Fprintf(w, "URL\t%s\n", workItem.GetURL(_organization, _project))
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
			fmt.Fprintf(w, "Description\t\n%s\n\n", normalizeDataString(workItem.Details.ReproSteps))
		} else {
			fmt.Fprintf(w, "Description\tLoading...\n")
		}
	} else {
		fmt.Fprintf(w, "Description\t\n%s\n\n", normalizeDataString(workItem.Description))
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

	if workItem.Details != nil {
		fmt.Fprintf(w, "\nAdditional details\n")
		fmt.Fprintf(w, "Acceptance Criteria\t\n%s\n\n", normalizeDataString(workItem.Details.AcceptanceCriteria))
		fmt.Fprintf(w, "Board Column\t%s\n", workItem.Details.BoardColumn)
		fmt.Fprintf(w, "Comment Count\t%d\n", workItem.Details.CommentCount)
		latestComment := normalizeDataString(workItem.Details.LatestComment)
		if latestComment == "" {
			latestComment = "[red][Cannot fetch latest comment][white]"
		}
		fmt.Fprintf(w, "Latest Comment\t\n%s\n\n", latestComment)
		// fmt.Fprintf(w, "PR Refs\t%s\n", strings.Join(workItem.GetPRs(), ", "))
		if len(workItem.PRDetails) > 0 {
			fmt.Fprintf(w, "Associated Pull Requests\n\n")
			for _, pr := range workItem.PRDetails {
				fmt.Fprintf(w, "\t- %d %s\n", pr.ID, pr.Title)
				if pr.IsDraft {
					fmt.Fprintf(w, "\t  Is Draft\t[yellow][Yes][white]\n")
				}
				statusColor := "[white]"
				if pr.Status == "completed" {
					statusColor = "[green]"
				} else if pr.Status == "abandoned" {
					statusColor = "[red]"
				} else if pr.Status == "active" {
					statusColor = "[yellow]"
				}
				fmt.Fprintf(w, "\t  Status\t%s%s[white]\n", statusColor, pr.Status)
				fmt.Fprintf(w, "\t  Author\t%s\n", pr.Author)
				fmt.Fprintf(w, "\t  URL\t%s\n", pr.RepositoryURL)
				fmt.Fprintf(w, "\t  Created Date\t%s\n", pr.CreatedDate)
				fmt.Fprintf(w, "\t  Closed By\t%s\n", pr.ClosedBy)
				fmt.Fprintf(w, "\t  Closed Date\t%s\n", pr.ClosedDate)
				fmt.Fprintf(w, "\t  From branch\t%s\n", pr.SourceRefName)
				fmt.Fprintf(w, "\t  To branch\t%s\n", pr.TargetRefName)
				fmt.Fprintf(w, "\t  Reviewers\t%s\n", strings.Join(pr.Reviewers, ", "))
			}
		}
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
	var client *azuredevops.Client
	var searchText string
	var previousSearchText string

	// Add search-related variables
	var searchMode bool = false
	var searchInput *tview.InputField
	var searchMatches []struct{ row, col int }
	var currentMatchIndex int = -1

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
			color := tcell.ColorWhite
			// color := tcell.ColorGray
			// // Row 0 is the header
			// if row > 0 {
			// 	color = tcell.ColorWhite
			// }
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
	mainWindow := tview.NewFlex().
		SetDirection(tview.FlexRow)

	mainFlex := tview.NewFlex().
		SetDirection(tview.FlexColumn).
		AddItem(table, 0, 1, true)
		// Details panel not added initially

	// Variable to track if details are visible
	detailsVisible := false

	// Function to clear highlights - declare before use
	clearHighlights := func() {
		// Reset all cell styles to default
		rowCount := table.GetRowCount()
		colCount := table.GetColumnCount()

		for row := 0; row < rowCount; row++ {
			for col := 0; col < colCount; col++ {
				cell := table.GetCell(row, col)
				if cell != nil {
					// Restore original colors based on your table's styling logic
					color := tcell.ColorWhite
					if col == 0 && row > 0 {
						color = tcell.ColorPink
					}
					cell.SetTextColor(color)
				}
			}
		}
	}

	// Function to highlight the current match - declare before use
	highlightMatch := func() {
		if currentMatchIndex >= 0 && currentMatchIndex < len(searchMatches) {
			// Clear previous highlights
			clearHighlights()

			match := searchMatches[currentMatchIndex]

			// Select the cell with the match
			table.Select(match.row, match.col)

			// Update the search input to show current match position - directly update instead of using QueueUpdateDraw
			searchInput.SetLabel(fmt.Sprintf("Match %d/%d /",
				currentMatchIndex+1, len(searchMatches)))
		}
	}

	// Create a search input field
	closeSearch := func() {
		searchMode = false
		mainWindow.RemoveItem(searchInput)
		app.SetFocus(table)
	}

	searchInput = tview.NewInputField().
		SetLabel("/").
		SetFieldWidth(30).
		SetFieldTextColor(tcell.ColorGreen).
		SetFieldBackgroundColor(tcell.ColorBlack).
		SetDoneFunc(func(key tcell.Key) {
			if key == tcell.KeyEnter {
				// Perform search when Enter is pressed
				searchText = strings.TrimSpace(searchInput.GetText())
				if searchText == "" {
					// Exit search mode if search text is empty
					closeSearch()
					return
				}

				if searchText != previousSearchText {
					// Clear previous matches
					searchMatches = nil
					currentMatchIndex = -1
					previousSearchText = searchText

					// Search for matches in the table
					rowCount := table.GetRowCount()
					colCount := table.GetColumnCount()
					searchTextLower := strings.ToLower(searchText)

					for row := 1; row < rowCount; row++ {
						for col := 0; col < colCount; col++ {
							cell := table.GetCell(row, col)
							if cell != nil {
								cellText := cell.Text
								if strings.Contains(strings.ToLower(cellText), searchTextLower) {
									searchMatches = append(searchMatches, struct{ row, col int }{row, col})
								}
							}
						}
					}

					if len(searchMatches) > 0 {
						// Highlight the first match
						currentMatchIndex = 0
					}
				} else {
					// Highlight next match (if any)
					currentMatchIndex += 1
				}

				if currentMatchIndex >= len(searchMatches) {
					currentMatchIndex = 0 // Wrap around
				}

				// Handle search results
				if len(searchMatches) > 0 {
					highlightMatch()
				} else {
					// Show "No matches" message - directly update the label instead of using QueueUpdateDraw
					searchInput.SetLabel("No matches. /")
				}
				closeSearch()
			} else if key == tcell.KeyEscape {
				// Exit search mode on Escape
				closeSearch()

				// Clear any highlights
				clearHighlights()
			}
		})

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
					workItems[index].GetPRDetails(client)
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

	closeDetailPanel := func() {
		// Release active panel for keyboard context
		activePanel = ""
		// Hide details
		mainFlex.RemoveItem(detailsPanel)
		detailsVisible = false
		detailsPanel.SetText("")
	}

	mainWindow.AddItem(mainFlex, 0, 1, true)
	// Add input capture for toggling details panel
	mainWindow.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		// Handle search mode activation with "/"
		if event.Rune() == '/' {
			if !searchMode {
				searchMode = true
				searchInput.SetText(searchText)
				searchInput.SetLabel("/")
				mainWindow.AddItem(searchInput, 1, 0, false)
				app.SetFocus(searchInput)
				return nil
			} else {
				closeSearch()
				return nil
			}
		}

		// Handle tab key to toggle details panel
		if event.Key() == tcell.KeyTab && len(workItems) > 0 {
			if detailsVisible {
				closeDetailPanel()
			} else {
				// Set active panel for keyboard context
				activePanel = "details"
				// Show details
				mainFlex.AddItem(detailsPanel, 0, 1, false)
				detailsVisible = true
				displayCurrentWorkItemDetails()
			}
			return nil
		} else if activePanel == "details" && (event.Rune() == 'q') {
			closeDetailPanel()
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
		_organization = config.Organization
		_project = config.Project
		client = azuredevops.NewClient(config)
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

	return "Work Items", mainWindow
}
