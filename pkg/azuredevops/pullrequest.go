package azuredevops

import (
	"fmt"
	"sort"
	"strconv"
	"strings"
	"time"
)

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
	RepositoryApiURL    string      `json:"Repository ApiURL"`
	Project             string      `json:"Project"`
	Reviewers           []string    `json:"Reviewers"`
	ReviewersVotes      []int       `json:"Reviewers Votes"`
	SourceRefName       string      `json:"Source Ref Name"`
	Status              string      `json:"Status"`
	TargetRefName       string      `json:"Target Ref Name"`
	Title               string      `json:"Title"`
	WorkItemRefs        []string    `json:"Work Item Refs"`
	IsDetailFetched     bool        `json:"-"`
}

// Get the PR URL
func (pr *PullRequestDetails) GetURL() string {
	return fmt.Sprintf("%s/pullrequest/%d", pr.RepositoryURL, pr.ID)
}

// Get PR URL given organization
// Used when RepositoryURL is not available
func (pr *PullRequestDetails) GetOrgURL(organization string) string {
	return fmt.Sprintf("%s%s/_git/%s/pullrequest/%d", organization, pr.Project, pr.Repository, pr.ID)
}

// Get number of approvals
func (pr *PullRequestDetails) GetApprovals() int {
	approvals := 0
	for _, vote := range pr.ReviewersVotes {
		if vote == 10 || vote == 5 {
			approvals++
		}
	}
	return approvals
}

// Get shortened branch name with refs/heads/
func (pr *PullRequestDetails) GetShortBranchName() string {
	return strings.TrimPrefix(pr.SourceRefName, "refs/heads/")
}

// Get shortened branch name with refs/heads/
func (pr *PullRequestDetails) GetShortTargetBranchName() string {
	return strings.TrimPrefix(pr.TargetRefName, "refs/heads/")
}

type VoteInfo struct {
	Reviewer    string
	Description string
	Value       int
}

// Get the votes info
func (pr *PullRequestDetails) GetVotesInfo() []VoteInfo {
	// Ref: https://learn.microsoft.com/en-us/rest/api/azure/devops/git/pull-request-reviewers/create-pull-request-reviewer?view=azure-devops-rest-6.0&tabs=HTTP
	// 10 - approved 5 - approved with suggestions 0 - no vote -5 - waiting for author -10 - rejected
	// Map reviewers with votes
	voteIdxMap := map[int]string{
		10:  "approved",
		5:   "approved with suggestions",
		0:   "no vote",
		-5:  "waiting for author",
		-10: "rejected",
	}
	// Convert map to slice for sorting
	votes := make([]VoteInfo, 0, len(pr.Reviewers))
	for idx, reviewer := range pr.Reviewers {
		votes = append(votes, VoteInfo{
			Reviewer:    reviewer,
			Description: voteIdxMap[pr.ReviewersVotes[idx]],
			Value:       pr.ReviewersVotes[idx],
		})
	}
	// Sort slice by vote value in descending order, with secondary sort by reviewer name
	sort.Slice(votes, func(i, j int) bool {
		if votes[i].Value != votes[j].Value {
			return votes[i].Value > votes[j].Value
		}
		return votes[i].Reviewer < votes[j].Reviewer
	})
	return votes
}

// Retrieve more details from the Pull Request itself
func (pr *PullRequestDetails) GetMorePRDetails() (*PullRequestDetails, error) {
	_shallowPR, _ := _fetchPRDetails(strconv.Itoa(pr.ID))

	pr.IsDetailFetched = true
	pr.Description = _shallowPR.Description
	pr.RepositoryURL = _shallowPR.RepositoryURL
	pr.WorkItemRefs = _shallowPR.WorkItemRefs
	return pr, nil
}
