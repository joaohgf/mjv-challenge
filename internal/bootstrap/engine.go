package bootstrap

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/joaohgf/mjv-challenge/config"
	repositorymodel "github.com/joaohgf/mjv-challenge/internal/outbound/repository/model"
	mongoadapter "github.com/joaohgf/mjv-challenge/pkg/mongo"
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
	engine.GET("/health", health)
	engine.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))
	if mongo != nil {
		buildOrder(engine, mongo, settings)
	}
	return engine
}

// health returns no content to indicate that the HTTP process is reachable.
func health(context *gin.Context) {
	context.Status(http.StatusNoContent)
}
