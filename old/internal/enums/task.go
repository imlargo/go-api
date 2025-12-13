package enums

type TaskStatus string

const (
	TaskStatusPending    TaskStatus = "pending"
	TaskStatusQueued     TaskStatus = "queued"
	TaskStatusProcessing TaskStatus = "processing"
	TaskStatusCompleted  TaskStatus = "completed"
	TaskStatusFailed     TaskStatus = "failed"
	TaskStatusCanceled   TaskStatus = "canceled"
)

func (t TaskStatus) String() string {
	return string(t)
}

type GenerationStatus string

const (
	GenerationStatusQueuing    GenerationStatus = "queuing"
	GenerationStatusProcessing GenerationStatus = "processing"
	GenerationStatusCompleted  GenerationStatus = "completed"
	GenerationStatusFailed     GenerationStatus = "failed"
	GenerationStatusPartial    GenerationStatus = "partial"
)

func (g GenerationStatus) String() string {
	return string(g)
}

type GenerationErrorCode string

const (
	GenerationErrorCodeNone                GenerationErrorCode = ""
	GenerationErrorCodeLimitExceeded       GenerationErrorCode = "limit_exceeded"
	GenerationErrorCodeNoContentAvailable  GenerationErrorCode = "no_content_available"
	GenerationErrorCodeNoContentFiles      GenerationErrorCode = "no_content_files"
	GenerationErrorCodeUnsupportedType     GenerationErrorCode = "unsupported_type"
	GenerationErrorCodeUnsupportedPlatform GenerationErrorCode = "unsupported_platform"
	GenerationErrorCodeUnknown             GenerationErrorCode = "unknown"
)

func (e GenerationErrorCode) String() string {
	return string(e)
}

type TaskPriority int

const (
	TaskPriorityLow    TaskPriority = 0
	TaskPriorityNormal TaskPriority = 5
	TaskPriorityHigh   TaskPriority = 10
)

func (p TaskPriority) Int() int {
	return int(p)
}

func (p TaskPriority) String() string {
	switch p {
	case TaskPriorityHigh:
		return "high"
	case TaskPriorityNormal:
		return "normal"
	case TaskPriorityLow:
		return "low"
	default:
		return "unknown"
	}
}

type SyncStatus string

const (
	SyncStatusSyncing   SyncStatus = "syncing"
	SyncStatusCompleted SyncStatus = "completed"
	SyncStatusFailed    SyncStatus = "failed"
	SyncStatusPartial   SyncStatus = "partial"
)

func (s SyncStatus) String() string {
	return string(s)
}
