package azuredevops

import (
	"bufio"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"
)

// Config holds the Azure DevOps connection settings
type Config struct {
	Organization string
	Token        string
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
}

// ProjectsResponse represents the API response for projects
type ProjectsResponse struct {
	Count   int       `json:"count"`
	Value   []Project `json:"value"`
	Message string    `json:"message,omitempty"`
}

// Client represents an Azure DevOps API client
type Client struct {
	Config     *Config
	HTTPClient *http.Client
	baseURL    string // Added for testing
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
	
	// Always read token from environment variable
	token := os.Getenv("AZURE_DEVOPS_TOKEN")

	var missingVars []string
	if org == "" {
		missingVars = append(missingVars, "AZURE_DEVOPS_ORG")
	}
	if token == "" {
		missingVars = append(missingVars, "AZURE_DEVOPS_TOKEN")
	}

	if len(missingVars) > 0 {
		return nil, fmt.Errorf("missing required configuration: %s", strings.Join(missingVars, ", "))
	}

	return &Config{
		Organization: org,
		Token:        token,
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
		HTTPClient: &http.Client{
			Timeout: 10 * time.Second,
		},
		baseURL: fmt.Sprintf("https://dev.azure.com/%s", config.Organization),
	}
}

// FetchProjects retrieves projects from Azure DevOps API
func (c *Client) FetchProjects() ([]Project, error) {
	// Prepare the API URL
	baseURL := c.baseURL
	if baseURL == "" {
		baseURL = fmt.Sprintf("https://dev.azure.com/%s", c.Config.Organization)
	}
	url := fmt.Sprintf("%s/_apis/projects?api-version=7.0", baseURL)

	// Create a new request
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("error creating request: %v", err)
	}

	// Set authorization header using Basic Authentication with PAT
	// Username is empty, password is the PAT
	auth := base64.StdEncoding.EncodeToString([]byte(":" + c.Config.Token))
	req.Header.Add("Authorization", "Basic "+auth)

	// Execute the request
	resp, err := c.HTTPClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("error connecting to Azure DevOps API: %v", err)
	}
	defer resp.Body.Close()

	// Read response body
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("error reading response: %v", err)
	}

	// Check for HTTP error
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("API request failed with status %s: %s", resp.Status, string(body))
	}

	// Parse response JSON
	var projectsResp ProjectsResponse
	if err := json.Unmarshal(body, &projectsResp); err != nil {
		return nil, fmt.Errorf("error parsing JSON response: %v", err)
	}

	return projectsResp.Value, nil
}

