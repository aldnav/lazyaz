package main

import (
	"fmt"
	"log"
	"reflect"

	"github.com/aldnav/lazyaz/pkg/azuredevops"
)

// ExportToTemplate exports the given domain object to a template
// It accepts WorkItem, PullRequestDetails, or PipelineRun
func ExportToTemplate(domain interface{}) (string, error) {
	switch d := domain.(type) {
	case azuredevops.WorkItem:
		log.Printf("Exporting WorkItem: %+v", d)
	case azuredevops.PullRequestDetails:
		log.Printf("Exporting PullRequestDetails: %+v", d)
	case azuredevops.PipelineRun:
		log.Printf("Exporting PipelineRun: %+v", d)
	default:
		err := fmt.Errorf("unsupported domain type: %s", reflect.TypeOf(domain))
		log.Printf("%v", err)
		return "NOK", err
	}

	fmt.Printf("Parameter passed: %+v\n", domain)
	return "OK", nil
}
