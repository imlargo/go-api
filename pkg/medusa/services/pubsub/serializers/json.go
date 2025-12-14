// Package serializers provides various message serialization implementations
package serializers

import (
	"encoding/json"
)

// JSONSerializer implements JSON serialization
type JSONSerializer struct{}

func (s *JSONSerializer) Marshal(v interface{}) ([]byte, error) {
	return json.Marshal(v)
}

func (s *JSONSerializer) Unmarshal(data []byte, v interface{}) error {
	return json.Unmarshal(data, v)
}

func (s *JSONSerializer) ContentType() string {
	return "application/json"
}
