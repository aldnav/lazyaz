package main

import (
	"fmt"
	"log"
	"strings"

	"github.com/aldnav/lazyaz/pkg/azuredevops"
	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
	"golang.org/x/text/cases"
	"golang.org/x/text/language"
)

const pipelineTableData = `Run ID|Build Number|Status|Result|Pipeline|Source Branch|Queue Time|Requested For
1|Loading...|Loading...|Loading...|Loading...|Loading...|Loading...|Loading...
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
	tableData := "Run ID|Build Number|Status|Result|Pipeline|Source Branch|Queue Time|Requested For\n"
	for i, run := range runs {
		tableData += fmt.Sprintf(
			"%d|%s|%s|%s|%s|%s|%s|%s",
			run.ID,
			run.BuildNumber,
			cases.Title(language.English).String(run.Status),
			cases.Title(language.English).String(run.Result),
			run.DefinitionName,
			run.SourceBranch,
			run.QueueTime.Format("2006-01-02 15:04:05"),
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
		for column, cell := range strings.Split(line, "|") {
			color := tcell.ColorWhite

			if row > 0 {
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
			}
			tableCell := tview.NewTableCell(cell).
				SetTextColor(color).
				SetAlign(tview.AlignLeft).
				SetSelectable(row != 0)
			table.SetCell(row, column, tableCell)
		}
	}
	table.Select(0, 0)
}

func PipelinesPage(nextSlide func()) (title string, content tview.Primitive) {
	// var definitions []azuredevops.Pipeline
	var runs []azuredevops.PipelineRun
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
		SetDirection(tview.FlexRow).
		AddItem(table, 0, 1, true)

	go func() {
		var err error
		// definitions, err = fetchDefinitions()
		// if err != nil {
		// 	log.Fatalf("Error fetching pipeline definitions: %v", err)
		// }
		runs, err = fetchRuns()
		if err != nil {
			log.Fatalf("Error fetching pipeline runs: %v", err)
		}
		redrawRunsTable(table, runs)
	}()

	return "Pipelines", flex
}
