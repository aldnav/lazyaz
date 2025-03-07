package azuredevops

import (
	"encoding/json"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestClient_FetchProjects(t *testing.T) {
	// Setup mock server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Verify request headers
		authHeader := r.Header.Get("Authorization")
		if authHeader == "" {
			t.Error("Authorization header is missing")
			w.WriteHeader(http.StatusUnauthorized)
			return
		}

		// Verify request path
		if r.URL.Path != "/_apis/projects" {
			t.Errorf("Expected path /_apis/projects, got %s", r.URL.Path)
			w.WriteHeader(http.StatusNotFound)
			return
		}

		// Return mock response
		mockProjects := ProjectsResponse{
			Count: 2,
			Value: []Project{
				{
					ID:          "project1",
					Name:        "Test Project 1",
					Description: "Description for Test Project 1",
					URL:         "https://dev.azure.com/org/project1",
					State:       "wellFormed",
					LastUpdated: time.Now(),
				},
				{
					ID:          "project2",
					Name:        "Test Project 2",
					Description: "Description for Test Project 2",
					URL:         "https://dev.azure.com/org/project2",
					State:       "wellFormed",
					LastUpdated: time.Now(),
				},
			},
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(mockProjects)
	}))
	defer server.Close()

	// Create a test client that points to the mock server
	config := &Config{
		Organization: "testorg",
		Token:        "testtoken",
		Project:      "testproject",
	}
	client := NewClient(config)
	
	// Override the HTTP client to use the test server's URL
	client.HTTPClient = server.Client()
	
	// Replace the server URL in the FetchProjects method
	// Override the base URL for testing
	client.baseURL = server.URL

	// Execute the test
	projects, err := client.FetchProjects()

	// Verify the results
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	if len(projects) != 2 {
		t.Fatalf("Expected 2 projects, got %d", len(projects))
	}
}

func TestNewConfig(t *testing.T) {
	// Test with missing environment variables and no config file
	t.Run("Missing variables", func(t *testing.T) {
		os.Clearenv()
		
		config, err := NewConfig()
		
		if err == nil {
			t.Error("Expected error for missing configuration, got nil")
		}
		
		if config != nil {
			t.Errorf("Expected nil config, got %+v", config)
		}
	})
	
	// Test with valid environment variables
	t.Run("Valid variables", func(t *testing.T) {
		os.Clearenv()
		os.Setenv("AZURE_DEVOPS_ORG", "testorg")
		os.Setenv("AZURE_DEVOPS_TOKEN", "testtoken")
		os.Setenv("AZURE_DEVOPS_PROJECT", "testproject")
		
		config, err := NewConfig()
		
		if err != nil {
			t.Errorf("Expected no error, got %v", err)
		}
		
		if config == nil {
			t.Fatal("Expected config, got nil")
		}
		
		if config.Organization != "testorg" {
			t.Errorf("Expected organization 'testorg', got '%s'", config.Organization)
		}
		
		if config.Token != "testtoken" {
			t.Errorf("Expected token 'testtoken', got '%s'", config.Token)
		}
		
		if config.Project != "testproject" {
			t.Errorf("Expected project 'testproject', got '%s'", config.Project)
		}
	})
	
	// Test reading from config file with temporary mock file
	t.Run("Read organization and project from config file", func(t *testing.T) {
		// Create mock config file
		home, err := os.UserHomeDir()
		if err != nil {
			t.Skip("Unable to determine home directory, skipping test")
		}
		
		configDir := filepath.Join(home, ".azure", "azuredevops")
		err = os.MkdirAll(configDir, 0755)
		if err != nil {
			t.Skip("Unable to create config directory, skipping test")
		}
		
		configPath := filepath.Join(configDir, "config")
		
		// Create backup of existing file if it exists
		existingConfig := ""
		if _, err := os.Stat(configPath); err == nil {
			configContent, err := ioutil.ReadFile(configPath)
			if err == nil {
				existingConfig = string(configContent)
			}
		}
		
		// Write test config content with both organization and project
		configContent := `[defaults]
organization = configorg
project = configproject
`
		err = ioutil.WriteFile(configPath, []byte(configContent), 0644)
		if err != nil {
			t.Skip("Unable to write config file, skipping test")
		}
		
		// Clean up after test
		defer func() {
			if existingConfig != "" {
				ioutil.WriteFile(configPath, []byte(existingConfig), 0644)
			} else {
				os.Remove(configPath)
			}
		}()
		
		// Run the test with token from env variable but org from config
		os.Clearenv()
		os.Setenv("AZURE_DEVOPS_TOKEN", "testtoken")
		
		config, err := NewConfig()
		
		if err != nil {
			t.Errorf("Expected no error, got %v", err)
		}
		
		if config == nil {
			t.Fatal("Expected config, got nil")
		}
		
		if config.Organization != "configorg" {
			t.Errorf("Expected organization 'configorg' from config file, got '%s'", config.Organization)
		}
		
		if config.Project != "configproject" {
			t.Errorf("Expected project 'configproject' from config file, got '%s'", config.Project)
		}
		
		if config.Token != "testtoken" {
			t.Errorf("Expected token 'testtoken', got '%s'", config.Token)
		}
	})
	
	// Test reading only organization from config file (no project)
	t.Run("Read only organization from config file", func(t *testing.T) {
		// Create mock config file
		home, err := os.UserHomeDir()
		if err != nil {
			t.Skip("Unable to determine home directory, skipping test")
		}
		
		configDir := filepath.Join(home, ".azure", "azuredevops")
		err = os.MkdirAll(configDir, 0755)
		if err != nil {
			t.Skip("Unable to create config directory, skipping test")
		}
		
		configPath := filepath.Join(configDir, "config")
		
		// Create backup of existing file if it exists
		existingConfig := ""
		if _, err := os.Stat(configPath); err == nil {
			configContent, err := ioutil.ReadFile(configPath)
			if err == nil {
				existingConfig = string(configContent)
			}
		}
		
		// Write test config content with only organization
		configContent := `[defaults]
organization = configorg
`
		err = ioutil.WriteFile(configPath, []byte(configContent), 0644)
		if err != nil {
			t.Skip("Unable to write config file, skipping test")
		}
		
		// Clean up after test
		defer func() {
			if existingConfig != "" {
				ioutil.WriteFile(configPath, []byte(existingConfig), 0644)
			} else {
				os.Remove(configPath)
			}
		}()
		
		// Run the test with token and project from env variable
		os.Clearenv()
		os.Setenv("AZURE_DEVOPS_TOKEN", "testtoken")
		os.Setenv("AZURE_DEVOPS_PROJECT", "envproject")
		
		config, err := NewConfig()
		
		if err != nil {
			t.Errorf("Expected no error, got %v", err)
		}
		
		if config == nil {
			t.Fatal("Expected config, got nil")
		}
		
		if config.Organization != "configorg" {
			t.Errorf("Expected organization 'configorg' from config file, got '%s'", config.Organization)
		}
		
		if config.Project != "envproject" {
			t.Errorf("Expected project 'envproject' from env, got '%s'", config.Project)
		}
		
		if config.Token != "testtoken" {
			t.Errorf("Expected token 'testtoken', got '%s'", config.Token)
		}
	})
	
	// Test fallback to environment variable when config file doesn't have organization
	t.Run("Fallback to env variable", func(t *testing.T) {
		// Create mock config file
		home, err := os.UserHomeDir()
		if err != nil {
			t.Skip("Unable to determine home directory, skipping test")
		}
		
		configDir := filepath.Join(home, ".azure", "azuredevops")
		err = os.MkdirAll(configDir, 0755)
		if err != nil {
			t.Skip("Unable to create config directory, skipping test")
		}
		
		configPath := filepath.Join(configDir, "config")
		
		// Create backup of existing file if it exists
		existingConfig := ""
		if _, err := os.Stat(configPath); err == nil {
			configContent, err := ioutil.ReadFile(configPath)
			if err == nil {
				existingConfig = string(configContent)
			}
		}
		
		// Write test config content without organization
		configContent := `[defaults]
someotherkey = somevalue
project = configproject
`
		err = ioutil.WriteFile(configPath, []byte(configContent), 0644)
		if err != nil {
			t.Skip("Unable to write config file, skipping test")
		}
		
		// Clean up after test
		defer func() {
			if existingConfig != "" {
				ioutil.WriteFile(configPath, []byte(existingConfig), 0644)
			} else {
				os.Remove(configPath)
			}
		}()
		
		// Run the test with both from env variable 
		os.Clearenv()
		os.Setenv("AZURE_DEVOPS_ORG", "fallbackorg")
		os.Setenv("AZURE_DEVOPS_TOKEN", "testtoken")
		os.Setenv("AZURE_DEVOPS_PROJECT", "fallbackproject")
		
		config, err := NewConfig()
		
		if err != nil {
			t.Errorf("Expected no error, got %v", err)
		}
		
		if config == nil {
			t.Fatal("Expected config, got nil")
		}
		
		if config.Organization != "fallbackorg" {
			t.Errorf("Expected fallback to organization 'fallbackorg' from env, got '%s'", config.Organization)
		}
		
		if config.Project != "configproject" {
			t.Errorf("Expected project 'configproject' from config file, got '%s'", config.Project)
		}
		
		if config.Token != "testtoken" {
			t.Errorf("Expected token 'testtoken', got '%s'", config.Token)
		}
	})
	
	// Test using both org and project from environment variables when neither is in config
	t.Run("Both org and project from env variables", func(t *testing.T) {
		// Create mock config file
		home, err := os.UserHomeDir()
		if err != nil {
			t.Skip("Unable to determine home directory, skipping test")
		}
		
		configDir := filepath.Join(home, ".azure", "azuredevops")
		err = os.MkdirAll(configDir, 0755)
		if err != nil {
			t.Skip("Unable to create config directory, skipping test")
		}
		
		configPath := filepath.Join(configDir, "config")
		
		// Create backup of existing file if it exists
		existingConfig := ""
		if _, err := os.Stat(configPath); err == nil {
			configContent, err := ioutil.ReadFile(configPath)
			if err == nil {
				existingConfig = string(configContent)
			}
		}
		
		// Write test config content without organization or project
		configContent := `[defaults]
someotherkey = somevalue
`
		err = ioutil.WriteFile(configPath, []byte(configContent), 0644)
		if err != nil {
			t.Skip("Unable to write config file, skipping test")
		}
		
		// Clean up after test
		defer func() {
			if existingConfig != "" {
				ioutil.WriteFile(configPath, []byte(existingConfig), 0644)
			} else {
				os.Remove(configPath)
			}
		}()
		
		// Run the test with all values from env variables
		os.Clearenv()
		os.Setenv("AZURE_DEVOPS_ORG", "envorg")
		os.Setenv("AZURE_DEVOPS_TOKEN", "testtoken")
		os.Setenv("AZURE_DEVOPS_PROJECT", "envproject")
		
		config, err := NewConfig()
		
		if err != nil {
			t.Errorf("Expected no error, got %v", err)
		}
		
		if config == nil {
			t.Fatal("Expected config, got nil")
		}
		
		if config.Organization != "envorg" {
			t.Errorf("Expected organization 'envorg' from env, got '%s'", config.Organization)
		}
		
		if config.Project != "envproject" {
			t.Errorf("Expected project 'envproject' from env, got '%s'", config.Project)
		}
		
		if config.Token != "testtoken" {
			t.Errorf("Expected token 'testtoken', got '%s'", config.Token)
		}
	})
}

