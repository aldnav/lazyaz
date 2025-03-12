package azuredevops

import (
	"bufio"
	"bytes"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"time"
)

// execCommand is a variable that allows for mocking exec.Command in tests
var execCommand = exec.Command

// Config holds the Azure DevOps connection settings
type Config struct {
	Organization string
	Project      string
}

// Project represents an Azure DevOps project
type Project struct {
	ID          string    `json:"id"`
	Name        string    `json:"name"`
	Description string    `json:"description"`
	URL         string    `json:"url"`
	State       string    `json:"state"`
	LastUpdated time.Time `json:"lastUpdateTime"`
	Visibility  string    `json:"visibility"`
}

// AzCliProjectsResponse represents the Azure CLI response for projects
type AzCliProjectsResponse struct {
	Value []struct {
		ID             string    `json:"id"`
		Name           string    `json:"name"`
		Description    string    `json:"description"`
		URL            string    `json:"url"`
		State          string    `json:"state"`
		Visibility     string    `json:"visibility"`
		LastUpdateTime time.Time `json:"lastUpdateTime"`
	} `json:"value"`
	Count int `json:"count"`
}

// AzCliProjectResponse represents the Azure CLI response for a single project
type AzCliProjectResponse struct {
	ID             string    `json:"id"`
	Name           string    `json:"name"`
	Description    string    `json:"description"`
	URL            string    `json:"url"`
	State          string    `json:"state"`
	Visibility     string    `json:"visibility"`
	LastUpdateTime time.Time `json:"lastUpdateTime"`
}

// Client represents an Azure DevOps client
type Client struct {
	Config *Config
}

type WorkItem struct {
	ID            int                  `json:"Id"`
	WorkItemType  string               `json:"Work Item Type"`
	Title         string               `json:"Title"`
	AssignedTo    string               `json:"Assigned To"`
	State         string               `json:"State"`
	Tags          string               `json:"Tags"`
	IterationPath string               `json:"Iteration Path"`
	CreatedDate   time.Time            `json:"CreatedDate"`
	CreatedBy     string               `json:"CreatedBy"`
	ChangedDate   time.Time            `json:"ChangedDate"`
	ChangedBy     string               `json:"ChangedBy"`
	Description   string               `json:"Description"`
	Details       *WorkItemDetails     `json:"-"`
	PRDetails     []PullRequestDetails `json:"-"`
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
}

type PullRequestDetails struct {
	Author              string      `json:"Author"`
	ClosedBy            string      `json:"Closed By"`
	ClosedDate          time.Time   `json:"Closed Date"`
	CreatedDate         time.Time   `json:"Created Date"`
	Description         string      `json:"Description"`
	ID                  int         `json:"ID"`
	IsDraft             bool        `json:"Is Draft"`
	Labels              interface{} `json:"Labels"`
	MergeFailureMessage interface{} `json:"Merge Failure Message"`
	MergeFailureType    interface{} `json:"Merge Failure Type"`
	MergeStatus         string      `json:"Merge Status"`
	Repository          string      `json:"Repository"`
	RepositoryURL       string      `json:"Repository URL"`
	Reviewers           []string    `json:"Reviewers"`
	ReviewersVotes      []int       `json:"Reviewers Votes"`
	SourceRefName       string      `json:"Source Ref Name"`
	Status              string      `json:"Status"`
	TargetRefName       string      `json:"Target Ref Name"`
	Title               string      `json:"Title"`
	WorkItemRefs        []string    `json:"Work Item Refs"`
}

// NewConfig creates a new Config, first trying to read from config file, then falling back to environment variables
func NewConfig() (*Config, error) {
	// Try to read organization and project from config file first
	org, project := readConfigFromFile()

	// If organization not found in config file, fall back to environment variable
	if org == "" {
		org = os.Getenv("AZURE_DEVOPS_ORG")
	}

	// If project not found in config file, fall back to environment variable
	if project == "" {
		project = os.Getenv("AZURE_DEVOPS_PROJECT")
	}

	var missingVars []string
	if org == "" {
		missingVars = append(missingVars, "AZURE_DEVOPS_ORG")
	}

	if len(missingVars) > 0 {
		return nil, fmt.Errorf("missing required configuration: %s", strings.Join(missingVars, ", "))
	}

	return &Config{
		Organization: org,
		Project:      project,
	}, nil
}

// readConfigFromFile attempts to read the organization and project from ~/.azure/azuredevops/config
// Returns empty strings if file doesn't exist, can't be read, or doesn't contain the values
func readConfigFromFile() (string, string) {
	// Get user home directory
	home, err := os.UserHomeDir()
	if err != nil {
		return "", ""
	}

	// Build path to config file
	configPath := filepath.Join(home, ".azure", "azuredevops", "config")

	// Check if file exists
	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		return "", ""
	}

	// Open and read the config file
	file, err := os.Open(configPath)
	if err != nil {
		return "", ""
	}
	defer file.Close()

	// Parse the INI-style config file
	scanner := bufio.NewScanner(file)
	inDefaultsSection := false

	var organization, project string

	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())

		// Skip empty lines and comments
		if line == "" || strings.HasPrefix(line, "#") || strings.HasPrefix(line, ";") {
			continue
		}

		// Check for section headers
		if strings.HasPrefix(line, "[") && strings.HasSuffix(line, "]") {
			section := strings.ToLower(line[1 : len(line)-1])
			inDefaultsSection = (section == "defaults")
			continue
		}

		// Process key-value pairs in the defaults section
		if inDefaultsSection && strings.Contains(line, "=") {
			parts := strings.SplitN(line, "=", 2)
			key := strings.TrimSpace(parts[0])
			value := strings.TrimSpace(parts[1])

			if strings.ToLower(key) == "organization" {
				organization = value
			} else if strings.ToLower(key) == "project" {
				project = value
			}
		}
	}

	// Handle scanner errors
	if err := scanner.Err(); err != nil {
		return "", ""
	}

	return organization, project
}

// NewClient creates a new Azure DevOps client
func NewClient(config *Config) *Client {
	return &Client{
		Config: config,
	}
}

// runAzCommand executes an Azure CLI command and returns the output
func runAzCommand(args ...string) ([]byte, error) {
	cmd := execCommand("az", args...)

	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	err := cmd.Run()
	if err != nil {
		return nil, fmt.Errorf("az command failed: %v\nStderr: %s", err, stderr.String())
	}

	return stdout.Bytes(), nil
}

// FetchProjects retrieves projects using Azure CLI
func (c *Client) FetchProjects() ([]Project, error) {
	// Run the az devops project list command
	output, err := runAzCommand("devops", "project", "list", "--detect", "--output", "json")
	if err != nil {
		return nil, fmt.Errorf("error fetching projects: %v", err)
	}

	// Parse the CLI output
	var response AzCliProjectsResponse
	if err := json.Unmarshal(output, &response); err != nil {
		return nil, fmt.Errorf("error parsing CLI output: %v", err)
	}

	// Convert to our Project struct format
	projects := make([]Project, len(response.Value))
	for i, p := range response.Value {
		projects[i] = Project{
			ID:          p.ID,
			Name:        p.Name,
			Description: p.Description,
			URL:         p.URL,
			State:       p.State,
			LastUpdated: p.LastUpdateTime,
			Visibility:  p.Visibility,
		}
	}

	return projects, nil
}

// GetProject retrieves a specific project's details using Azure CLI
func (c *Client) GetProject(projectName string) (*Project, error) {
	// Run the az devops project show command
	output, err := runAzCommand("devops", "project", "show", "--project", projectName, "--detect", "--output", "json")
	if err != nil {
		return nil, fmt.Errorf("error fetching project '%s': %v", projectName, err)
	}

	// Parse the CLI output
	var response AzCliProjectResponse
	if err := json.Unmarshal(output, &response); err != nil {
		return nil, fmt.Errorf("error parsing CLI output: %v", err)
	}

	// Convert to our Project struct
	project := &Project{
		ID:          response.ID,
		Name:        response.Name,
		Description: response.Description,
		URL:         response.URL,
		State:       response.State,
		LastUpdated: response.LastUpdateTime,
		Visibility:  response.Visibility,
	}

	return project, nil
}

// GetWorkItemsForFilter retrieves work items for a given filter
func (c *Client) GetWorkItemsForFilter(filter string) ([]WorkItem, error) {
	// Get the work items for the filter
	var wiql string
	switch filter {
	case "me":
		wiql = workItemQueryMeSincePastMonth
	case "was-ever-me":
		wiql = workItemQueryWasEverMeSincePastMonth
	case "all":
		wiql = workItemsQueryAll
	default:
		wiql = workItemQueryMeSincePastMonth
	}

	output, err := runAzCommand("boards", "query", "--wiql", wiql, "--query", jmespathWorkItemQuery, "--output", "json")
	if err != nil {
		return nil, fmt.Errorf("error fetching work items: %v", err)
	}

	// Parse the output
	var workItems []WorkItem
	if err := json.Unmarshal(output, &workItems); err != nil {
		return nil, fmt.Errorf("error parsing work items: %v", err)
	}

	return workItems, nil
}

// GetWorkItemsAssignedToUser retrieves work items assigned to the current user
func (c *Client) GetWorkItemsAssignedToUser() ([]WorkItem, error) {
	output, err := runAzCommand("boards", "query", "--wiql", workItemQueryMeSincePastMonth, "--query", jmespathWorkItemQuery, "--output", "json")
	if err != nil {
		return nil, fmt.Errorf("error fetching work items: %v", err)
	}

	// Parse the output
	var workItems []WorkItem
	if err := json.Unmarshal(output, &workItems); err != nil {
		return nil, fmt.Errorf("error parsing work items: %v", err)
	}

	// Iterate over the work items and get more details
	// for _, workItem := range workItems {
	// 	_, err = workItem.GetMoreWorkItemDetails()
	// 	if err != nil {
	// 		return nil, fmt.Errorf("error fetching work item details: %v", err)
	// 	}
	// }

	return workItems, nil
}

// GetMoreWorkItemDetails retrieves the details of a specific work item
// Given a WorkItem, it will use the ID to fetch more details
func (c *WorkItem) GetMoreWorkItemDetails() (*WorkItem, error) {
	output, err := runAzCommand("boards", "work-item", "show", "--id", strconv.Itoa(c.ID), "--query", jmespathWorkItemDetailsQuery, "--output", "json")
	if err != nil {
		return nil, fmt.Errorf("error fetching work item details: %v", err)
	}

	// Parse the output
	var detail WorkItemDetails
	if err := json.Unmarshal(output, &detail); err != nil {
		return nil, fmt.Errorf("error parsing work item details: %v", err)
	}
	c.Details = &detail

	return c, nil
}

// Get URL of Work Item as it appears in the browser
func (c *WorkItem) GetURL(organization string, project string) string {
	return fmt.Sprintf("%s%s/_workitems/edit/%d", organization, project, c.ID)
}

// GetPRs retrieves the PRs associated with the work item
func (c *WorkItem) GetPRs() []string {
	// For each PR ref, get the last part of the URL when split by "%2F"
	prs := []string{}
	for _, prRef := range c.Details.PRRefs {
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

// Get the PR URL
func (pr *PullRequestDetails) GetURL() string {
	return fmt.Sprintf("%s/pullrequest/%d", pr.RepositoryURL, pr.ID)
}

// Retrieve PR details by PR ID
func (c *Client) GetPRDetails(prID string) (*PullRequestDetails, error) {
	output, err := runAzCommand("repos", "pr", "show", "--id", prID, "--query", jmespathPRDetailsQuery, "--output", "json")
	if err != nil {
		return nil, fmt.Errorf("error fetching PR details: %v", err)
	}

	// Parse the output
	var detail PullRequestDetails
	if err := json.Unmarshal(output, &detail); err != nil {
		return nil, fmt.Errorf("error parsing PR details: %v", err)
	}

	return &detail, nil
}
