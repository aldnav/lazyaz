package main

import (
	"fmt"
	"log"
	"os"

	"github.com/aldnav/lazyaz/pkg/azuredevops"
	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
)

func main() {
	// Configure logging to stderr
	log.SetOutput(os.Stderr)
	log.SetPrefix("[lazyaz] ")
	log.Println("Application starting...")

	// Initialize a new application
	app := tview.NewApplication()

	// Create main layout
	flex := tview.NewFlex().SetDirection(tview.FlexRow)

	// Create header text view
	headerView := tview.NewTextView().
		SetTextAlign(tview.AlignCenter).
		SetDynamicColors(true)

	// Create projects list
	projectsList := tview.NewList().
		ShowSecondaryText(true).
		SetHighlightFullLine(true).
		SetMainTextColor(tcell.ColorWhite).
		SetSecondaryTextColor(tcell.ColorGray)
	// Load configuration
	log.Println("Loading configuration...")
	config, err := azuredevops.NewConfig()
	if err != nil {
		// Display error message if configuration is missing
		log.Printf("Configuration error: %v", err)
		headerView.SetText(fmt.Sprintf("[red]Error: %s[white]\n\nPlease set the required environment variables and restart the application.", err))
	} else {
		log.Printf("Configuration loaded successfully. Organization: %s, Project: %s", config.Organization, config.Project)
		// Display loading message
		loadingMessage := fmt.Sprintf("[yellow]Connecting to Azure DevOps...\n\nOrganization: %s[white]", config.Organization)
		if config.Project != "" {
			loadingMessage += fmt.Sprintf("\n[yellow]Default Project: %s[white]", config.Project)
		}
		headerView.SetText(loadingMessage)

		// Create an Azure DevOps client
		log.Println("Creating Azure DevOps client...")
		client := azuredevops.NewClient(config)

		// Fetch projects in a goroutine to keep UI responsive
		go func() {
			log.Println("Fetching projects from Azure DevOps API...")
			projects, err := client.FetchProjects()
			if err != nil {
				// Update UI with error message
				log.Printf("API connection error: %v", err)
				app.QueueUpdateDraw(func() {
					headerView.SetText(fmt.Sprintf("[red]Error connecting to Azure DevOps:[white]\n%s", err))
				})
				return
			}

			log.Printf("Successfully fetched %d projects from Azure DevOps", len(projects))

			// Update UI with projects
			app.QueueUpdateDraw(func() {
				headerText := fmt.Sprintf("[green]Connected to Azure DevOps![white]\n\nFound [yellow]%d[white] projects in organization: [yellow]%s[white]", len(projects), config.Organization)
				if config.Project != "" {
					headerText += fmt.Sprintf("\nDefault project: [yellow]%s[white]", config.Project)
				}
				headerView.SetText(headerText)

				// Add projects to the list
				defaultProjectIndex := -1
				for i, project := range projects {
					desc := project.Description
					if desc == "" {
						desc = "No description"
					}

					displayName := project.Name
					// Add visual indicator for default project
					if config.Project != "" && project.Name == config.Project {
						displayName = fmt.Sprintf("[green]%s [DEFAULT][white]", project.Name)
						defaultProjectIndex = i
					}

					projectsList.AddItem(
						displayName,
						fmt.Sprintf("ID: %s | State: %s", project.ID, project.State),
						rune('a'+i%26), // Use letters as shortcuts
						nil,
					)
				}

				// Select the default project in the list if found
				if defaultProjectIndex >= 0 {
					projectsList.SetCurrentItem(defaultProjectIndex)
				}
			})
		}()
	}

	// Set up the layout
	headerView.SetBorder(true).SetTitle("LazyAz").SetTitleAlign(tview.AlignCenter)
	projectsList.SetBorder(true).SetTitle("Projects").SetTitleAlign(tview.AlignCenter)

	// Add components to the layout
	flex.AddItem(headerView, 5, 1, false)
	flex.AddItem(projectsList, 0, 3, true)

	// Set the flex container as the root primitive and run the application
	log.Println("Starting LazyAz terminal UI...")

	// Add global keyboard handler for quitting with 'q'
	app.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		if event.Rune() == 'q' {
			app.Stop()
			return nil
		}
		return event
	})

	if err := app.SetRoot(flex, true).Run(); err != nil {
		log.Printf("Terminal UI error: %v", err)
		panic(err)
	}
	log.Println("Application exiting...")
}
