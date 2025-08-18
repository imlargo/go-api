package cache

import (
	"encoding/json"
	"fmt"
	"strconv"
	"time"
)

type CacheService interface {
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

type cacheServiceImpl struct {
	cacheRepository CacheRepository
}

func NewCacheService(cacheRepository CacheRepository) CacheService {
	return &cacheServiceImpl{
		cacheRepository: cacheRepository,
	}
}

func (s *cacheServiceImpl) Set(key string, value interface{}, expiration time.Duration) error {
	return s.cacheRepository.Set(key, value, expiration)
}

func (s *cacheServiceImpl) Delete(key string) error {
	return s.cacheRepository.Delete(key)
}

func (s *cacheServiceImpl) Exists(key string) (bool, error) {
	return s.cacheRepository.Exists(key)
}

func (s *cacheServiceImpl) Ping() error {
	return s.cacheRepository.Ping()
}

func (s *cacheServiceImpl) GetString(key string) (string, error) {
	return s.cacheRepository.Get(key)
}

func (s *cacheServiceImpl) GetInt64(key string) (int64, error) {
	val, err := s.cacheRepository.Get(key)
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

func (s *cacheServiceImpl) GetFloat64(key string) (float64, error) {
	val, err := s.cacheRepository.Get(key)
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

func (s *cacheServiceImpl) GetBool(key string) (bool, error) {
	val, err := s.cacheRepository.Get(key)
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

func (s *cacheServiceImpl) GetJSON(key string, dest interface{}) error {
	val, err := s.cacheRepository.Get(key)
	if err != nil {
		return err
	}

	if err := json.Unmarshal([]byte(val), dest); err != nil {
		return fmt.Errorf("error unmarshaling cached JSON: %w", err)
	}

	return nil
}
