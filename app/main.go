package main

import (
	"fmt"
	"log"
	"os"
	"strconv"

	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
)

type Slide func(nextSlide func()) (title string, content tview.Primitive)

var app = tview.NewApplication()
var activePanel string

func main() {
	log.SetOutput(os.Stderr)
	log.SetPrefix("[lazyaz] ")
	log.Println("Application starting...")

	slides := []Slide{
		WorkItemsPage,
		PullRequestsPage,
		PipelinesPage,
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

	// Creating the main layout
	layout := tview.NewFlex().
		SetDirection(tview.FlexRow).
		AddItem(pages, 0, 1, true).
		AddItem(info, 1, 1, false)

	// Shortcuts to navigate between slides
	app.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		if event.Key() == tcell.KeyEscape || (event.Rune() == 'q' && activePanel == "") {
			app.Stop()
		} else if event.Rune() == ']' {
			nextSlide()
			return nil
		} else if event.Rune() == '[' {
			prevSlide()
			return nil
		} else if r := event.Rune(); r >= '1' && r <= '9' {
			// Convert rune to integer (0-based index)
			slideIndex := int(r - '1')
			goToSlide(slideIndex)
			return nil
		}
		return event
	})
	app.EnableMouse(true)

	// Integrate with Azure DevOps

	// Start the application.
	if err := app.SetRoot(layout, true).EnableMouse(true).EnablePaste(true).Run(); err != nil {
		log.Printf("Terminal UI error: %v", err)
		panic(err)
	}
	log.Println("Application exiting...")
}
