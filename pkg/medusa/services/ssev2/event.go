package ssev2

// Event represents a single SSE event
type Event struct {
	ID    string
	Event string
	Data  interface{}
	Retry int
}
