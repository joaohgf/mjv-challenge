package telemetry

import (
	"context"
	"fmt"
	"time"

	"github.com/gin-gonic/gin"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/propagation"
	"go.opentelemetry.io/otel/trace"
)

// GinMiddleware creates a server span and HTTP metrics for every request.
func GinMiddleware() gin.HandlerFunc {
	return func(context *gin.Context) {
		parent := propagation.HeaderCarrier(context.Request.Header)
		request := context.Request.WithContext(traceContext(context.Request.Context(), parent))
		started := time.Now()
		spanContext, span := StartSpan(request.Context(), request.Method+" "+request.URL.Path, trace.SpanKindServer)
		context.Request = request.WithContext(spanContext)
		context.Next()
		route := context.FullPath()
		if route == "" {
			route = request.URL.Path
		}
		span.SetName(request.Method + " " + route)
		if context.Writer.Status() >= 500 {
			End(span, fmt.Errorf("HTTP response status %d", context.Writer.Status()))
		} else {
			End(span, nil)
		}
		RecordHTTP(spanContext, request.Method, route, context.Writer.Status(), time.Since(started))
	}
}

// traceContext restores a remote request parent from HTTP headers.
func traceContext(ctx context.Context, carrier propagation.TextMapCarrier) context.Context {
	return otel.GetTextMapPropagator().Extract(ctx, carrier)
}
