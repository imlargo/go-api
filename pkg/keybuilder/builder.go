package cachekey

import (
	"crypto/md5"
	"fmt"
	"sort"
	"strings"
)

// Builder defines the interface for building cache keys with various patterns
type Builder interface {
	Build(parts ...string) string
	BuildWithParams(prefix string, params map[string]interface{}) string
	BuildForEntity(entity, id string) string
	BuildForQuery(service, method string, params map[string]interface{}) string
	GetPrefix() string
	BuildPattern(parts ...string) string
}

// DefaultKeyBuilder implements KeyBuilder with configurable app name, version, and separator
type DefaultKeyBuilder struct {
	appName      string // Application name (normalized to lowercase)
	separator    string // Character used to separate key parts
	version      string // Application version (normalized to lowercase)
	maxParamSize int    // Maximum length for parameter strings before hashing
}

// NewBuilder creates a new DefaultKeyBuilder with default configuration
func NewBuilder(appName, version string) Builder {
	return &DefaultKeyBuilder{
		appName:      strings.ToLower(appName),
		separator:    ":",
		version:      strings.ToLower(version),
		maxParamSize: 100,
	}
}

// Build constructs a cache key from the provided parts, prefixed with app name and version
func (kb *DefaultKeyBuilder) Build(parts ...string) string {
	allParts := []string{kb.appName, kb.version}
	allParts = append(allParts, parts...)

	var filteredParts []string
	for _, part := range allParts {
		if cleaned := strings.TrimSpace(part); cleaned != "" {
			filteredParts = append(filteredParts, kb.sanitize(cleaned))
		}
	}

	return strings.Join(filteredParts, kb.separator)
}

// BuildWithParams creates a cache key with parameters, hashing long parameter strings
func (kb *DefaultKeyBuilder) BuildWithParams(prefix string, params map[string]interface{}) string {
	if len(params) == 0 {
		return kb.Build(prefix)
	}

	paramsStr := kb.buildParamsString(params)

	// Hash parameter string if it exceeds maximum size
	if len(paramsStr) > kb.maxParamSize {
		paramsStr = fmt.Sprintf("hash_%x", md5.Sum([]byte(paramsStr)))
	}

	return kb.Build(prefix, paramsStr)
}

// BuildForEntity creates a standardized cache key for entities
func (kb *DefaultKeyBuilder) BuildForEntity(entity, id string) string {
	return kb.Build("entity", entity, id)
}

// BuildForQuery creates a cache key for service method calls with parameters
func (kb *DefaultKeyBuilder) BuildForQuery(service, method string, params map[string]interface{}) string {
	prefix := strings.Join([]string{"query", service, method}, kb.separator)
	return kb.BuildWithParams(prefix, params)
}

// GetPrefix returns the base prefix (app name and version) for all keys
func (kb *DefaultKeyBuilder) GetPrefix() string {
	return strings.Join([]string{kb.appName, kb.version}, kb.separator)
}

// BuildPattern creates a wildcard pattern for cache key matching
func (kb *DefaultKeyBuilder) BuildPattern(parts ...string) string {
	key := kb.Build(parts...)
	return key + kb.separator + "*"
}

// buildParamsString converts parameter map to a sorted, deterministic string representation
func (kb *DefaultKeyBuilder) buildParamsString(params map[string]interface{}) string {
	if len(params) == 0 {
		return ""
	}

	// Sort keys for consistent ordering
	keys := make([]string, 0, len(params))
	for k := range params {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	var paramPairs []string
	for _, k := range keys {
		value := fmt.Sprintf("%v", params[k])
		paramPairs = append(paramPairs, fmt.Sprintf("%s=%s", kb.sanitize(k), kb.sanitize(value)))
	}

	return strings.Join(paramPairs, "&")
}

// sanitize cleans strings for use in cache keys by removing/replacing problematic characters
func (kb *DefaultKeyBuilder) sanitize(s string) string {
	s = strings.TrimSpace(s)
	s = strings.ReplaceAll(s, " ", "_")
	s = strings.ReplaceAll(s, kb.separator, "_")
	s = strings.ReplaceAll(s, "*", "_")
	return strings.ToLower(s)
}
