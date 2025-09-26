package handlers

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
	"go.uber.org/zap"
)

type HealthHandlerTestSuite struct {
	suite.Suite
	router   *gin.Engine
	recorder *httptest.ResponseRecorder
}

type mockRedis struct {
	shouldFail bool
}

func (m *mockRedis) Ping() error {
	if m.shouldFail {
		return assert.AnError
	}
	return nil
}

func (suite *HealthHandlerTestSuite) SetupTest() {
	gin.SetMode(gin.TestMode)
	suite.router = gin.New()
	suite.recorder = httptest.NewRecorder()

	logger := zap.NewNop().Sugar()
	handler := NewHandler(logger)
	
	// Setup with working dependencies
	redis := &mockRedis{shouldFail: false}
	healthHandler := NewHealthHandler(handler, nil, redis) // DB is nil for basic test
	
	suite.router.GET("/health", healthHandler.Health)
	suite.router.GET("/ready", healthHandler.Readiness)
	suite.router.GET("/live", healthHandler.Liveness)
}

func (suite *HealthHandlerTestSuite) TestLivenessEndpoint() {
	req, err := http.NewRequest("GET", "/live", nil)
	suite.NoError(err)

	suite.router.ServeHTTP(suite.recorder, req)

	suite.Equal(http.StatusOK, suite.recorder.Code)
	
	var response map[string]string
	err = json.Unmarshal(suite.recorder.Body.Bytes(), &response)
	suite.NoError(err)
	suite.Equal("alive", response["status"])
	suite.NotEmpty(response["timestamp"])
}

func (suite *HealthHandlerTestSuite) TestReadinessEndpoint() {
	req, err := http.NewRequest("GET", "/ready", nil)
	suite.NoError(err)

	suite.recorder = httptest.NewRecorder()
	suite.router.ServeHTTP(suite.recorder, req)

	suite.Equal(http.StatusOK, suite.recorder.Code)
	
	var response map[string]string
	err = json.Unmarshal(suite.recorder.Body.Bytes(), &response)
	suite.NoError(err)
	suite.Equal("ready", response["status"])
	suite.NotEmpty(response["timestamp"])
}

func (suite *HealthHandlerTestSuite) TestHealthEndpointWithoutDB() {
	req, err := http.NewRequest("GET", "/health", nil)
	suite.NoError(err)

	suite.recorder = httptest.NewRecorder()
	suite.router.ServeHTTP(suite.recorder, req)

	suite.Equal(http.StatusOK, suite.recorder.Code)
	
	var response HealthResponse
	err = json.Unmarshal(suite.recorder.Body.Bytes(), &response)
	suite.NoError(err)
	suite.Equal("healthy", response.Status)
	suite.Equal("not configured", response.Checks["database"])
	suite.Equal("healthy", response.Checks["redis"])
	suite.NotEmpty(response.Timestamp)
}

func (suite *HealthHandlerTestSuite) TestHealthEndpointWithFailingRedis() {
	// Setup router with failing Redis
	logger := zap.NewNop().Sugar()
	handler := NewHandler(logger)
	redis := &mockRedis{shouldFail: true}
	healthHandler := NewHealthHandler(handler, nil, redis)
	
	router := gin.New()
	router.GET("/health", healthHandler.Health)

	req, err := http.NewRequest("GET", "/health", nil)
	suite.NoError(err)

	recorder := httptest.NewRecorder()
	router.ServeHTTP(recorder, req)

	suite.Equal(http.StatusServiceUnavailable, recorder.Code)
	
	var response HealthResponse
	err = json.Unmarshal(recorder.Body.Bytes(), &response)
	suite.NoError(err)
	suite.Equal("unhealthy", response.Status)
	suite.Contains(response.Checks["redis"], "unreachable")
}

func TestHealthHandlerSuite(t *testing.T) {
	suite.Run(t, new(HealthHandlerTestSuite))
}