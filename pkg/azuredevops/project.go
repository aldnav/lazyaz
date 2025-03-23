package azuredevops

import "time"

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
