package bootstrap

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/joaohgf/mjv-challenge/config"
)

func TestNewEngineRegistersHealthRoute(t *testing.T) {
	gin.SetMode(gin.TestMode)
	engine := NewEngine(nil, config.Load())
	response := httptest.NewRecorder()

	engine.ServeHTTP(response, httptest.NewRequest(http.MethodGet, "/health", nil))

	if response.Code != http.StatusNoContent {
		t.Fatalf("expected status %d, got %d", http.StatusNoContent, response.Code)
	}
}
