package alias

// EntityId is a base type
type EntityId string

// JobId is a type alias (not a new type)
type JobId = EntityId

// TaskId is another type alias
type TaskId = EntityId

// AliasInterface uses type aliases
type AliasInterface interface {
	GetJob(jobId JobId) (string, error)
	GetTask(taskId TaskId) error
	Process(jobId JobId, taskId TaskId) (EntityId, error)
}

