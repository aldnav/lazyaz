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
	"github.com/rivo/tview"
)

const tableData = `ID|Work Item Type|Created On|Assigned To|State|Title
0|Fetching work items...|Loading...|Loading...|Loading...|Loading...`

var _project string
var _organization string

func workItemsToTableData(workItems []azuredevops.WorkItem) string {
	tableData := "ID|Work Item Type|Created On|Assigned To|State|Title\n"
	for i, workItem := range workItems {
		tableData += fmt.Sprintf("%d|%s|%s|%s|%s|%s", workItem.ID, workItem.WorkItemType, workItem.CreatedDate.In(localTzLocation).Format("2006-01-02 03:04 PM -0700"), workItem.AssignedTo, workItem.State, strings.ReplaceAll(workItem.Title, "|", ""))
		// Add newline only if it's not the last item
		if i < len(workItems)-1 {
			tableData += "\n"
		}
	}
	return tableData
}

var _typeColors = map[string]tcell.Color{
	"Epic":                 tcell.ColorOrange,
	"Feature":              tcell.ColorViolet,
	"User Story":           tcell.ColorTeal,
	"Bug":                  tcell.ColorRed,
	"Task":                 tcell.ColorYellow,
	"Issue":                tcell.ColorDarkGreen,
	"Product Backlog Item": tcell.ColorLightBlue,
	"Requirement":          tcell.ColorLightBlue,
	"Impediment":           tcell.ColorLightSeaGreen,
	// Some custom types are supported
	"Script":     tcell.ColorOrange,
	"Test Case":  tcell.ColorLightPink,
	"Test Suite": tcell.ColorLightPink,
}

var _stateColors = map[string]tcell.Color{
	"To Do":             tcell.ColorBlue,
	"New":               tcell.ColorBlue,
	"Code Review":       tcell.ColorBlue,
	"Development":       tcell.ColorBlue,
	"In Progress":       tcell.ColorYellow,
	"Done":              tcell.ColorGreen,
	"ICR":               tcell.ColorGreen,
	"Closed":            tcell.ColorGreen,
	"Pending":           tcell.ColorOrange,
	"Awaiting Decision": tcell.ColorOrange,
	"On Hold":           tcell.ColorOrange,
	"Acceptance":        tcell.ColorOrange,
	"Ready for test":    tcell.ColorRed,
	"Test":              tcell.ColorRed,
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
			// Work item type column
			if column == 1 {
				color = tcell.ColorWhite
				if _, ok := _typeColors[cell]; ok {
					color = _typeColors[cell]
				}
			}
			// State column
			if column == 4 {
				color = tcell.ColorWhite
				if _, ok := _stateColors[cell]; ok {
					color = _stateColors[cell]
				}
			}

			// "Assigned to" column
			if column == 3 && row > 0 {
				if row-1 < len(workItems) && activeUser != nil && curWorkItem.IsAssignedToUser(activeUser) {
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
	return name == activeUser.DisplayName || name == activeUser.Username
}

func workItemToDetailsData(workItem *azuredevops.WorkItem) string {
	workItemIsAssignedToUser := workItem.IsAssignedToUser(activeUser)
	var buf bytes.Buffer
	w := tabwriter.NewWriter(&buf, 0, 0, 2, ' ', tabwriter.TabIndent)

	var keyColor = "[blue]"
	var valueColor = "[white]"

	fmt.Fprintf(w, "%sTitle%s\t%s\n", keyColor, valueColor, workItem.Title)
	fmt.Fprintf(w, "%sID%s\t%d\n", keyColor, valueColor, workItem.ID)
	fmt.Fprintf(w, "%sURL%s\t%s\n", keyColor, valueColor, workItem.GetURL(_organization, _project))
	fmt.Fprintf(w, "%sWork Item Type%s\t%s\n", keyColor, valueColor, workItem.WorkItemType)
	if workItemIsAssignedToUser {
		fmt.Fprintf(w, "%sAssigned To%s\t[green]%s[white]\n", keyColor, valueColor, workItem.AssignedTo)
	} else {
		fmt.Fprintf(w, "%sAssigned To%s\t%s\n", keyColor, valueColor, workItem.AssignedTo)
	}
	fmt.Fprintf(w, "%sState%s\t%s\n", keyColor, valueColor, workItem.State)

	if workItem.Details != nil {
		fmt.Fprintf(w, "%sArea Path%s\t%s\n", keyColor, valueColor, workItem.Details.SystemAreaPath)
		fmt.Fprintf(w, "%sPriority%s\t%d\n", keyColor, valueColor, workItem.Details.Priority)
		fmt.Fprintf(w, "%sSeverity%s\t%s\n", keyColor, valueColor, workItem.Details.Severity)
	} else {
		fmt.Fprintf(w, "%sArea Path%s\t[yellow]Loading...[white]\n", keyColor, valueColor)
		fmt.Fprintf(w, "%sPriority%s\t[yellow]Loading...[white]\n", keyColor, valueColor)
		fmt.Fprintf(w, "%sSeverity%s\t[yellow]Loading...[white]\n", keyColor, valueColor)
	}

	fmt.Fprintf(w, "%sIteration Path%s\t%s\n", keyColor, valueColor, workItem.IterationPath)

	if workItem.WorkItemType == "Bug" {
		if workItem.Details != nil {
			fmt.Fprintf(w, "%sDescription%s\t\n%s\n\n", keyColor, valueColor, normalizeDataString(workItem.Details.ReproSteps))
		} else {
			fmt.Fprintf(w, "%sDescription%s\t[yellow]Loading...[white]\n", keyColor, valueColor)
		}
	} else {
		fmt.Fprintf(w, "%sDescription%s\t\n%s\n\n", keyColor, valueColor, normalizeDataString(workItem.Description))
	}

	// Add tags section
	fmt.Fprintf(w, "%sTags%s\t", keyColor, valueColor)
	if len(workItem.Tags) > 0 {
		fmt.Fprintf(w, "%s\n", workItem.Tags)
	} else {
		fmt.Fprintf(w, "[]\n")
	}

	fmt.Fprintf(w, "%sCreated On%s\t%s\n", keyColor, valueColor, workItem.CreatedDate)
	if isSameAsUser(workItem.CreatedBy) {
		fmt.Fprintf(w, "%sCreated By%s\t[green]%s[white]\n", keyColor, valueColor, workItem.CreatedBy)
	} else {
		fmt.Fprintf(w, "%sCreated By%s\t%s\n", keyColor, valueColor, workItem.CreatedBy)
	}
	fmt.Fprintf(w, "%sChanged On%s\t%s\n", keyColor, valueColor, workItem.ChangedDate)
	if isSameAsUser(workItem.ChangedBy) {
		fmt.Fprintf(w, "%sChanged By%s\t[green]%s[white]\n", keyColor, valueColor, workItem.ChangedBy)
	} else {
		fmt.Fprintf(w, "%sChanged By%s\t%s\n", keyColor, valueColor, workItem.ChangedBy)
	}

	if workItem.Details != nil {
		fmt.Fprintf(w, "\n%sAdditional details%s\n", keyColor, valueColor)
		fmt.Fprintf(w, "%sAcceptance Criteria%s\t\n%s\n\n", keyColor, valueColor, normalizeDataString(workItem.Details.AcceptanceCriteria))
		fmt.Fprintf(w, "%sBoard Column%s\t%s\n", keyColor, valueColor, workItem.Details.BoardColumn)
		fmt.Fprintf(w, "%sComment Count%s\t%d\n", keyColor, valueColor, workItem.Details.CommentCount)
		latestComment := normalizeDataString(workItem.Details.LatestComment)
		if latestComment == "" {
			latestComment = "[red][Cannot fetch latest comment][white]"
		}
		fmt.Fprintf(w, "%sLatest Comment%s\t\n%s\n\n", keyColor, valueColor, latestComment)
		// fmt.Fprintf(w, "PR Refs\t%s\n", strings.Join(workItem.GetPRs(), ", "))
		if len(workItem.PRDetails) > 0 {
			fmt.Fprintf(w, "%sAssociated Pull Requests%s\n\n", keyColor, valueColor)
			for _, pr := range workItem.PRDetails {
				fmt.Fprintf(w, "\t- %d %s\n", pr.ID, pr.Title)
				if pr.IsDraft {
					fmt.Fprintf(w, "\t  %sIs Draft%s\t[yellow][Yes][white]\n", keyColor, valueColor)
				}
				statusColor := "[white]"
				if pr.Status == "completed" {
					statusColor = "[green]"
				} else if pr.Status == "abandoned" {
					statusColor = "[red]"
				} else if pr.Status == "active" {
					statusColor = "[yellow]"
				}
				fmt.Fprintf(w, "\t  %sStatus%s\t%s%s[white]\n", keyColor, valueColor, statusColor, pr.Status)
				if isSameAsUser(pr.Author) {
					fmt.Fprintf(w, "\t  %sAuthor%s\t[green]%s[white]\n", keyColor, valueColor, pr.Author)
				} else {
					fmt.Fprintf(w, "\t  %sAuthor%s\t%s\n", keyColor, valueColor, pr.Author)
				}
				fmt.Fprintf(w, "\t  %sURL%s\t%s\n", keyColor, valueColor, pr.GetURL())
				fmt.Fprintf(w, "\t  %sCreated Date%s\t%s\n", keyColor, valueColor, pr.CreatedDate)
				if pr.ClosedBy != "" {
					if isSameAsUser(pr.ClosedBy) {
						fmt.Fprintf(w, "\t  %sClosed By%s\t[green]%s[white]\n", keyColor, valueColor, pr.ClosedBy)
					} else {
						fmt.Fprintf(w, "\t  %sClosed By%s\t%s\n", keyColor, valueColor, pr.ClosedBy)
					}
					fmt.Fprintf(w, "\t  %sClosed Date%s\t%s\n", keyColor, valueColor, pr.ClosedDate)
				} else {
					fmt.Fprintf(w, "\t  %sClosed By%s\t-\n", keyColor, valueColor)
					fmt.Fprintf(w, "\t  %sClosed Date%s\t-\n", keyColor, valueColor)
				}
				fmt.Fprintf(w, "\t  %sFrom branch%s\t%s\n", keyColor, valueColor, pr.SourceRefName)
				fmt.Fprintf(w, "\t  %sTo branch%s\t%s\n", keyColor, valueColor, pr.TargetRefName)
				votesInfo := pr.GetVotesInfo()
				fmt.Fprintf(w, "\t  %sReviewers%s\n", keyColor, valueColor)
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
	} else {
		fmt.Fprintf(w, "%s\nAdditional details%s\t[yellow]Loading...[white]\n", keyColor, valueColor)
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
	detailsPanelIsExpanded := false

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
		detailsPanelIsExpanded = false
		detailsPanel.SetText("")
		// Cleanup expanded details panel
		detailsPanelIsExpanded = false
	}

	toggleExpandedDetailsPanel := func() {
		if !detailsVisible {
			return
		}
		if detailsPanelIsExpanded {
			// Reset the details panel
			detailsPanelIsExpanded = false
			mainFlex.RemoveItem(detailsPanel)
			mainFlex.AddItem(detailsPanel, 0, 1, false)
			detailsPanel.SetBorderColor(tcell.ColorWhite)
			app.SetFocus(table)
		} else {
			// Expand the details panel
			detailsPanelIsExpanded = true
			detailsPanel.SetBorderColor(DetailsPanelBorderColorExpanded)
			mainFlex.RemoveItem(detailsPanel)
			mainFlex.AddItem(detailsPanel, 0, 100, false)
			app.SetFocus(detailsPanel)
		}
	}

	toggleDetailsPanel := func() {
		if detailsVisible {
			closeDetailPanel()
		} else {
			// Set active panel for keyboard context
			activePanel = "details"
			// Show details
			detailsPanel.SetBorderColor(tcell.ColorWhite)
			mainFlex.AddItem(detailsPanel, 0, 1, false)
			detailsVisible = true
			displayCurrentWorkItemDetails()
		}
	}

	detailsPanel.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		// Pressing 'Enter' while the `detailsPanelIsExpanded` will close the details panel
		if event.Key() == tcell.KeyEnter && detailsVisible && detailsPanelIsExpanded {
			closeDetailPanel()
			app.SetFocus(table)
			return nil
		}
		return event
	})

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
					table.SetCell(0, 0, tview.NewTableCell("No work items found. Try other filters (press \\ and Up or Down)").
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
		if activePanel == "details" && event.Rune() == 'q' && !searchMode {
			closeDetailPanel()
			app.SetFocus(table)
			return nil
		}

		// Handle 'd' key to toggle details panel full view (if details are visible)
		if activePanel == "details" && event.Rune() == 'd' && !searchMode {
			toggleExpandedDetailsPanel()
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
				table.Clear()
				table.SetCell(0, 0, tview.NewTableCell("No work items found. Try other filters (press \\ and Up or Down)").
					SetTextColor(tcell.ColorRed).
					SetAlign(tview.AlignCenter))
			})
		}
	}()

	return "Work Items", mainWindow
}
