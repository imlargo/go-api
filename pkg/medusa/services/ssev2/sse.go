package ssev2

import (
	"encoding/json"
	"fmt"
	"io"
)

// Writer handles SSE protocol writing
type Writer struct {
	w io.Writer
}

// NewWriter creates a new SSE writer
func NewWriter(w io.Writer) *Writer {
	return &Writer{w: w}
}

// Write writes an event in SSE format
func (w *Writer) Write(event Event) error {
	if event.ID != "" {
		if _, err := fmt.Fprintf(w.w, "id: %s\n", event.ID); err != nil {
			return err
		}
	}

	if event.Event != "" {
		if _, err := fmt.Fprintf(w.w, "event: %s\n", event.Event); err != nil {
			return err
		}
	}

	if event.Retry > 0 {
		if _, err := fmt.Fprintf(w.w, "retry: %d\n", event.Retry); err != nil {
			return err
		}
	}

	// Marshal data
	var dataStr string
	switch v := event.Data.(type) {
	case string:
		dataStr = v
	case []byte:
		dataStr = string(v)
	default:
		jsonData, err := json.Marshal(v)
		if err != nil {
			return err
		}
		dataStr = string(jsonData)
	}

	if _, err := fmt.Fprintf(w.w, "data: %s\n\n", dataStr); err != nil {
		return err
	}

	// Flush if possible
	if flusher, ok := w.w.(interface{ Flush() }); ok {
		flusher.Flush()
	}

	return nil
}
