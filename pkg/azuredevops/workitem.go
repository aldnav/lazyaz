package azuredevops

import (
	"encoding/json"
	"fmt"
	"strconv"
	"strings"
	"time"
)

type WorkItem struct {
	ID                   int                  `json:"Id"`
	WorkItemType         string               `json:"Work Item Type"`
	Title                string               `json:"Title"`
	AssignedTo           string               `json:"Assigned To"`
	AssignedToUniqueName string               `json:"Assigned To Unique Name"`
	State                string               `json:"State"`
	Tags                 string               `json:"Tags"`
	IterationPath        string               `json:"Iteration Path"`
	CreatedDate          time.Time            `json:"CreatedDate"`
	CreatedBy            string               `json:"CreatedBy"`
	ChangedDate          time.Time            `json:"ChangedDate"`
	ChangedBy            string               `json:"ChangedBy"`
	Description          string               `json:"Description"`
	Details              *WorkItemDetails     `json:"-"`
	PRDetails            []PullRequestDetails `json:"-"`
}

type WorkItemDetails struct {
	ReproSteps         string   `json:"Repro Steps"`
	SystemAreaPath     string   `json:"System.AreaPath"`
	AcceptanceCriteria string   `json:"Acceptance Criteria"`
	BoardColumn        string   `json:"Board Column"`
	BoardColumnDone    bool     `json:"Board Column Done"`
	CommentCount       int      `json:"Comment Count"`
	LatestComment      string   `json:"Latest Comment"`
	PRRefs             []string `json:"PR Refs"`
	Priority           int      `json:"Priority"`
	Severity           string   `json:"Severity"`
}

// GetMoreWorkItemDetails retrieves the details of a specific work item
// Given a WorkItem, it will use the ID to fetch more details
func (wit *WorkItem) GetMoreWorkItemDetails() (*WorkItem, error) {
	output, err := runAzCommand("boards", "work-item", "show", "--id", strconv.Itoa(wit.ID), "--query", jmespathWorkItemDetailsQuery, "--output", "json")
	if err != nil {
		return nil, fmt.Errorf("error fetching work item details: %v", err)
	}

	// Parse the output
	var detail WorkItemDetails
	if err := json.Unmarshal(output, &detail); err != nil {
		return nil, fmt.Errorf("error parsing work item details: %v", err)
	}
	wit.Details = &detail

	return wit, nil
}

// Get URL of Work Item as it appears in the browser
func (wit *WorkItem) GetURL(organization string, project string) string {
	return fmt.Sprintf("%s%s/_workitems/edit/%d", organization, project, wit.ID)
}

// GetPRs retrieves the PRs associated with the work item
func (wit *WorkItem) GetPRs() []string {
	// For each PR ref, get the last part of the URL when split by "%2F"
	prs := []string{}
	for _, prRef := range wit.Details.PRRefs {
		prs = append(prs, strings.Split(prRef, "%2F")[len(strings.Split(prRef, "%2F"))-1])
	}
	return prs
}

// Get associated pull request details.
func (wit *WorkItem) GetPRDetails(c *Client) ([]PullRequestDetails, error) {
	if len(wit.PRDetails) > 0 {
		return wit.PRDetails, nil
	}
	prs := []PullRequestDetails{}
	for _, prRef := range wit.GetPRs() {
		pr, err := c.GetPRDetails(prRef)
		if err != nil {
			return nil, fmt.Errorf("error fetching PR details: %v", err)
		}
		prs = append(prs, *pr)
	}
	wit.PRDetails = prs
	return prs, nil
}

// Determines if the work item is assigned to the current user
func (wit *WorkItem) IsAssignedToUser(user *UserProfile) bool {
	return wit.AssignedToUniqueName == user.Mail
}
