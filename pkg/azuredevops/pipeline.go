package azuredevops

import (
	"fmt"
	"strings"
	"time"
)

type Pipeline struct {
	ID               int    `json:"id"`
	Name             string `json:"name"`
	Path             string `json:"path"`
	Status           string `json:"status"`
	DefaultQueue     string `json:"defaultQueue"`
	Project          string `json:"project"`
	Author           string `json:"author"`
	AuthorUniqueName string `json:"authorUniqueName"`
	PipelineType     string `json:"pipelineType"`
}

type PipelineRun struct {
	ID                     int       `json:"id"`
	BuildNumber            string    `json:"buildNumber"`
	DefinitionID           int       `json:"definitionId"`
	DefinitionName         string    `json:"definitionName"`
	DefinitionPath         string    `json:"definitionPath"`
	Deleted                bool      `json:"deleted"`
	DeletedBy              string    `json:"deletedBy"`
	DeletedDate            time.Time `json:"deletedDate"`
	DeletedReason          string    `json:"deletedReason"`
	FinishTime             time.Time `json:"finishTime"`
	KeepForever            bool      `json:"keepForever"`
	LogsURL                string    `json:"logsUrl"`
	LogsType               string    `json:"logsType"`
	Priority               string    `json:"priority"`
	Queue                  string    `json:"queue"`
	QueueTime              time.Time `json:"queueTime"`
	ProjectID              string    `json:"projectId"`
	ProjectURL             string    `json:"projectUrl"`
	Reason                 string    `json:"reason"`
	Repository             string    `json:"repository"`
	RepositoryType         string    `json:"repositoryType"`
	RequestedBy            string    `json:"requestedBy"`
	RequestedByUniqueName  string    `json:"requestedByUniqueName"`
	RequestedFor           string    `json:"requestedFor"`
	RequestedForUniqueName string    `json:"requestedForUniqueName"`
	RetainedByRelease      bool      `json:"retainedByRelease"`
	Result                 string    `json:"result"`
	SourceBranch           string    `json:"sourceBranch"`
	SourceVersion          string    `json:"sourceVersion"`
	StartTime              time.Time `json:"startTime"`
	Status                 string    `json:"status"`
}

func (r *PipelineRun) GetWebURL() string {
	baseURL := strings.Split(r.ProjectURL, "_apis")[0] + r.ProjectID
	return fmt.Sprintf("%s/_build/results?buildId=%d", baseURL, r.ID)
}
