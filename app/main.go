package main

import (
	"fmt"
	"log"
	"os"
	"strconv"
	"time"

	"github.com/aldnav/lazyaz/pkg/azuredevops"
	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
)

type Slide func(nextSlide func()) (title string, content tview.Primitive)

var app = tview.NewApplication()
var activePanel string
var client *azuredevops.Client

var activeUser *azuredevops.UserProfile
var userProfileErr error

var localTzLocation *time.Location

func main() {
	log.SetOutput(os.Stderr)
	log.SetPrefix("[lazyaz] ")
	log.Println("Application starting...")
	var err error
	localTzLocation, err = time.LoadLocation("Local")
	if err != nil {
		log.Printf("Error loading local timezone: %v\n", err)
		log.Println("Using UTC")
		localTzLocation = time.UTC
	} else {
		log.Printf("Using local timezone: %v\n", localTzLocation)
	}

	// Integrate with Azure DevOps early on init
	config, configErr := azuredevops.NewConfig()
	if configErr != nil {
		log.Printf("Configuration error: %v", configErr)
	}
	_organization = config.Organization
	_project = config.Project
	client = azuredevops.NewClient(config)
	// Get current user
	activeUser, userProfileErr = client.GetUserProfile()
	if userProfileErr != nil {
		log.Printf("Error fetching user profile: %v", userProfileErr)
	}

	slides := []Slide{
		WorkItemsPage,
		PullRequestsPage,
		// TODO Implement pipelines page
		// PipelinesPage,
	}

	pages := tview.NewPages()

	// Bottom bar
	info := tview.NewTextView().
		SetDynamicColors(true).
		SetRegions(true).
		SetWrap(false).
		SetHighlightedFunc(func(added, removed, remaining []string) {
			if len(added) > 0 {
				pages.SwitchToPage(added[0])
				app.SetFocus(pages)
			}
		})

	prevSlide := func() {
		slide, _ := strconv.Atoi(info.GetHighlights()[0])
		slide = (slide - 1 + len(slides)) % len(slides)
		info.Highlight(strconv.Itoa(slide)).
			ScrollToHighlight()
	}
	nextSlide := func() {
		slide, _ := strconv.Atoi(info.GetHighlights()[0])
		slide = (slide + 1) % len(slides)
		info.Highlight(strconv.Itoa(slide)).
			ScrollToHighlight()
	}
	goToSlide := func(index int) {
		if index >= 0 && index < len(slides) {
			info.Highlight(strconv.Itoa(index)).ScrollToHighlight()
		}
	}
	// Populate the pages
	for index, slide := range slides {
		title, primitive := slide(nextSlide)
		pages.AddPage(strconv.Itoa(index), primitive, true, index == 0)
		fmt.Fprintf(info, `[black]["%d"][limegreen:black] %d %s [""][black]`, index, index+1, title)
	}
	info.Highlight("0")

	// Connection status
	connectionStatus := tview.NewTextView().
		SetText("🚀 Connecting to Azure DevOps...").
		SetTextAlign(tview.AlignRight).
		SetTextColor(tcell.ColorYellow)

	infoBar := tview.NewFlex().
		SetDirection(tview.FlexColumn).
		AddItem(info, 0, 1, false).
		AddItem(connectionStatus, 0, 1, false)

	// Shortcuts to navigate between slides
	app.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		// TODO Enable "q" to stop the application
		if event.Key() == tcell.KeyEscape {
			app.Stop()
		} else if event.Rune() == ']' {
			nextSlide()
			return nil
		} else if event.Rune() == '[' {
			prevSlide()
			return nil
		} else if r := event.Rune(); r >= '1' && r <= '9' && true == false {
			// TODO Temp disable selection by number because of search with numbers
			// Convert rune to integer (0-based index)
			slideIndex := int(r - '1')
			goToSlide(slideIndex)
			return nil
		}
		return event
	})
	app.EnableMouse(true)

	if configErr != nil && userProfileErr != nil {
		connectionStatus.SetText("🚨 Error connecting to Azure DevOps: Inspect logs for more details.")
		connectionStatus.SetTextColor(tcell.ColorRed)
	} else {
		connectionStatus.SetText(fmt.Sprintf("✅ Connected to %s ", _organization))
		connectionStatus.SetTextColor(tcell.ColorGreen)
	}

	// Creating the main layout
	layout := tview.NewFlex().
		SetDirection(tview.FlexRow).
		AddItem(pages, 0, 1, true).
		AddItem(infoBar, 1, 1, false)

	// Start the application.
	if err := app.SetRoot(layout, true).EnableMouse(true).EnablePaste(true).Run(); err != nil {
		log.Printf("Terminal UI error: %v", err)
		panic(err)
	}
	log.Println("Application exiting...")
}
