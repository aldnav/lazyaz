package azuredevops

import (
	"errors"
	"os"
	"os/exec"
	"path/filepath"
	"testing"
)

// mockExecCommand is used to mock exec.Command for testing
func mockExecCommand(mockOutput string, mockError error) func(command string, args ...string) *exec.Cmd {
	return func(command string, args ...string) *exec.Cmd {
		cmd := exec.Command("echo", mockOutput)
		return cmd
	}
}

// mockExecCommandError is used to mock exec.Command for testing error scenarios
func mockExecCommandError(mockError error) func(command string, args ...string) *exec.Cmd {
	return func(command string, args ...string) *exec.Cmd {
		cmd := exec.Command("test")
	// This will cause the command to fail with the specified error
	cmd.Stderr = exec.Command("echo", mockError.Error()).Stdout
		return cmd
	}
}

func TestClient_FetchProjects(t *testing.T) {
	// Create mock response JSON
	mockResponseJSON := `{
		"count": 2,
		"value": [
			{
				"id": "project1",
				"name": "Test Project 1",
				"description": "Description for Test Project 1",
				"url": "https://dev.azure.com/testorg/project1",
				"state": "wellFormed",
				"visibility": "private",
				"lastUpdateTime": "2023-01-01T12:00:00Z"
			},
			{
				"id": "project2",
				"name": "Test Project 2",
				"description": "Description for Test Project 2",
				"url": "https://dev.azure.com/testorg/project2",
				"state": "wellFormed",
				"visibility": "private",
				"lastUpdateTime": "2023-01-01T12:00:00Z"
			}
		]
	}`

	// Store original exec.Command
	origExecCommand := execCommand
	defer func() { execCommand = origExecCommand }()

	// Mock the exec.Command to return our test data
	execCommand = mockExecCommand(mockResponseJSON, nil)

	// Create config and client
	config := &Config{
		Organization: "testorg",
		Project:      "testproject",
	}
	client := NewClient(config)

	// Execute the test
	projects, err := client.FetchProjects()

	// Verify the results
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	if len(projects) != 2 {
		t.Fatalf("Expected 2 projects, got %d", len(projects))
	}

	// Verify project details
	if projects[0].Name != "Test Project 1" {
		t.Errorf("Expected project name 'Test Project 1', got '%s'", projects[0].Name)
	}

	if projects[1].Name != "Test Project 2" {
		t.Errorf("Expected project name 'Test Project 2', got '%s'", projects[1].Name)
	}
}

func TestClient_FetchProjects_Error(t *testing.T) {
	// Store original exec.Command
	origExecCommand := execCommand
	defer func() { execCommand = origExecCommand }()

	// Mock the exec.Command to return an error
	execCommand = mockExecCommandError(errors.New("CLI execution failed"))

	// Create config and client
	config := &Config{
		Organization: "testorg",
		Project:      "testproject",
	}
	client := NewClient(config)

	// Execute the test
	projects, err := client.FetchProjects()

	// Verify the results
	if err == nil {
		t.Fatal("Expected error, got nil")
	}

	if projects != nil {
		t.Fatalf("Expected nil projects, got %+v", projects)
	}
}

func TestClient_GetProject(t *testing.T) {
	// Create mock response JSON
	mockResponseJSON := `{
		"id": "project1",
		"name": "Test Project 1",
		"description": "Description for Test Project 1",
		"url": "https://dev.azure.com/testorg/project1",
		"state": "wellFormed",
		"visibility": "private",
		"lastUpdateTime": "2023-01-01T12:00:00Z"
	}`

	// Store original exec.Command
	origExecCommand := execCommand
	defer func() { execCommand = origExecCommand }()

	// Mock the exec.Command to return our test data
	execCommand = mockExecCommand(mockResponseJSON, nil)

	// Create config and client
	config := &Config{
		Organization: "testorg",
		Project:      "testproject",
	}
	client := NewClient(config)

	// Execute the test
	project, err := client.GetProject("Test Project 1")

	// Verify the results
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	if project == nil {
		t.Fatal("Expected project, got nil")
	}

	if project.Name != "Test Project 1" {
		t.Errorf("Expected project name 'Test Project 1', got '%s'", project.Name)
	}

	if project.ID != "project1" {
		t.Errorf("Expected project ID 'project1', got '%s'", project.ID)
	}
}

func TestClient_GetProject_Error(t *testing.T) {
	// Store original exec.Command
	origExecCommand := execCommand
	defer func() { execCommand = origExecCommand }()

	// Mock the exec.Command to return an error
	execCommand = mockExecCommandError(errors.New("CLI execution failed"))

	// Create config and client
	config := &Config{
		Organization: "testorg",
		Project:      "testproject",
	}
	client := NewClient(config)

	// Execute the test
	project, err := client.GetProject("Test Project 1")

	// Verify the results
	if err == nil {
		t.Fatal("Expected error, got nil")
	}

	if project != nil {
		t.Fatalf("Expected nil project, got %+v", project)
	}
}

func TestNewConfig(t *testing.T) {
	// Test with missing environment variables and no config file
	t.Run("Missing variables", func(t *testing.T) {
		os.Clearenv()
		
		config, err := NewConfig()
		
		if err == nil {
			t.Error("Expected error for missing organization configuration, got nil")
		}
		
		if config != nil {
			t.Errorf("Expected nil config, got %+v", config)
		}
	})
	
	// Test with valid environment variables
	t.Run("Valid variables", func(t *testing.T) {
		os.Clearenv()
		os.Setenv("AZURE_DEVOPS_ORG", "testorg")
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
		
		if config.Project != "testproject" {
			t.Errorf("Expected project 'testproject', got '%s'", config.Project)
		}
	})
	
	// Helper function to set up and clean up config file for tests
	setupConfigTest := func(t *testing.T, configContent string) (string, func()) {
		// Get home directory
		home, err := os.UserHomeDir()
		if err != nil {
			t.Skipf("Unable to determine home directory: %v - skipping test", err)
			return "", func() {}
		}
		
		// Create config directory
		configDir := filepath.Join(home, ".azure", "azuredevops")
		if err = os.MkdirAll(configDir, 0755); err != nil {
			t.Skipf("Unable to create config directory: %v - skipping test", err)
			return "", func() {}
		}
		
		configPath := filepath.Join(configDir, "config")
		
		// Create backup of existing file if it exists
		existingConfig := ""
		if _, err := os.Stat(configPath); err == nil {
			configData, readErr := os.ReadFile(configPath)
			if readErr == nil {
				existingConfig = string(configData)
			} else {
				t.Logf("Warning: could not read existing config for backup: %v", readErr)
			}
		}
		
		// Write test config content
		if err = os.WriteFile(configPath, []byte(configContent), 0644); err != nil {
			t.Skipf("Unable to write config file: %v - skipping test", err)
			return "", func() {}
		}
		
		// Return cleanup function
		cleanup := func() {
			if existingConfig != "" {
				_ = os.WriteFile(configPath, []byte(existingConfig), 0644)
			} else {
				_ = os.Remove(configPath)
			}
		}
		
		return configPath, cleanup
	}

	// Test reading from config file with temporary mock file
	t.Run("Read organization and project from config file", func(t *testing.T) {
		// Create mock config file with both organization and project
		configContent := `[defaults]
organization = configorg
project = configproject
`
		_, cleanup := setupConfigTest(t, configContent)
		defer cleanup()
		
		
		// Run the test with token from env variable but org from config
		os.Clearenv()
		
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
	})
	
	// Test reading only organization from config file (no project)
	t.Run("Read only organization from config file", func(t *testing.T) {
		// Create mock config file with only organization
		configContent := `[defaults]
organization = configorg
`
		_, cleanup := setupConfigTest(t, configContent)
		defer cleanup()
		
		// Run the test with token and project from env variable
		os.Clearenv()
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
	})
	
	// Test fallback to environment variable when config file doesn't have organization
	t.Run("Fallback to env variable", func(t *testing.T) {
		// Create mock config file without organization but with project
		configContent := `[defaults]
someotherkey = somevalue
project = configproject
`
		_, cleanup := setupConfigTest(t, configContent)
		defer cleanup()
		
		// Run the test with both from env variable 
		os.Clearenv()
		os.Setenv("AZURE_DEVOPS_ORG", "fallbackorg")
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
	})
	
	// Test using both org and project from environment variables when neither is in config
	t.Run("Both org and project from env variables", func(t *testing.T) {
		// Create mock config file without organization or project
		configContent := `[defaults]
someotherkey = somevalue
`
		_, cleanup := setupConfigTest(t, configContent)
		defer cleanup()
		
		// Run the test with all values from env variables
		os.Clearenv()
		os.Setenv("AZURE_DEVOPS_ORG", "envorg")
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
	})
}

