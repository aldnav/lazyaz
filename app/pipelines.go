package main

import (
	"bytes"
	"fmt"
	"log"
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

				if column == 7 {
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

	fmt.Fprintf(w, "%sRun ID%s\t%d\n", keyColor, valueColor, run.ID)
	fmt.Fprintf(w, "%sBuild Number%s\t%s\n", keyColor, valueColor, run.BuildNumber)
	fmt.Fprintf(w, "%sStatus%s\t%s\n", keyColor, valueColor, cases.Title(language.English).String(run.Status))
	fmt.Fprintf(w, "%sResult%s\t%s\n", keyColor, valueColor, cases.Title(language.English).String(run.Result))
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
	// Details panel variables
	var detailsVisible bool
	var detailsPanelIsExpanded bool
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

	flex := tview.NewFlex().
		SetDirection(tview.FlexColumn).
		AddItem(table, 0, 1, true)

	// Create a details panel
	detailsPanel := tview.NewTextView().
		SetScrollable(true).
		SetWordWrap(true)
	detailsPanel.
		SetDynamicColors(true).
		SetBorder(true).
		SetTitle(" Pipeline Run ")

	flex.AddItem(detailsPanel, 0, 0, false)
	// flex.ResizeItem(detailsPanel, 0, 0)

	displayPipelineRunDetails := func(runs []azuredevops.PipelineRun, index int) {
		if index >= 0 && index < len(runs) {
			currentPipelineRun := runs[index]
			details := pipelineRunToDetailsData(&currentPipelineRun)
			detailsPanel.SetText(details)
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
			detailsPanel.SetText("")
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
		detailsPanel.SetText("")
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

	go func() {
		var err error
		// definitions, err = fetchDefinitions()
		// if err != nil {
		// 	log.Fatalf("Error fetching pipeline definitions: %v", err)
		// }
		runs, err = fetchRuns()
		app.SetFocus(table)
		if err != nil {
			log.Fatalf("Error fetching pipeline runs: %v", err)
		}
		redrawRunsTable(table, runs)
	}()

	// Manage input capture
	flex.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		// Handle 'q' key to close details panel
		// if activePanel == "details" && event.Rune() == 'q' && !searchMode {
		if activePanel == "details" && event.Rune() == 'q' {
			closeDetailPanel()
			app.SetFocus(table)
			return nil
		}

		// Handle 'd' key to toggle details panel full view (if details are visible)
		// if activePanel == "details" && event.Rune() == 'd' && !searchMode {
		if activePanel == "details" && event.Rune() == 'd' {
			toggleExpandedDetailsPanel()
			return nil
		}
		return event
	})

	// For some reason, a parent mainWindow flex row does not propagate/handle resizeItem
	// doesn't have an effect. While pullrequests.go works fine.
	return "Pipelines", flex
}
