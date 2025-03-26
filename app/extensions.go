package main

import (
	"bytes"
	"fmt"
	"log"
	"os"
	"reflect"
	"text/template"

	"github.com/aldnav/lazyaz/pkg/azuredevops"
	"golang.design/x/clipboard"
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

// ExportToTemplate exports the given domain object to a template
// It accepts WorkItem, PullRequestDetails, or PipelineRun
func ExportToTemplate(domain interface{}) (string, error) {
	switch d := domain.(type) {
	case azuredevops.WorkItem:
		template, _ := loadTemplate("workitem")
		applied, applyErr := applyTemplate(template, d)
		if applyErr != nil {
			log.Printf("Failed to apply template: %s", applyErr)
			return "NOK", applyErr
		}
		// Copy the rendered template to clipboard
		err := copyToClipboard(applied)
		if err != nil {
			log.Printf("Failed to copy to clipboard: %s", err)
			return "Failed to copy to clipboard", err
		}
	case azuredevops.PullRequestDetails:
		log.Printf("Exporting PullRequestDetails: %+v", d)
	case azuredevops.PipelineRun:
		log.Printf("Exporting PipelineRun: %+v", d)
	default:
		err := fmt.Errorf("unsupported domain type: %s", reflect.TypeOf(domain))
		log.Printf("%v", err)
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

func copyToClipboard(text string) error {
	err := clipboard.Init()
	if err != nil {
		return fmt.Errorf("failed to initialize clipboard: %s", err)
	}
	clipboard.Write(clipboard.FmtText, []byte(text))
	return nil
}
