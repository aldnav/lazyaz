package main

import (
	"fmt"
	"slices"
)

// Registry holds all registered extensions
type Registry struct {
	Extensions map[string]ExtensionConfig
}

// NewRegistry creates a new registry instance
func NewRegistry() *Registry {
	return &Registry{
		Extensions: make(map[string]ExtensionConfig),
	}
}

// InitRegistry initializes the registry with extensions from the configuration
func InitRegistry() *Registry {
	registry := NewRegistry()

	// Find and load configuration
	appConfig, configPath, err := FindConfig()
	if err != nil {
		logger.Error("Error loading configuration", "error", err)
		return registry
	}
	logger.Debug(fmt.Sprintf("Loaded configuration from %s", configPath))

	// Register extensions from config
	if appConfig.Extensions != nil {
		for id, extension := range appConfig.Extensions {
			canInitialize, reason := extension.TryInitialize(id)
			if !canInitialize {
				logger.Warn(fmt.Sprintf("Skipping extension %s, cannot initialize: %s", id, reason))
				continue
			}
			entryPoint := extension.EntryPoint(id)
			if entryPoint == nil {
				// Try to find the entrypoint for each extension, if not found, skip the extension
				logger.Warn(fmt.Sprintf("Skipping extension %s, entry point not found", id))
				continue
			}
			registry.Extensions[id] = extension
			logger.Debug(fmt.Sprintf("Registered extension: %s - %s", id, extension.Name))
		}
	}

	return registry
}

// Get returns an extension by its ID
func (r *Registry) Get(id string) (ExtensionConfig, bool) {
	extension, exists := r.Extensions[id]
	return extension, exists
}

// GetFor returns a list of extensions that apply to the specified domain
func (r *Registry) GetFor(domain string) []ExtensionConfig {
	var result []ExtensionConfig
	var allowedDomains = []string{"workitems", "pullrequests", "pipelines"}
	if !slices.Contains(allowedDomains, domain) {
		return result
	}

	for _, extension := range r.Extensions {
		if slices.Contains(extension.AppliesTo, domain) {
			result = append(result, extension)
		}
	}

	return result
}
