package main

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
	"reflect"
	"strings"

	"github.com/BurntSushi/toml"
)

// AppConfig represents the application configuration
type AppConfig struct {
	WorkItems  WorkItemsConfig            `toml:"workitems"`
	Extensions map[string]ExtensionConfig `toml:"extensions"`
}

// WorkItemsConfig represents the configuration for work items
type WorkItemsConfig struct {
	Extensions []string `toml:"extensions"`
}

// ExtensionConfig represents the configuration for an extension
type ExtensionConfig struct {
	Name               string   `toml:"name"`
	Description        string   `toml:"description"`
	TemplatesDirectory string   `toml:"templates_directory"`
	AppliesTo          []string `toml:"applies_to"`
}

// EntryPoint returns the function corresponding to the extension ID
func (e ExtensionConfig) EntryPoint(id string) interface{} {
	// Convert from snake_case to CamelCase
	parts := strings.Split(strings.Replace(id, ".", "_", -1), "_")
	var camelCase string
	for _, part := range parts {
		if len(part) > 0 {
			camelCase += strings.ToUpper(part[:1]) + part[1:]
		}
	}

	// Use reflection to find the function dynamically
	// This allows for extension without hardcoding
	funcValue := reflect.ValueOf(ExportToTemplate)

	// In a more complete implementation, we would search for the function
	// in the package's exported functions using reflection

	return funcValue.Interface()
}

// LoadConfig loads the configuration from the specified file path
func LoadConfig(configPath string) (*AppConfig, error) {
	config := &AppConfig{}

	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		return config, fmt.Errorf("config file not found: %s", configPath)
	}

	if _, err := toml.DecodeFile(configPath, config); err != nil {
		return nil, fmt.Errorf("error decoding config file: %v", err)
	}

	return config, nil
}

// GetDefaultConfigPath returns the default path for the config file
func GetDefaultConfigPath() string {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		log.Printf("Error getting user home directory: %v", err)
		return "lazyaz.toml"
	}

	return filepath.Join(homeDir, ".lazyaz.toml")
}

// FindConfig looks for the config file in the current directory and home directory
func FindConfig() (*AppConfig, string, error) {
	// First, try the current directory
	currentDirConfig := "lazyaz.toml"
	if _, err := os.Stat(currentDirConfig); err == nil {
		config, err := LoadConfig(currentDirConfig)
		return config, currentDirConfig, err
	}

	// Then try the home directory
	homeConfig := GetDefaultConfigPath()
	if _, err := os.Stat(homeConfig); err == nil {
		config, err := LoadConfig(homeConfig)
		return config, homeConfig, err
	}

	// Return empty config if no config file found
	return &AppConfig{}, "", fmt.Errorf("no config file found")
}
