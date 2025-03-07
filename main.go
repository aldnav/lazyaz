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
		log.Printf("Configuration loaded successfully. Organization: %s", config.Organization)
		// Display loading message
		headerView.SetText(fmt.Sprintf("[yellow]Connecting to Azure DevOps...\n\nOrganization: %s[white]", config.Organization))
		
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
				headerView.SetText(fmt.Sprintf("[green]Connected to Azure DevOps![white]\n\nFound [yellow]%d[white] projects in organization: [yellow]%s[white]", len(projects), config.Organization))
				
			// Add projects to the list
			for i, project := range projects {
					desc := project.Description
					if desc == "" {
						desc = "No description"
					}
					projectsList.AddItem(
						project.Name, 
						fmt.Sprintf("ID: %s | State: %s", project.ID, project.State),
						rune('a'+i%26), // Use letters as shortcuts
						nil,
					)
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
	if err := app.SetRoot(flex, true).Run(); err != nil {
		log.Printf("Terminal UI error: %v", err)
		panic(err)
	}
	log.Println("Application exiting...")
}

