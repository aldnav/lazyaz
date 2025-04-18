package main

import (
	"bytes"
	"fmt"
	"log"
	"strconv"
	"strings"
	"text/tabwriter"

	"github.com/aldnav/lazyaz/pkg/azuredevops"
	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
	"golang.org/x/text/cases"
	"golang.org/x/text/language"
)

const pipelineTableData = `Run ID|Build Number|Status|Result|Pipeline|Source Branch|Queue Time|Reason|Requested For
1|Loading...|Loading...|Loading...|Loading...|Loading...|Loading...|Loading...|Loading...
`

// Fetch pipeline definitions
func fetchDefinitions() ([]azuredevops.Pipeline, error) {
	definitions, err := client.GetPipelineDefinitions()
	if err != nil {
		return nil, fmt.Errorf("error fetching pipeline definitions: %v", err)
	}
	return definitions, nil
}

// Fetch pipeline runs
func fetchRuns() ([]azuredevops.PipelineRun, error) {
	runs, err := client.GetPipelineRuns()
	if err != nil {
		return nil, fmt.Errorf("error fetching pipeline runs: %v", err)
	}
	return runs, nil
}

func fetchRunsFiltered(pipelineDefinitionId int) ([]azuredevops.PipelineRun, error) {
	runs, err := client.GetPipelineRunsFiltered(pipelineDefinitionId, "", "", "", "", "")
	if err != nil {
		return nil, fmt.Errorf("error fetching pipeline runs: %v", err)
	}
	return runs, nil
}

func _runsToTableData(runs []azuredevops.PipelineRun) string {

	tableData := "Run ID|Build Number|Status|Result|Pipeline|Source Branch|Queue Time|Reason|Requested For\n"
	for i, run := range runs {
		tableData += fmt.Sprintf(
			"%d|%s|%s|%s|%s|%s|%s|%s|%s",
			run.ID,
			run.BuildNumber,
			cases.Title(language.English).String(run.Status),
			cases.Title(language.English).String(run.Result),
			run.DefinitionName,
			run.SourceBranch,
			run.QueueTime.Format("2006-01-02 15:04:05"),
			run.Reason,
			run.RequestedFor,
		)
		if i < len(runs)-1 {
			tableData += "\n"
		}
	}
	return tableData
}

var _runResultColors = map[string]tcell.Color{
	"Succeeded":           tcell.ColorGreen,
	"Partially Succeeded": tcell.ColorYellow,
	"None":                tcell.ColorWhite,
	"Failed":              tcell.ColorRed,
	"Canceled":            tcell.ColorGray,
}

var _runStatusColors = map[string]tcell.Color{
	"Postponed":   tcell.ColorGray,
	"Completed":   tcell.ColorGreen,
	"Not Started": tcell.ColorBlue,
	"None":        tcell.ColorWhite,
	"In Progress": tcell.ColorYellow,
	"Canceling":   tcell.ColorOrange,
}

func redrawRunsTable(table *tview.Table, runs []azuredevops.PipelineRun) {
	table.Clear()
	tableData := _runsToTableData(runs)
	for row, line := range strings.Split(tableData, "\n") {
		dataIdx := row - 1
		for column, cell := range strings.Split(line, "|") {
			color := tcell.ColorWhite

			if row > 0 {
				run := runs[dataIdx]
				if run.Deleted {
					color = tcell.ColorGray
				}

				if column == 0 {
					// ID column
					color = tcell.ColorRed
				}

				if column == 2 {
					// Status column
					if _, ok := _runStatusColors[cell]; ok {
						color = _runStatusColors[cell]
					}
				}

				if column == 3 {
					// Result column
					if _, ok := _runResultColors[cell]; ok {
						color = _runResultColors[cell]
					}
				}

				if column == 8 {
					// Requested For column
					if isSameAsUser(cell, activeUser) {
						color = tcell.ColorGreen
					}
				}
			}
			tableCell := tview.NewTableCell(cell).
				SetTextColor(color).
				SetAlign(tview.AlignLeft).
				SetSelectable(row != 0)
			table.SetCell(row, column, tableCell)

			// If it is the last column, set expanded
			if column == len(strings.Split(line, "|"))-1 {
				tableCell.SetExpansion(1)
			}
		}
	}
	table.Select(0, 0)
}

func pipelineRunToDetailsData(run *azuredevops.PipelineRun) string {
	var buf bytes.Buffer
	w := tabwriter.NewWriter(&buf, 0, 0, 2, ' ', tabwriter.TabIndent)

	var keyColor = "[blue]"
	var valueColor = "[white]"
	var highlightColor = ""

	fmt.Fprintf(w, "%sRun ID%s\t%d\n", keyColor, valueColor, run.ID)
	fmt.Fprintf(w, "%sBuild Number%s\t%s\n", keyColor, valueColor, run.BuildNumber)
	fmt.Fprintf(w, "%sStatus%s\t%s\n", keyColor, valueColor, cases.Title(language.English).String(run.Status))
	if run.Result == "failed" {
		highlightColor = "[red]"
	} else if run.Result == "succeeded" {
		highlightColor = "[green]"
	} else {
		highlightColor = ""
	}
	fmt.Fprintf(w, "%sResult%s\t%s%s[white]\n", keyColor, valueColor, highlightColor, cases.Title(language.English).String(run.Result))
	highlightColor = ""
	fmt.Fprintf(w, "%sPipeline%s\t%s\n", keyColor, valueColor, run.DefinitionName)
	fmt.Fprintf(w, "%sURL%s\t%s\n", keyColor, valueColor, run.GetWebURL())
	fmt.Fprintf(w, "%sSource Branch%s\t%s\n", keyColor, valueColor, run.SourceBranch)
	fmt.Fprintf(w, "%sQueue%s\t%s\n", keyColor, valueColor, run.Queue)
	fmt.Fprintf(w, "%sQueue Time%s\t%s\n", keyColor, valueColor, run.QueueTime.Format("2006-01-02 15:04:05"))
	fmt.Fprintf(w, "%sStart Time%s\t%s\n", keyColor, valueColor, run.StartTime.Format("2006-01-02 15:04:05"))
	fmt.Fprintf(w, "%sFinish Time%s\t%s\n", keyColor, valueColor, run.FinishTime.Format("2006-01-02 15:04:05"))
	fmt.Fprintf(w, "%sReason%s\t%s\n", keyColor, valueColor, run.Reason)
	if isSameAsUser(run.RequestedFor, activeUser) {
		fmt.Fprintf(w, "%sRequested For%s\t%s\n", keyColor, valueColor, "[green]"+run.RequestedFor+"[white]")
	} else {
		fmt.Fprintf(w, "%sRequested For%s\t%s\n", keyColor, valueColor, run.RequestedFor)
	}
	if isSameAsUser(run.RequestedBy, activeUser) {
		fmt.Fprintf(w, "%sRequested By%s\t%s\n", keyColor, valueColor, "[green]"+run.RequestedBy+"[white]")
	} else {
		fmt.Fprintf(w, "%sRequested By%s\t%s\n", keyColor, valueColor, run.RequestedBy)
	}
	fmt.Fprintf(w, "%sPriority%s\t%s\n", keyColor, valueColor, run.Priority)
	if run.RepositoryName != "" {
		fmt.Fprintf(w, "%sRepository%s\t%s\n", keyColor, valueColor, run.RepositoryName)
	} else {
		fmt.Fprintf(w, "%sRepository%s\t%s\n", keyColor, valueColor, run.RepositoryID)
	}
	fmt.Fprintf(w, "%sRepository Type%s\t%s\n", keyColor, valueColor, run.RepositoryType)
	fmt.Fprintf(w, "%sLogs URL%s\t%s\n", keyColor, valueColor, run.LogsURL)

	w.Flush()
	return buf.String()
}

func PipelinesPage(nextSlide func()) (title string, content tview.Primitive) {
	// var definitions []azuredevops.Pipeline
	var runs []azuredevops.PipelineRun
	var currentIndex int
	isFetching := make(chan bool, 1)
	// Details panel variables
	var detailsVisible bool
	var detailsPanelIsExpanded bool
	// Actions specific for pipelines
	actionsPanel := tview.NewFlex().
		SetDirection(tview.FlexColumn)
	// Dropdown variables
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
		)
	actionsPanel.AddItem(dropdown, 0, 1, false)
	var pipelineIds []int
	var currentPipelineDefinitionId int
	// Search related variables
	var searchText, previousSearchText string
	var searchMode bool = false
	var searchInput *tview.InputField
	var searchMatches []struct{ row, col int }
	var currentMatchIndex int = -1
	searchStatus := tview.NewTextView().SetText("").SetTextAlign(tview.AlignRight)
	actionsPanel.AddItem(searchStatus, 0, 1, false)

	table := tview.NewTable().
		SetFixed(1, 1).
		SetBorders(false).
		SetSelectable(true, false).
		SetSeparator(' ')

	table.SetSelectedStyle(tcell.StyleDefault.
		Foreground(tcell.ColorBlack).
		Background(tcell.ColorLimeGreen))

	for row, line := range strings.Split(pipelineTableData, "\n") {
		for column, cell := range strings.Split(line, "|") {
			color := tcell.ColorWhite
			tableCell := tview.NewTableCell(cell).
				SetTextColor(color).
				SetAlign(tview.AlignLeft).
				SetSelectable(row != 0)
			table.SetCell(row, column, tableCell)
		}
	}

	mainWindow := tview.NewFlex().
		SetDirection(tview.FlexRow)
	flex := tview.NewFlex().
		SetDirection(tview.FlexColumn).
		AddItem(table, 0, 1, true)

	// Create a details panel
	detailsPanel := tview.NewFlex().
		SetDirection(tview.FlexRow)
	detailsPanel.SetBorder(true).
		SetTitle(" Pipeline Run ")
	detailsTextView := tview.NewTextView().
		SetScrollable(true).
		SetWordWrap(true)
	detailsTextView.SetDynamicColors(true)
	detailsPanel.AddItem(detailsTextView, 0, 1, false)

	flex.AddItem(detailsPanel, 0, 0, false)
	mainWindow.AddItem(flex, 0, 1, true)
	mainWindow.AddItem(actionsPanel, 1, 1, false)
	// flex.ResizeItem(detailsPanel, 0, 0)

	// Add buttons to the extensions
	detailActionsPanel := tview.NewFlex().
		SetDirection(tview.FlexColumn)
	detailsPanel.AddItem(detailActionsPanel, 1, 1, false)
	AttachPipelineRunExtensions(detailActionsPanel, table, &runs)

	displayPipelineRunDetails := func(runs []azuredevops.PipelineRun, index int) {
		if index >= 0 && index < len(runs) {
			currentPipelineRun := runs[index]
			details := pipelineRunToDetailsData(&currentPipelineRun)
			detailsTextView.SetText(details)
		}
	}

	displayCurrentPipelineRunDetails := func() {
		displayPipelineRunDetails(runs, currentIndex)
	}

	table.SetSelectionChangedFunc(func(row, column int) {
		currentIndex = row - 1
		if currentIndex < 0 {
			currentIndex = 0
		}
		if detailsVisible {
			detailsTextView.SetText("")
			displayCurrentPipelineRunDetails()
		}
	})

	closeDetailPanel := func() {
		// Release active panel
		activePanel = ""
		// Hide details
		flex.ResizeItem(detailsPanel, 0, 0)
		detailsVisible = false
		detailsPanelIsExpanded = false
		detailsTextView.SetText("")
	}

	toggleDetailsPanel := func() {
		if detailsVisible {
			closeDetailPanel()
		} else {
			// Set active panel for keyboard context
			activePanel = "details"
			// Show details
			detailsPanel.SetBorderColor(tcell.ColorWhite)
			flex.ResizeItem(detailsPanel, 0, 1)
			detailsVisible = true
			displayCurrentPipelineRunDetails()
		}
	}

	table.SetSelectedFunc(func(row, column int) {
		toggleDetailsPanel()
	})

	toggleExpandedDetailsPanel := func() {
		if !detailsVisible {
			return
		}
		if detailsPanelIsExpanded {
			// Collapse the details panel to original size
			detailsPanelIsExpanded = false
			flex.ResizeItem(detailsPanel, 0, 1)
			detailsPanel.SetBorderColor(tcell.ColorWhite)
			app.SetFocus(table)
		} else {
			// Expand the details panel
			detailsPanelIsExpanded = true
			flex.ResizeItem(detailsPanel, 0, 100) // TODO See value?
			detailsPanel.SetBorderColor(DetailsPanelBorderColorExpanded)
			app.SetFocus(detailsPanel)
		}
	}

	// ** Pipeline dropdown **
	// Handle dropdown options selection
	handleDropdownSelection := func() {
		dropdown.SetSelectedFunc(func(text string, index int) {
			app.SetFocus(table)
			currentPipelineDefinitionId = pipelineIds[index]

			// Refresh runs based on filter options
			go func() {
				select {
				case isFetching <- true:
					dropdown.SetLabel("Fetching ")
					// TODO Support other filters
					var err error
					if currentPipelineDefinitionId == 0 {
						runs, err = fetchRuns()
					} else {
						runs, err = fetchRunsFiltered(currentPipelineDefinitionId)
					}
					<-isFetching
					if err != nil {
						log.Printf("Error fetching pipeline runs: %v", err)
					}
					if len(runs) > 0 {
						app.QueueUpdateDraw(func() {
							dropdown.SetLabel("")
							currentIndex = 0
							redrawRunsTable(table, runs)
							closeDetailPanel()
							app.SetFocus(table)
							table.Select(0, 0)
						})
					} else {
						app.QueueUpdateDraw(func() {
							dropdown.SetLabel("")
							table.SetCell(0, 0, tview.NewTableCell("No runs found. Try other filters (press \\ and Up or Down)").
								SetTextColor(tcell.ColorRed).
								SetAlign(tview.AlignCenter))
						})
					}
				default:
					// Another fetch is in progress, skip this one
					return
				}
			}()
		})
	}
	// Handle dropdown options display
	setOptionsFromDefinitions := func(definitions []azuredevops.Pipeline) {
		// Set options from definitions' names
		pipelineNames := make([]string, len(definitions)+1)
		pipelineNames[0] = "All"
		pipelineIds = append(pipelineIds, 0)
		for i, definition := range definitions {
			pipelineNames[i+1] = definition.Name + " [" + strconv.Itoa(definition.ID) + "]"
			pipelineIds = append(pipelineIds, definition.ID)
		}
		dropdown.SetOptions(pipelineNames, func(text string, index int) {
			currentPipelineDefinitionId = pipelineIds[index] // Why even have this here when it gets repeated on setSelectedFunc?
		})
		dropdown.SetCurrentOption(0)
		handleDropdownSelection()
	}

	// Handle search
	closeSearch := func() {
		searchMode = false
		mainWindow.RemoveItem(searchInput)
		app.SetFocus(table)
	}

	// Highlight the current match
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

				if searchText == previousSearchText {
					// Highlight next match (if any)
					currentMatchIndex += 1
				} else {
					// Perform new search
					searchMatches = nil
					currentMatchIndex = -1
					previousSearchText = searchText

					// Search for the matches in the pull requests
					var normalizedSearchText = strings.ToLower(searchText)
					for idx, run := range runs {
						// Search from the following:
						// ID, Title, Status, Author, Created On, Approvals, Repository
						if strings.Contains(strconv.Itoa(run.ID), normalizedSearchText) ||
							strings.Contains(strings.ToLower(run.BuildNumber), normalizedSearchText) ||
							strings.Contains(strings.ToLower(run.SourceBranch), normalizedSearchText) ||
							strings.Contains(strings.ToLower(run.RequestedFor), normalizedSearchText) ||
							strings.Contains(run.RepositoryName, normalizedSearchText) {
							searchMatches = append(searchMatches, struct{ row, col int }{idx + 1, 0})
						}
					}

					if len(searchMatches) > 0 {
						// Highlight the first match
						currentMatchIndex = 0
					}
				}

				if currentMatchIndex >= len(searchMatches) {
					currentMatchIndex = 0 // Wrap around
				}

				// Handle search results
				if len(searchMatches) > 0 {
					highlightMatch()
				} else {
					// Show "No matches" message
					searchStatus.SetLabel("No matches!")
				}
				closeSearch()
			} else if key == tcell.KeyEscape {
				closeSearch()
			}
		})

	loadData := func() {
		select {
		case isFetching <- true:
			var err error
			definitions, err := fetchDefinitions()
			if err != nil {
				<-isFetching // Release the lock
				log.Fatalf("Error fetching pipeline definitions: %v", err)
				AnnounceError("❌ Error fetching pipeline definitions")
			}
			setOptionsFromDefinitions(definitions)
			runs, err = fetchRuns()
			<-isFetching // Release the lock
			if err != nil {
				log.Fatalf("Error fetching pipeline runs: %v", err)
				AnnounceError("❌ Error fetching pipeline runs")
			} else {
				Announce("✅ Refresh done", 3)
			}
			redrawRunsTable(table, runs)
			app.QueueUpdateDraw(func() {
				if detailsVisible {
					displayCurrentPipelineRunDetails()
				}
			})
		default:
			// Another fetch is in progress, skip this one
			return
		}
	}

	go loadData()

	// Manage input capture
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

		// Handle 'r' key to refresh the data
		if event.Rune() == 'r' && !searchMode {
			Announce("⏳ Refreshing pipelines...", -1)
			go loadData()
			return nil
		}
		return event
	})

	// For some reason, a parent mainWindow flex row does not propagate/handle resizeItem
	// doesn't have an effect. While pullrequests.go works fine.
	return "Pipelines", mainWindow
}
