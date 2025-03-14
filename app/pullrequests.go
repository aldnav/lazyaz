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

func prToDetailsData(pr *azuredevops.PullRequestDetails) string {
	var buf bytes.Buffer
	w := tabwriter.NewWriter(&buf, 0, 0, 2, ' ', tabwriter.TabIndent)

	fmt.Fprintf(w, "Title\t%s\n", pr.Title)
	fmt.Fprintf(w, "ID\t%d\n", pr.ID)
	fmt.Fprintf(w, "Status\t%s\n", cases.Title(language.English).String(pr.Status))
	fmt.Fprintf(w, "Draft\t%s\n", cases.Title(language.English).String(strconv.FormatBool(pr.IsDraft)))
	fmt.Fprintf(w, "Merge Status\t%s\n", cases.Title(language.English).String(pr.MergeStatus))
	if _isSameAsUser(pr.Author) {
		fmt.Fprintf(w, "Creator\t%s\n", "[green]"+pr.Author+"[white]")
	} else {
		fmt.Fprintf(w, "Creator\t%s\n", pr.Author)
	}
	fmt.Fprintf(w, "Created On\t%s\n", pr.CreatedDate)
	fmt.Fprintf(w, "Repository\t%s\n", pr.Repository)
	fmt.Fprintf(w, "Source Branch\t%s\n", pr.GetShortBranchName())
	fmt.Fprintf(w, "Target Branch\t%s\n", pr.GetShortTargetBranchName())
	fmt.Fprintf(w, "URL\t%s\n", pr.GetOrgURL(client.Config.Organization))
	fmt.Fprintf(w, "Reviews\n")

	for _, vote := range pr.GetVotesInfo() {
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

	fmt.Fprintf(w, "\nDescription\n")
	fmt.Fprintf(w, "%s\n", normalizeDataString(pr.Description))

	w.Flush()
	return buf.String()
}

func PullRequestsPage(nextSlide func()) (title string, content tview.Primitive) {
	var prs []azuredevops.PullRequestDetails
	var currentIndex int
	// Details panel variables
	var detailsVisible bool = false
	// Add search-related variables
	var searchText, previousSearchText string
	var searchMode bool = false
	var searchInput *tview.InputField
	var searchMatches []struct{ row, col int }
	var currentMatchIndex int = -1
	// Dropdown variables
	var pullRequestFilter string

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
		SetDirection(tview.FlexColumn).
		AddItem(table, 0, 1, true)

	// Create a details panel
	detailsPanel := tview.NewTextView().
		SetScrollable(true).
		SetWordWrap(true)
	detailsPanel.
		SetDynamicColors(true).
		SetBorder(true).
		SetTitle(" Pull Request ")

	// Actions specific for pull requests
	actionsPanel := tview.NewFlex().
		SetDirection(tview.FlexColumn)
	dropdown := tview.NewDropDown().
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
		SetOptions([]string{"Mine", "Assigned to me", "All", "Active", "Completed", "Abandoned"}, nil)
	dropdown.SetCurrentOption(0)
	searchStatus := tview.NewTextView().SetText("").SetTextAlign(tview.AlignRight)
	actionsPanel.AddItem(dropdown, 0, 1, false)
	actionsPanel.AddItem(searchStatus, 0, 1, false)

	mainWindow.AddItem(tableFlex, 0, 1, true)
	mainWindow.AddItem(actionsPanel, 1, 1, false)

	// Handle table enter key (View Pull Request details)

	displayPullRequestDetails := func(prs []azuredevops.PullRequestDetails, index int) {
		if index >= 0 && index < len(prs) {
			currentPullRequest := prs[index]
			details := prToDetailsData(&currentPullRequest)
			detailsPanel.SetText(details)
		}
	}

	displayCurrentPullRequestDetails := func() {
		displayPullRequestDetails(prs, currentIndex)
	}

	// When the table highlight is changed
	table.SetSelectionChangedFunc(func(row, column int) {
		currentIndex = row - 1
		if currentIndex < 0 {
			currentIndex = 0
		}
		if detailsVisible {
			detailsPanel.SetText("")
			displayPullRequestDetails(prs, currentIndex)
		}
	})

	closeDetailPanel := func() {
		// Release active panel for keyboard context
		activePanel = ""
		// Hide details
		tableFlex.RemoveItem(detailsPanel)
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
			tableFlex.AddItem(detailsPanel, 0, 1, false)
			detailsVisible = true
			displayCurrentPullRequestDetails()
		}
	}

	table.SetSelectedFunc(func(row, column int) {
		toggleDetailsPanel()
	})

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

	// Handle dropdown selection of Pull Requests
	dropdown.SetSelectedFunc(func(text string, index int) {
		app.SetFocus(table)
		var potentialPullRequestFilter string
		var getPRsFunc func() ([]azuredevops.PullRequestDetails, error)
		switch text {
		case "Mine":
			potentialPullRequestFilter = "mine"
		case "Assigned to me":
			potentialPullRequestFilter = "assigned-to-me"
		case "All":
			potentialPullRequestFilter = "all"
			getPRsFunc = client.GetAllPRs
		case "Active":
			potentialPullRequestFilter = "active"
			getPRsFunc = client.GetActivePRs
		case "Completed":
			potentialPullRequestFilter = "completed"
			getPRsFunc = client.GetCompletedPRs
		case "Abandoned":
			potentialPullRequestFilter = "abandoned"
			getPRsFunc = client.GetAbandonedPRs
		}
		if potentialPullRequestFilter != pullRequestFilter {
			pullRequestFilter = potentialPullRequestFilter
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

		// Refresh the pull requests
		go func() {
			dropdown.SetLabel("Fetching ")
			var err error
			if pullRequestFilter == "mine" {
				prs, err = client.GetPRsCreatedByUser(_activeUser.Mail, "")
			} else if pullRequestFilter == "assigned-to-me" {
				prs, err = client.GetPRsAssignedToUser(_activeUser.Mail)
			} else {
				prs, err = getPRsFunc()
			}
			if err != nil {
				log.Printf("Error fetching pull requests: %v", err)
			}
			if len(prs) > 0 {
				app.QueueUpdateDraw(func() {
					dropdown.SetLabel("")
					// currentIndex = 0
					_redrawTable(table, prs)
					// closeDetailPanel()
					app.SetFocus(table)
					table.Select(0, 0)
				})
			} else {
				app.QueueUpdateDraw(func() {
					table.Clear()
					table.SetCell(0, 0, tview.NewTableCell("No pull requests found").
						SetTextColor(tcell.ColorRed).
						SetAlign(tview.AlignCenter))
					app.SetFocus(table)
				})
			}
		}()
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

		// Handle 'q' key to close details panel
		if activePanel == "details" && event.Rune() == 'q' && !searchMode {
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

	return "Pull Requests", mainWindow
}
