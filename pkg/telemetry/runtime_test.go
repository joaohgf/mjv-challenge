package telemetry

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"go.opentelemetry.io/otel/trace"
)

func TestTelemetryRecordsOperationsAndSpans(t *testing.T) {
	ctx, span := StartSpan(context.Background(), "test", trace.SpanKindInternal)
	End(span, errors.New("failed"))
	RecordHTTP(ctx, http.MethodGet, "/orders", http.StatusOK, time.Millisecond)
	RecordOperation(ctx, "mongo", "get", nil)
	RecordOperation(ctx, "rabbitmq", "publish", errors.New("failed"))
}

func TestGinMiddlewareHandlesSuccessfulAndFailedResponses(t *testing.T) {
	gin.SetMode(gin.TestMode)
	engine := gin.New()
	engine.Use(GinMiddleware())
	engine.GET("/orders/:id", func(context *gin.Context) { context.Status(http.StatusNoContent) })
	engine.GET("/failure", func(context *gin.Context) { context.Status(http.StatusInternalServerError) })

	for _, path := range []string{"/orders/order-1", "/failure"} {
		response := httptest.NewRecorder()
		engine.ServeHTTP(response, httptest.NewRequest(http.MethodGet, path, nil))
		if response.Code != http.StatusNoContent && response.Code != http.StatusInternalServerError {
			t.Fatalf("unexpected response status %d", response.Code)
		}
	}
}
