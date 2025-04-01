package main

import (
	"bytes"
	"fmt"
	"log"
	"os"
	"os/exec"
	"reflect"
	"runtime"
	"strings"
	"text/template"

	"github.com/aldnav/lazyaz/pkg/azuredevops"
)

const DEFAULT_WORKITEM_TEMPLATE = `
# Work Item {{.ID}}

Title: {{.Title}}

Description:
{{.Description}}

`

const DEFAULT_PULLREQUEST_TEMPLATE = `
# Pull Request {{.ID}}

Title: {{.Title}}

Description:
{{.Description}}
`

const DEFAULT_PIPELINERUN_TEMPLATE = `
# Pipeline Run {{.ID}}

{{.BuildNumber}}
`

// ErrNoClipboard is returned when no clipboard utility is available
var ErrNoClipboard = fmt.Errorf("no clipboard utility found (install xsel or xclip)")

// ExportToTemplate exports the given domain object to a template
// It accepts WorkItem, PullRequestDetails, or PipelineRun
func ExportToTemplate(domain interface{}) (string, error) {
	var tmpl string
	var d interface{}

	switch v := domain.(type) {
	case azuredevops.WorkItem:
		var err error
		tmpl, err = loadTemplate("workitem")
		if err != nil {
			return "NOK", err
		}
		d = v
	case azuredevops.PullRequestDetails:
		var err error
		tmpl, err = loadTemplate("pullrequest")
		if err != nil {
			return "NOK", err
		}
		d = v
	case azuredevops.PipelineRun:
		var err error
		tmpl, err = loadTemplate("pipelinerun")
		if err != nil {
			return "NOK", err
		}
		d = v
	default:
		err := fmt.Errorf("unsupported domain type: %s", reflect.TypeOf(domain))
		log.Printf("%v", err)
		return "NOK", err
	}

	applied, applyErr := applyTemplate(tmpl, d)
	if applyErr != nil {
		log.Printf("Failed to apply template: %s", applyErr)
		return "NOK", applyErr
	}
	// Copy the rendered template to clipboard
	err := copyToClipboard(applied)
	if err != nil {
		log.Printf("Failed to copy to clipboard: %s", err)
		return "NOK", err
	}

	// fmt.Printf("Parameter passed: %+v\n", domain)
	return "OK", nil
}

func loadTemplate(domain string) (string, error) {
	tmpl := ""
	switch domain {
	case "workitem":
		tmpl = "workitem.md"
	case "pullrequest":
		tmpl = "pullrequest.md"
	case "pipelinerun":
		tmpl = "pipelinerun.md"
	}
	if tmpl == "" {
		return "", fmt.Errorf("unsupported domain type: %s", domain)
	}
	// Read from home / .lazyaz/templates/
	content, err := os.ReadFile(fmt.Sprintf("%s/.lazyaz/templates/%s", os.Getenv("HOME"), tmpl))
	if err != nil {
		defaultTemplate := ""
		switch domain {
		case "workitem":
			defaultTemplate = DEFAULT_WORKITEM_TEMPLATE
		case "pullrequest":
			defaultTemplate = DEFAULT_PULLREQUEST_TEMPLATE
		case "pipelinerun":
			defaultTemplate = DEFAULT_PIPELINERUN_TEMPLATE
		}
		return defaultTemplate, nil
	}
	return string(content), nil
}

func applyTemplate(thetemplate string, domain interface{}) (string, error) {
	tmpl, err := template.New("template").Parse(thetemplate)
	if err != nil {
		return "", fmt.Errorf("failed to parse template: %s", err)
	}
	var rendered bytes.Buffer
	err = tmpl.Execute(&rendered, domain)
	if err != nil {
		return "", fmt.Errorf("failed to execute template: %s", err)
	}
	return rendered.String(), nil
}

// copyToClipboard copies the given text to the clipboard
func copyToClipboard(text string) error {
	var cmd *exec.Cmd
	switch runtime.GOOS {
	case "darwin":
		cmd = exec.Command("pbcopy")
	case "windows":
		cmd = exec.Command("clip")
	case "linux":
		// Try xsel first, fall back to xclip
		if _, err := exec.LookPath("xsel"); err == nil {
			cmd = exec.Command("xsel", "-bi")
		} else if _, err := exec.LookPath("xclip"); err == nil {
			cmd = exec.Command("xclip", "-selection", "clipboard")
		} else {
			return ErrNoClipboard
		}
	default:
		return fmt.Errorf("unsupported platform")
	}

	cmd.Stdin = strings.NewReader(text)
	return cmd.Run()
}

// CheckClipboardUtility checks if the clipboard utility is available
func CheckClipboardUtility() bool {
	switch runtime.GOOS {
	case "darwin":
		// macOS uses pbcopy which is built-in
		return true
	case "windows":
		// Windows uses clip which is built-in
		return true
	case "linux":
		// Linux needs xsel or xclip
		if _, err := exec.LookPath("xsel"); err == nil {
			return true
		}
		if _, err := exec.LookPath("xclip"); err == nil {
			return true
		}
		return false
	default:
		return false
	}
}

// Open in browser
func OpenInBrowser(domain interface{}) (string, error) {
	url := ""
	switch v := domain.(type) {
	case azuredevops.WorkItem:
		url = v.GetURL(_organization, _project)
	case azuredevops.PullRequestDetails:
		url = v.GetURL()
	case azuredevops.PipelineRun:
		url = v.GetWebURL()
	default:
		return "NOK", fmt.Errorf("unsupported domain type: %s", reflect.TypeOf(domain))
	}
	if url == "" {
		return "NOK", fmt.Errorf("no URL found for domain type: %s", reflect.TypeOf(domain))
	}

	var cmd *exec.Cmd
	if runtime.GOOS == "windows" {
		cmd = exec.Command("explorer", url)
	} else {
		cmd = exec.Command("open", url)
	}

	if err := cmd.Run(); err != nil {
		return "NOK", fmt.Errorf("failed to open URL: %v", err)
	}

	return "OK", nil
}
