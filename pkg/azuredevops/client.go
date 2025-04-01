package azuredevops

import (
	"bufio"
	"bytes"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"slices"
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

type UserProfile struct {
	DisplayName string `json:"displayName"`
	ID          string `json:"id"`
	Mail        string `json:"mail"`
	GivenName   string `json:"givenName"`
	Surname     string `json:"surname"`
	Username    string `json:"-"` // Without the email domain
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
		// Try this file instead
		configPath = filepath.Join(home, ".azure", "config")
		if _, err := os.Stat(configPath); os.IsNotExist(err) {
			return "", ""
		}
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

func _fetchPRDetails(prID string) (*PullRequestDetails, error) {
	output, err := runAzCommand("repos", "pr", "show", "--id", prID, "--query", jmespathPRDetailsQuery, "--output", "json")
	if err != nil {
		return nil, fmt.Errorf("error fetching PR details: %v", err)
	}

	// Parse the output
	var detail PullRequestDetails
	if err := json.Unmarshal(output, &detail); err != nil {
		return nil, fmt.Errorf("error parsing PR details: %v", err)
	}
	detail.IsDetailFetched = true
	return &detail, nil
}

// Retrieve PR details by PR ID
func (c *Client) GetPRDetails(prID string) (*PullRequestDetails, error) {
	return _fetchPRDetails(prID)
}

// Retrieve current user profile
func (c *Client) GetUserProfile() (*UserProfile, error) {
	output, err := runAzCommand("ad", "signed-in-user", "show", "--query", jmespathUserProfileQuery, "--output", "json")
	if err != nil {
		return nil, fmt.Errorf("error fetching user profile: %v", err)
	}

	// Parse the output
	var profile UserProfile
	if err := json.Unmarshal(output, &profile); err != nil {
		return nil, fmt.Errorf("error parsing user profile: %v", err)
	}

	if profile.Mail == "" {
		log.Println("[WARNING] User profile is empty from the ActiveDirectory profile! This is required for most operations. Please contact the administrator to fix this or setup your profile.")
		log.Println("Trying to fetch from the command: ")
		var _query = `user.name`
		log.Println("az account show --query " + _query + " --output tsv")
		output, err = runAzCommand("account", "show", "--query", _query, "--output", "tsv")
		if err != nil {
			log.Println("[ERROR] Failed to fetch user profile from az account show command")
		} else if len(output) > 0 {
			profile.Mail = strings.TrimSuffix(string(output), "\n")
			profile.Username = strings.Split(profile.Mail, "@")[0]
		}
	}

	if profile.Mail == "" {
		log.Println("[WARNING] User profile is empty! This is required for most operations. Please contact the administrator to fix this or setup your profile.")
		return nil, fmt.Errorf("cannot continue without a user profile mail")
	}

	return &profile, nil
}

// TODO Break up into multiple files
var PRStatuses = []string{"active", "abandoned", "completed", "all"}

// Get PRs created by the current user (by default these are opened PRs)
func (c *Client) GetPRsCreatedByUser(user string, status string) ([]PullRequestDetails, error) {
	if user == "" {
		return nil, fmt.Errorf("user is required")
	}
	cmdParams := []string{"repos", "pr", "list", "--include-links", "--creator", user, "--query", jmespathPRListsQuery, "--output", "json"}
	if status != "" && slices.Contains(PRStatuses, status) {
		cmdParams = append(cmdParams, "--status", status)
	}
	output, err := runAzCommand(cmdParams...)
	if err != nil {
		return nil, fmt.Errorf("error fetching PRs: %v", err)
	}

	// Parse the output
	var prs []PullRequestDetails
	if err := json.Unmarshal(output, &prs); err != nil {
		return nil, fmt.Errorf("error parsing PRs: %v", err)
	}

	return prs, nil
}

// Get PRs assigned to the current user
func (c *Client) GetPRsAssignedToUser(user string) ([]PullRequestDetails, error) {
	if user == "" {
		return nil, fmt.Errorf("user is required")
	}
	output, err := runAzCommand("repos", "pr", "list", "--include-links", "--reviewer", user, "--status", "active", "--top", "100", "--query", jmespathPRListsQuery, "--output", "json")
	if err != nil {
		return nil, fmt.Errorf("error fetching PRs: %v", err)
	}

	// Parse the output
	var prs []PullRequestDetails
	if err := json.Unmarshal(output, &prs); err != nil {
		return nil, fmt.Errorf("error parsing PRs: %v", err)
	}

	return prs, nil
}

func (c *Client) FetchPullRequestsByStatus(status string) ([]PullRequestDetails, error) {
	if status != "" && !slices.Contains(PRStatuses, status) {
		return nil, fmt.Errorf("invalid status: %s", status)
	}
	cmdParams := []string{"repos", "pr", "list", "--include-links", "--status", status, "--query", jmespathPRListsQuery, "--output", "json", "--top", "100"}
	output, err := runAzCommand(cmdParams...)
	if err != nil {
		return nil, fmt.Errorf("error fetching PRs: %v", err)
	}

	// Parse the output
	var prs []PullRequestDetails
	if err := json.Unmarshal(output, &prs); err != nil {
		return nil, fmt.Errorf("error parsing PRs: %v", err)
	}

	return prs, nil
}

func (c *Client) GetAllPRs() ([]PullRequestDetails, error) {
	return c.FetchPullRequestsByStatus("all")
}

func (c *Client) GetActivePRs() ([]PullRequestDetails, error) {
	return c.FetchPullRequestsByStatus("active")
}

func (c *Client) GetCompletedPRs() ([]PullRequestDetails, error) {
	return c.FetchPullRequestsByStatus("completed")
}

func (c *Client) GetAbandonedPRs() ([]PullRequestDetails, error) {
	return c.FetchPullRequestsByStatus("abandoned")
}

// Pipeline functions
func (c *Client) GetPipelineDefinitions() ([]Pipeline, error) {
	output, err := runAzCommand("pipelines", "list", "--query", jmespathPipelineDefinitionsQuery, "--output", "json")
	if err != nil {
		return nil, fmt.Errorf("error fetching pipeline definitions: %v", err)
	}

	// Parse the output
	var pipelines []Pipeline
	if err := json.Unmarshal(output, &pipelines); err != nil {
		return nil, fmt.Errorf("error parsing pipelines: %v", err)
	}
	return pipelines, nil
}

func (c *Client) GetPipelineRuns() ([]PipelineRun, error) {
	output, err := runAzCommand("pipelines", "runs", "list", "--query", jmespathPipelineRunsQuery, "--output", "json", "--top", "40")
	if err != nil {
		return nil, fmt.Errorf("error fetching pipeline runs: %v", err)
	}

	// Parse the output
	var runs []PipelineRun
	if err := json.Unmarshal(output, &runs); err != nil {
		return nil, fmt.Errorf("error parsing pipeline runs: %v", err)
	}
	return runs, nil
}

var pipelineRunsAllowedReasons = []string{
	"all",
	"batchedCI",
	"buildCompletion",
	"checkInShelveset",
	"individualCI",
	"manual",
	"pullRequest",
	"schedule",
	"triggered",
	"userCreated",
	"validateShelveset",
}

var pipelineRunsAllowedResults = []string{
	"all",
	"canceled",
	"failed",
	"none",
	"partiallySucceeded",
	"succeeded",
}

var pipelineRunsAllowedStatuses = []string{
	"all",
	"cancelling",
	"completed",
	"inProgress",
	"none",
	"notStarted",
	"postponed",
}

func (c *Client) GetPipelineRunsFiltered(
	pipelineID int,
	branch string,
	reason string,
	result string,
	status string,
	requestedFor string,
) ([]PipelineRun, error) {
	cmdParams := []string{"pipelines", "runs", "list", "--query", jmespathPipelineRunsQuery, "--output", "json", "--top", "40"}
	if pipelineID != 0 {
		cmdParams = append(cmdParams, "--pipeline-ids", strconv.Itoa(pipelineID))
	}
	if branch != "" {
		cmdParams = append(cmdParams, "--branch", branch)
	}
	if reason != "" {
		if !slices.Contains(pipelineRunsAllowedReasons, reason) {
			return nil, fmt.Errorf("invalid reason: %s", reason)
		}
		cmdParams = append(cmdParams, "--reason", reason)
	}
	if result != "" {
		if !slices.Contains(pipelineRunsAllowedResults, result) {
			return nil, fmt.Errorf("invalid result: %s", result)
		}
		if result != "all" {
			cmdParams = append(cmdParams, "--result", result)
		}
	}
	if status != "" {
		if !slices.Contains(pipelineRunsAllowedStatuses, status) {
			return nil, fmt.Errorf("invalid status: %s", status)
		}
		cmdParams = append(cmdParams, "--status", status)
	}
	if requestedFor != "" {
		cmdParams = append(cmdParams, "--requested-for", requestedFor)
	}
	output, err := runAzCommand(cmdParams...)
	if err != nil {
		return nil, fmt.Errorf("error fetching pipeline runs: %v", err)
	}

	// Parse the output
	var runs []PipelineRun
	if err := json.Unmarshal(output, &runs); err != nil {
		return nil, fmt.Errorf("error parsing pipeline runs: %v", err)
	}
	return runs, nil
}
