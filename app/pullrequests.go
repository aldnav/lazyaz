package main

import (
	"strings"

	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
)

const prTableData = `ID|Title|State|CreatedBy|CreatedDate
1|PR1|Open|John Doe|2021-01-01
2|PR2|Closed|Jane Smith|2021-01-02
3|PR3|Merged|Alice Johnson|2021-01-03
`

// Cover returns the cover page.
func PullRequestsPage(nextSlide func()) (title string, content tview.Primitive) {
	table := tview.NewTable().
		SetFixed(1, 1).
		SetBorders(false).
		SetSelectable(true, false).
		SetSeparator(' ')

	for row, line := range strings.Split(prTableData, "\n") {
		for column, cell := range strings.Split(line, "|") {
			color := tcell.ColorWhite
			if row == 0 {
				color = tcell.ColorYellow
			} else if column == 0 {
				color = tcell.ColorDarkCyan
			}
			align := tview.AlignLeft
			if row == 0 {
				align = tview.AlignCenter
			} else if column == 0 || column >= 4 {
				align = tview.AlignRight
			}
			tableCell := tview.NewTableCell(cell).
				SetTextColor(color).
				SetAlign(align).
				SetSelectable(row != 0 && column != 0)
			if column >= 1 && column <= 3 {
				tableCell.SetExpansion(1)
			}
			table.SetCell(row, column, tableCell)
		}
	}

	// Create a Flex layout that centers the logo and subtitle.
	flex := tview.NewFlex().
		SetDirection(tview.FlexRow).
		AddItem(table, 0, 1, true)

	return "Pull Requests", flex
}
