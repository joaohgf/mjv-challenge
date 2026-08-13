package bootstrap

import (
	"context"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/joaohgf/mjv-challenge/config"
	repositorymodel "github.com/joaohgf/mjv-challenge/internal/outbound/repository/model"
	mongoadapter "github.com/joaohgf/mjv-challenge/pkg/mongo"
	"github.com/joaohgf/mjv-challenge/pkg/telemetry"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
)

// NewEngine registers cross-cutting routes and composes the order HTTP flow.
func NewEngine(
	mongo *mongoadapter.Repository[*repositorymodel.Order],
	settings config.Config,
) *gin.Engine {
	engine := gin.New()
	engine.Use(gin.Recovery())
	engine.Use(telemetry.GinMiddleware())
	engine.GET("/health", health(mongoHealth(mongo)))
	engine.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))
	if mongo != nil {
		buildOrder(engine, mongo, settings)
	}
	return engine
}

// mongoHealth supplies the database probe used by the API readiness endpoint.
func mongoHealth(mongo *mongoadapter.Repository[*repositorymodel.Order]) func(context.Context) error {
	if mongo == nil {
		return nil
	}
	return mongo.Ping
}

// health reports readiness only while the database accepts requests.
func health(check func(context.Context) error) gin.HandlerFunc {
	return func(request *gin.Context) {
		if check == nil || check(request.Request.Context()) != nil {
			request.Status(http.StatusServiceUnavailable)
			return
		}
		request.Status(http.StatusNoContent)
	}
}
