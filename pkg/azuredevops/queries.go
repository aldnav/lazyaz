package azuredevops

const (
	// Work Items
	workItemQueryMeSincePastMonth        = `SELECT * FROM workitems WHERE [System.AssignedTo] = @me AND [System.CreatedDate] >= @Today - 90 ORDER BY [System.CreatedDate] DESC`
	workItemQueryWasEverMeSincePastMonth = `SELECT * FROM workitems WHERE EVER [System.AssignedTo] = @me AND [System.CreatedDate] >= @Today - 90 ORDER BY [System.CreatedDate] DESC`
	workItemsQueryAll                    = `SELECT * FROM workitems WHERE [System.CreatedDate] >= @Today - 90 ORDER BY [System.CreatedDate] DESC`
)

const jmespathWorkItemQuery = `[].{` +
	`Id: fields."System.Id", ` +
	`"Work Item Type": fields."System.WorkItemType", ` +
	`"Title": fields."System.Title", ` +
	`"Assigned To": fields."System.AssignedTo".displayName, ` +
	`"Assigned To Unique Name": fields."System.AssignedTo".uniqueName, ` +
	`"State": fields."System.State", ` +
	`"Tags": fields."System.Tags", ` +
	`"Iteration Path": fields."System.IterationPath", ` +
	`"CreatedDate": fields."System.CreatedDate", ` +
	`"CreatedBy": fields."System.CreatedBy".displayName, ` +
	`"ChangedDate": fields."System.ChangedDate", ` +
	`"ChangedBy": fields."System.ChangedBy".displayName, ` +
	`"Description": fields."System.Description"` +
	`}`
const jmespathWorkItemDetailsQuery = `{` +
	`"Repro Steps": fields."Microsoft.VSTS.TCM.ReproSteps", ` +
	`"System.AreaPath": fields."System.AreaPath", ` +
	`"Acceptance Criteria": fields."Microsoft.VSTS.Common.AcceptanceCriteria", ` +
	`"Board Column": fields."System.BoardColumn", ` +
	`"Board Column Done": fields."System.BoardColumnDone", ` +
	`"Comment Count": fields."System.CommentCount", ` +
	`"Latest Comment": fields."System.History", ` +
	`"PR refs": relations[?attributes.name=='Pull Request'].url, ` +
	`"Priority": fields."Microsoft.VSTS.Common.Priority", ` +
	`"Severity": fields."Microsoft.VSTS.Common.Severity"` +
	`}`
const jmespathPRDetailsQuery = `{` +
	`"Title": title, ` +
	`"Status": status, ` +
	`"ID": pullRequestId, ` +
	`"Author": createdBy.displayName, ` +
	`"Created Date": creationDate, ` +
	`"Description": description, ` +
	`"Is Draft": isDraft, ` +
	`"Labels": labels, ` +
	`"Merge Failure Message": mergeFailureMessage, ` +
	`"Merge Failure Type": mergeFailureType, ` +
	`"Merge Status": mergeStatus, ` +
	`"Repository": repository.name, ` +
	`"Repository URL": repository.webUrl, ` +
	`"Repository ApiURL": repository.url, ` +
	`"Project": repository.project.name, ` +
	`"Reviewers": reviewers[].displayName, ` +
	`"Reviewers Votes": reviewers[].vote, ` +
	`"Source Ref Name": sourceRefName, ` +
	`"Target Ref Name": targetRefName, ` +
	`"Work Item Refs": workItemRefs[].id, ` +
	`"Closed By": closedBy.displayName, ` +
	`"Closed Date": closedDate ` +
	`}`
const jmespathPRListsQuery = `[].` + jmespathPRDetailsQuery
const jmespathUserProfileQuery = `{"id": id, "displayName": displayName, "mail": mail, "givenName": givenName, "surname": surname}`
const jmespathPipelineDefinitionsQuery = `[].{id:id, name:name, path:path, status:queueStatus, defaultQueue:queue.name, project:project.name, author:authoredBy.displayName, authorUniqueName:authoredBy.uniqueName, pipelineType:type}`
const jmespathPipelineRunsQuery = `[].{id:id, buildNumber:buildNumber, definitionId: definition.id, definitionName: definition.name, definitionPath: definition.path, finishTime: finishTime, keepForever:keepForever, priority:priority, queue:queue.name, queueTime:queueTime, reason:reason, repository:repository.id, repositoryType:repository.type, requestedBy:requestedBy.displayName, requestedByUniqueName:requestedBy.uniqueName, requestedFor:requestedFor.displayName, requestedForUniqueName:requestedFor.uniqueName, result:result, sourceBranch:sourceBranch, sourceVersion:sourceVersion, startTime:startTime, status:status, logsUrl:logs.url, logsType:logs.type, retainedByRelease: retainedByRelease, deleted:deleted, deletedByd:deletedBy, deletedDate:deletedDate, deletedReason:deletedReason, projectId:project.id, projectUrl:project.url }`

// References:
// - https://learn.microsoft.com/en-us/azure/devops/boards/queries/query-operators-variables?view=azure-devops
// - https://learn.microsoft.com/en-us/azure/devops/boards/queries/wiql-syntax?view=azure-devops
