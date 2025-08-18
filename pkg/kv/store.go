package kv

import (
	"encoding/json"
	"fmt"
	"strconv"
	"time"
)

type KeyValueStore interface {
	Set(key string, value interface{}, expiration time.Duration) error
	Delete(key string) error
	Exists(key string) (bool, error)
	Ping() error

	GetString(key string) (string, error)
	GetInt64(key string) (int64, error)
	GetFloat64(key string) (float64, error)
	GetBool(key string) (bool, error)
	GetJSON(key string, dest interface{}) error
}

type keyValueStore struct {
	provider KvProvider
}

func NewKeyValueStore(provider KvProvider) KeyValueStore {
	return &keyValueStore{
		provider: provider,
	}
}

func (s *keyValueStore) Set(key string, value interface{}, expiration time.Duration) error {
	return s.provider.Set(key, value, expiration)
}

func (s *keyValueStore) Delete(key string) error {
	return s.provider.Delete(key)
}

func (s *keyValueStore) Exists(key string) (bool, error) {
	return s.provider.Exists(key)
}

func (s *keyValueStore) Ping() error {
	return s.provider.Ping()
}

func (s *keyValueStore) GetString(key string) (string, error) {
	return s.provider.Get(key)
}

func (s *keyValueStore) GetInt64(key string) (int64, error) {
	val, err := s.provider.Get(key)
	if err != nil {
		return 0, err
	}

	if result, parseErr := strconv.ParseInt(val, 10, 64); parseErr == nil {
		return result, nil
	}

	var result int64
	if err := json.Unmarshal([]byte(val), &result); err != nil {
		return 0, fmt.Errorf("error parsing cached value as int64: %w", err)
	}

	return result, nil
}

func (s *keyValueStore) GetFloat64(key string) (float64, error) {
	val, err := s.provider.Get(key)
	if err != nil {
		return 0, err
	}

	if result, parseErr := strconv.ParseFloat(val, 64); parseErr == nil {
		return result, nil
	}

	var result float64
	if err := json.Unmarshal([]byte(val), &result); err != nil {
		return 0, fmt.Errorf("error parsing cached value as float64: %w", err)
	}

	return result, nil
}

func (s *keyValueStore) GetBool(key string) (bool, error) {
	val, err := s.provider.Get(key)
	if err != nil {
		return false, err
	}

	if result, parseErr := strconv.ParseBool(val); parseErr == nil {
		return result, nil
	}

	var result bool
	if err := json.Unmarshal([]byte(val), &result); err != nil {
		return false, fmt.Errorf("error parsing cached value as bool: %w", err)
	}

	return result, nil
}

func (s *keyValueStore) GetJSON(key string, dest interface{}) error {
	val, err := s.provider.Get(key)
	if err != nil {
		return err
	}

	if err := json.Unmarshal([]byte(val), dest); err != nil {
		return fmt.Errorf("error unmarshaling cached JSON: %w", err)
	}

	return nil
}
