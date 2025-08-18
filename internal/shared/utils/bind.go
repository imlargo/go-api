package utils

import (
	"bytes"
	"encoding/json"
	"fmt"
	"reflect"
)

// MapToStructStrict converts a map to a struct with strict validation.
// It ensures that the map keys match the struct fields exactly and does not allow unknown fields.
// If the map contains keys that do not correspond to any field in the struct, an error
// will be returned.
func MapToStructStrict(data map[string]interface{}, result interface{}) error {
	// Initial checks
	if reflect.TypeOf(result).Kind() != reflect.Ptr {
		return fmt.Errorf("result must be a pointer to struct")
	}

	if reflect.ValueOf(result).Elem().Kind() != reflect.Struct {
		return fmt.Errorf("result must point to a struct")
	}

	// Convert map to JSON
	jsonData, err := json.Marshal(data)
	if err != nil {
		return fmt.Errorf("error converting map to JSON: %w", err)
	}

	// Create decoder with strict validation
	decoder := json.NewDecoder(bytes.NewReader(jsonData))
	decoder.DisallowUnknownFields()

	// Convert JSON to struct
	err = decoder.Decode(result)
	if err != nil {
		return fmt.Errorf("error converting JSON to struct (strict mode): %w", err)
	}

	return nil
}

// MapToStruct converts a map to a struct without strict validation.
// It allows the map to contain keys that do not correspond to any field in the struct.
// If the map contains keys that do not match any field in the struct, those keys will be ignored.
// Unknown fields in the map are ignored rather than causing an error.
// This is useful when you want to convert a map to a struct but do not require strict validation.
// It is less strict than MapToStructStrict and will not return an error for unknown fields.
func MapToStruct(data map[string]interface{}, result interface{}) error {
	// Initial checks
	if reflect.TypeOf(result).Kind() != reflect.Ptr {
		return fmt.Errorf("result must be a pointer to struct")
	}

	if reflect.ValueOf(result).Elem().Kind() != reflect.Struct {
		return fmt.Errorf("result must point to a struct")
	}

	// Convert map to JSON
	jsonData, err := json.Marshal(data)
	if err != nil {
		return fmt.Errorf("error converting map to JSON: %w", err)
	}

	// Decode JSON into struct
	err = json.Unmarshal(jsonData, result)
	if err != nil {
		return fmt.Errorf("error converting JSON to struct: %w", err)
	}

	return nil
}
