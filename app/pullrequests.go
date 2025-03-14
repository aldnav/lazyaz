package main

import (
	"fmt"
	"log"
	"strconv"
	"strings"

	"github.com/aldnav/lazyaz/pkg/azuredevops"
	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
	"golang.org/x/text/cases"
	"golang.org/x/text/language"
)

const prTableData = `ID|Title|Status|Merge Status|Creator|Created On|Approvals|Repository
1|Loading...|Loading...|Loading...|Loading...|Loading...|Loading...|Loading...
`

var _activeUser *azuredevops.UserProfile

func _prsToTableData(prs []azuredevops.PullRequestDetails) string {
	tableData := "ID|Title|Status|Merge Status|Creator|Created On|Approvals|Repository\n"
	for i, pr := range prs {
		tableData += fmt.Sprintf(
			"%d|%s|%s|%s|%s|%s|%d|%s",
			pr.ID,
			pr.Title,
			cases.Title(language.English).String(pr.Status),
			cases.Title(language.English).String(pr.MergeStatus),
			pr.Author,
			pr.CreatedDate.Format("2006-01-02"),
			pr.GetApprovals(),
			pr.Repository,
		)
		if i < len(prs)-1 {
			tableData += "\n"
		}
	}
	return tableData
}

// TODO Move to utilities
func _isSameAsUser(name string) bool {
	return name == _activeUser.DisplayName
}

func _redrawTable(table *tview.Table, prs []azuredevops.PullRequestDetails) {
	table.Clear()
	tableData := _prsToTableData(prs)
	for row, line := range strings.Split(tableData, "\n") {
		for column, cell := range strings.Split(line, "|") {
			tableCell := tview.NewTableCell(cell).
				SetSelectable(row != 0)

			// Colors
			color := tcell.ColorWhite
			if row > 0 {
				if column == 0 {
					// ID column
					color = tcell.ColorRed
				}
				if column == 2 {
					// Status column
					if cell == "Active" {
						color = tcell.ColorYellow
					} else if cell == "Completed" {
						color = tcell.ColorGreen
					} else if cell == "Abandoned" {
						color = tcell.ColorGray
					}
				}
				if column == 3 {
					// Merge Status column
					if cell == "Conflicts" {
						color = tcell.ColorRed
					}
				}
				if column == 4 {
					// Creator column
					if _isSameAsUser(cell) {
						color = tcell.ColorLimeGreen
					}
				}
				if column == 6 {
					// Approvals column
					if cell == "0" {
						color = tcell.ColorYellow
					}
					tableCell.SetAlign(tview.AlignRight)
				}
				// if column == 7 {
				// 	// ClosedBy column
				// 	if _isSameAsUser(cell) {
				// 		color = tcell.ColorLimeGreen
				// 	}
				// }
			}

			// If it is the last column, set expanded
			if column == len(strings.Split(line, "|"))-1 {
				tableCell.SetExpansion(1)
			}

			tableCell.SetTextColor(color)
			table.SetCell(row, column, tableCell)
		}
	}
	table.Select(0, 0)
}

func PullRequestsPage(nextSlide func()) (title string, content tview.Primitive) {
	var prs []azuredevops.PullRequestDetails
	// Add search-related variables
	var searchText, previousSearchText string
	var searchMode bool = false
	var searchInput *tview.InputField
	var searchMatches []struct{ row, col int }
	var currentMatchIndex int = -1

	table := tview.NewTable().
		SetFixed(1, 1).
		SetBorders(false).
		SetSelectable(true, false).
		SetSeparator(' ')

	table.SetSelectedStyle(tcell.StyleDefault.
		Foreground(tcell.ColorBlack).
		Background(tcell.ColorLimeGreen))

	for row, line := range strings.Split(prTableData, "\n") {
		for column, cell := range strings.Split(line, "|") {
			tableCell := tview.NewTableCell(cell).
				SetSelectable(row != 0)
			table.SetCell(row, column, tableCell)
		}
	}

	mainWindow := tview.NewFlex().
		SetDirection(tview.FlexRow)

	tableFlex := tview.NewFlex().
		SetDirection(tview.FlexRow).
		AddItem(table, 0, 1, true)

	// Actions specific for pull requests
	actionsPanel := tview.NewFlex().
		SetDirection(tview.FlexColumn)
	searchStatus := tview.NewTextView().SetText("").SetTextAlign(tview.AlignRight)
	actionsPanel.AddItem(searchStatus, 0, 1, false)

	mainWindow.AddItem(tableFlex, 0, 1, true)
	mainWindow.AddItem(actionsPanel, 1, 1, false)

	// Handle search
	closeSearch := func() {
		searchMode = false
		mainWindow.RemoveItem(searchInput)
		app.SetFocus(table)
	}

	// Function to highlight the current match
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

					for idx, pr := range prs {
						// Search from the following:
						// ID, Title, Status, Author, Created On, Approvals, Repository
						if strings.Contains(strconv.Itoa(pr.ID), searchText) ||
							strings.Contains(pr.Title, searchText) ||
							strings.Contains(pr.Status, searchText) ||
							strings.Contains(pr.Author, searchText) ||
							strings.Contains(pr.CreatedDate.Format("2006-01-02"), searchText) ||
							strings.Contains(pr.Repository, searchText) {
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

	// Load data
	go func() {
		// Get current user
		var err error
		_activeUser, err = client.GetUserProfile()
		if err != nil {
			log.Printf("Error fetching user profile: %v", err)
		}
		prs, err = client.GetPRsCreatedByUser(_activeUser.Mail, "")
		if err != nil {
			log.Printf("Error fetching pull requests: %v", err)
		}
		if len(prs) > 0 {
			app.QueueUpdateDraw(func() {
				_redrawTable(table, prs)
			})
		} else {
			app.QueueUpdateDraw(func() {
				table.SetCell(0, 0, tview.NewTableCell("No pull requests found").
					SetTextColor(tcell.ColorRed).
					SetAlign(tview.AlignCenter))
			})
		}
	}()

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

		return event
	})

	return "Pull Requests", mainWindow
}
