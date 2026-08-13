package bootstrap

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/joaohgf/mjv-challenge/config"
)

func TestNewEngineRegistersUnavailableHealthRouteWithoutMongo(t *testing.T) {
	gin.SetMode(gin.TestMode)
	engine := NewEngine(nil, config.Load())
	response := httptest.NewRecorder()

	engine.ServeHTTP(response, httptest.NewRequest(http.MethodGet, "/health", nil))

	if response.Code != http.StatusServiceUnavailable {
		t.Fatalf("expected status %d, got %d", http.StatusServiceUnavailable, response.Code)
	}
}

func TestHealthReturnsNoContentWhenMongoIsAvailable(t *testing.T) {
	response := healthResponse(func(context.Context) error { return nil })

	if response.Code != http.StatusNoContent {
		t.Fatalf("expected status %d, got %d", http.StatusNoContent, response.Code)
	}
}

func TestHealthReturnsUnavailableWhenMongoFails(t *testing.T) {
	response := healthResponse(func(context.Context) error { return errors.New("unavailable") })

	if response.Code != http.StatusServiceUnavailable {
		t.Fatalf("expected status %d, got %d", http.StatusServiceUnavailable, response.Code)
	}
}

func healthResponse(check func(context.Context) error) *httptest.ResponseRecorder {
	engine := gin.New()
	engine.GET("/health", health(check))
	response := httptest.NewRecorder()
	engine.ServeHTTP(response, httptest.NewRequest(http.MethodGet, "/health", nil))
	return response
}
