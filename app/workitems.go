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
		tableData += fmt.Sprintf("%d|%s|%s|%s|%s|%s", workItem.ID, workItem.WorkItemType, workItem.CreatedDate, workItem.AssignedTo, workItem.State, strings.ReplaceAll(workItem.Title, "|", ""))
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
		// Only access workItems for non-header rows
		var curWorkItem azuredevops.WorkItem
		if row > 0 && row-1 < len(workItems) {
			curWorkItem = workItems[row-1]
		}
		for column, cell := range strings.Split(line, "|") {
			color := tcell.ColorWhite
			// color := tcell.ColorGray
			// Row 0 is the header
			// if row > 0 {
			// 	color = tcell.ColorWhite
			// }
			if column == 0 && row > 0 {
				color = tcell.ColorRed
			}
			// State column
			if column == 4 {
				if cell == "To Do" || cell == "New" || cell == "Code Review" || cell == "Development" {
					color = tcell.ColorBlue
				} else if cell == "In Progress" {
					color = tcell.ColorYellow
				} else if cell == "Done" || cell == "ICR" || cell == "Closed" {
					color = tcell.ColorGreen
				} else if strings.Contains(cell, "Pending") || strings.Contains(cell, "Awaiting Decision") || cell == "On Hold" || cell == "Acceptance" {
					color = tcell.ColorOrange
				} else if cell == "Ready for test" || cell == "Test" {
					color = tcell.ColorRed
				}
			}

			// "Assigned to" column
			if column == 3 && row > 0 {
				if row-1 < len(workItems) && curWorkItem.IsAssignedToUser(activeUser) {
					color = tcell.ColorLimeGreen
				} else {
					color = tcell.ColorWhite
				}
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
	data = strings.ReplaceAll(data, "\u00A0", " ")
	data = strings.ReplaceAll(data, "&nbsp;", " ")
	return strip.StripTags(data)
}

func isSameAsUser(name string) bool {
	return name == activeUser.DisplayName
}

func workItemToDetailsData(workItem *azuredevops.WorkItem) string {
	workItemIsAssignedToUser := workItem.IsAssignedToUser(activeUser)
	var buf bytes.Buffer
	w := tabwriter.NewWriter(&buf, 0, 0, 2, ' ', tabwriter.TabIndent)

	// Build header section
	fmt.Fprintf(w, "Title\t%s\n", workItem.Title)
	fmt.Fprintf(w, "ID\t%d\n", workItem.ID)
	fmt.Fprintf(w, "URL\t%s\n", workItem.GetURL(_organization, _project))
	fmt.Fprintf(w, "Work Item Type\t%s\n", workItem.WorkItemType)
	if workItemIsAssignedToUser {
		fmt.Fprintf(w, "Assigned To\t[green]%s[white]\n", workItem.AssignedTo)
	} else {
		fmt.Fprintf(w, "Assigned To\t%s\n", workItem.AssignedTo)
	}
	fmt.Fprintf(w, "State\t%s\n", workItem.State)

	if workItem.Details != nil {
		fmt.Fprintf(w, "Area Path\t%s\n", workItem.Details.SystemAreaPath)
		fmt.Fprintf(w, "Priority\t%d\n", workItem.Details.Priority)
		fmt.Fprintf(w, "Severity\t%s\n", workItem.Details.Severity)
	} else {
		fmt.Fprintf(w, "Area Path\tLoading...\n")
		fmt.Fprintf(w, "Priority\tLoading...\n")
		fmt.Fprintf(w, "Severity\tLoading...\n")
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
	if isSameAsUser(workItem.CreatedBy) {
		fmt.Fprintf(w, "Created By\t[green]%s[white]\n", workItem.CreatedBy)
	} else {
		fmt.Fprintf(w, "Created By\t%s\n", workItem.CreatedBy)
	}
	fmt.Fprintf(w, "Changed On\t%s\n", workItem.ChangedDate)
	if isSameAsUser(workItem.ChangedBy) {
		fmt.Fprintf(w, "Changed By\t[green]%s[white]\n", workItem.ChangedBy)
	} else {
		fmt.Fprintf(w, "Changed By\t%s\n", workItem.ChangedBy)
	}

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
				if isSameAsUser(pr.Author) {
					fmt.Fprintf(w, "\t  Author\t[green]%s[white]\n", pr.Author)
				} else {
					fmt.Fprintf(w, "\t  Author\t%s\n", pr.Author)
				}
				fmt.Fprintf(w, "\t  URL\t%s\n", pr.GetURL())
				fmt.Fprintf(w, "\t  Created Date\t%s\n", pr.CreatedDate)
				if isSameAsUser(pr.ClosedBy) {
					fmt.Fprintf(w, "\t  Closed By\t[green]%s[white]\n", pr.ClosedBy)
				} else {
					fmt.Fprintf(w, "\t  Closed By\t%s\n", pr.ClosedBy)
				}
				fmt.Fprintf(w, "\t  Closed Date\t%s\n", pr.ClosedDate)
				fmt.Fprintf(w, "\t  From branch\t%s\n", pr.SourceRefName)
				fmt.Fprintf(w, "\t  To branch\t%s\n", pr.TargetRefName)
				votesInfo := pr.GetVotesInfo()
				fmt.Fprintf(w, "\t  Reviewers\n")
				for _, vote := range votesInfo {
					color := "[white]"
					if vote.Description == "approved" || vote.Description == "approved with suggestions" {
						color = "[green]"
					} else if vote.Description == "waiting for author" {
						color = "[yellow]"
					} else if vote.Description == "rejected" {
						color = "[red]"
					}
					fmt.Fprintf(w, "\t  \t%s\t%s%s[white]\n", vote.Reviewer, color, vote.Description)
				}
			}
		}
	}

	w.Flush()
	return buf.String()
}

func WorkItemsPage(nextSlide func()) (title string, content tview.Primitive) {
	log.SetOutput(os.Stderr)
	log.SetPrefix("[lazyaz] ")
	var workItems []azuredevops.WorkItem
	var currentIndex int
	var loadingWorkItemID int = -1
	// var client *azuredevops.Client
	var searchText string
	var previousSearchText string
	workItemFilter := "me"

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
	// Actions specific for work items
	actionsPanel := tview.NewFlex().
		SetDirection(tview.FlexColumn)
	dropdown := tview.NewDropDown().
		// SetLabel("Select an option (hit Enter): ").
		SetFieldBackgroundColor(tcell.ColorBlack).
		SetFieldTextColor(tcell.ColorWhite).
		SetListStyles(
			tcell.StyleDefault.
				Background(tcell.ColorBlack).
				Foreground(tcell.ColorWhite),
			tcell.StyleDefault.
				Background(tcell.ColorYellow).
				Foreground(tcell.ColorBlack),
		).
		// TODO "@Follows" and "@Mentions" are only working for web portal
		// https://learn.microsoft.com/en-us/azure/devops/boards/queries/query-operators-variables?view=azure-devops#query-macros-or-variables
		SetOptions([]string{"Assigned to me", "Was ever assigned to me", "All"}, nil)
	dropdown.SetCurrentOption(0)
	actionsPanel.AddItem(dropdown, 0, 1, false)
	searchStatus := tview.NewTextView().SetText("").SetTextAlign(tview.AlignRight)
	actionsPanel.AddItem(searchStatus, 0, 1, false)

	// Variable to track if details are visible
	detailsVisible := false

	// Function to highlight the current match - declare before use
	highlightMatch := func() {
		if currentMatchIndex >= 0 && currentMatchIndex < len(searchMatches) {
			match := searchMatches[currentMatchIndex]

			// Select the cell with the match
			table.Select(match.row, match.col)

			// Update the search input to show current match position - directly update instead of using QueueUpdateDraw
			searchStatus.SetLabel(fmt.Sprintf("Match %d/%d",
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
					searchStatus.SetLabel("")
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
					searchStatus.SetLabel("No matches!")
				}
				closeSearch()
			} else if key == tcell.KeyEscape {
				// Exit search mode on Escape
				closeSearch()

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
		if currentIndex < 0 {
			currentIndex = 0
		}
		if detailsVisible {
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

	toggleDetailsPanel := func() {
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
	}

	// Handle table enter key
	table.SetSelectedFunc(func(row, column int) {
		toggleDetailsPanel()
	})

	// Handle work item filter dropdown selection
	dropdown.SetSelectedFunc(func(text string, index int) {
		app.SetFocus(table)
		var potentialWorkItemFilter string
		switch text {
		case "Assigned to me":
			potentialWorkItemFilter = "me"
		case "Was ever assigned to me":
			potentialWorkItemFilter = "was-ever-me"
		case "All":
			potentialWorkItemFilter = "all"
		}
		if potentialWorkItemFilter != workItemFilter {
			workItemFilter = potentialWorkItemFilter
		} else {
			return
		}

		// Reset search variables
		searchText = ""
		previousSearchText = ""
		searchMode = false
		searchStatus.SetLabel("")
		searchMatches = nil
		currentMatchIndex = -1

		// Refresh the work items
		go func() {
			dropdown.SetLabel("Fetching ")
			var err error
			workItems, err = client.GetWorkItemsForFilter(workItemFilter)
			if err != nil {
				log.Printf("Error fetching work items: %v", err)
			}
			if len(workItems) > 0 {
				app.QueueUpdateDraw(func() {
					dropdown.SetLabel("")
					// Reset the index
					currentIndex = 0
					redrawTable(table, workItems)
					// Close the details panel
					closeDetailPanel()
					app.SetFocus(table)
					table.Select(0, 0)
				})
			} else {
				app.QueueUpdateDraw(func() {
					dropdown.SetLabel("")
					table.SetCell(0, 0, tview.NewTableCell("No work items found").
						SetTextColor(tcell.ColorRed).
						SetAlign(tview.AlignCenter))
					app.SetFocus(table)
				})
			}
		}()
	})

	mainWindow.AddItem(mainFlex, 0, 1, true)
	mainWindow.AddItem(actionsPanel, 1, 1, false)
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

		// Handle 'q' key to close details panel
		if activePanel == "details" && (event.Rune() == 'q') {
			closeDetailPanel()
			return nil
		}

		// Handle '\' key to activate dropdown
		if event.Rune() == '\\' {
			app.SetFocus(dropdown)
			return nil
		}
		return event
	})

	// Integrate with Azure DevOps
	go func() {
		var err error
		workItems, err = client.GetWorkItemsForFilter(workItemFilter)
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
