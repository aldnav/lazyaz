package azuredevops

const (
	workItemQueryMeSincePastMonth        = `SELECT * FROM workitems WHERE [System.AssignedTo] = @me AND [System.CreatedDate] >= @Today - 90`
	workItemQueryWasEverMeSincePastMonth = `SELECT * FROM workitems WHERE EVER [System.AssignedTo] = @me AND [System.CreatedDate] >= @Today - 90`
)

const jmespathWorkItemQuery = `[].{Id: fields."System.Id", "Work Item Type": fields."System.WorkItemType", "Title": fields."System.Title", "Assigned To": fields."System.AssignedTo".displayName, "State": fields."System.State", "Tags": fields."System.Tags", "Iteration Path": fields."System.IterationPath", "CreatedDate": fields."System.CreatedDate", "CreatedBy": fields."System.CreatedBy".displayName, "ChangedDate": fields."System.ChangedDate", "ChangedBy": fields."System.ChangedBy".displayName, "Description": fields."System.Description"}`
const jmespathWorkItemDetailsQuery = `{"Repro Steps": fields."Microsoft.VSTS.TCM.ReproSteps", "System.AreaPath": fields."System.AreaPath"}`
