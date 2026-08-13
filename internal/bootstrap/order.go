package bootstrap

import (
	"github.com/gin-gonic/gin"

	"github.com/joaohgf/mjv-challenge/config"
	"github.com/joaohgf/mjv-challenge/internal/core/usecase"
	httpadapter "github.com/joaohgf/mjv-challenge/internal/inbound/http/adapter"
	httpmapper "github.com/joaohgf/mjv-challenge/internal/inbound/http/mapper"
	outboxadapter "github.com/joaohgf/mjv-challenge/internal/outbound/outbox/adapter"
	repositoryadapter "github.com/joaohgf/mjv-challenge/internal/outbound/repository/adapter"
	repositorymapper "github.com/joaohgf/mjv-challenge/internal/outbound/repository/mapper"
	repositorymodel "github.com/joaohgf/mjv-challenge/internal/outbound/repository/model"
	mongoadapter "github.com/joaohgf/mjv-challenge/pkg/mongo"
)

// buildOrder keeps order-specific adapters and use cases out of cmd/api.
func buildOrder(
	router gin.IRoutes,
	mongo *mongoadapter.Repository[*repositorymodel.Order],
	settings config.Config,
) {
	repository := repositoryadapter.NewRepository(mongo, &repositorymapper.Order{})
	outbox := outboxadapter.NewStore(mongo.Collection(settings.Database.OutboxCollectionName), &repositorymapper.Order{}, settings.Queue.OutboxLease, mongo.WithOperationTimeout)
	creator := usecase.NewCreateOrder(outboxadapter.NewCreator(mongo, repository, outbox))
	getter := usecase.NewGetOrder(repository)
	mapper := &httpmapper.Order{}
	orderHandler := httpadapter.NewOrderHandler(mapper, mapper, creator, getter)
	router.POST("/orders", orderHandler.Create)
	router.GET("/orders/:id", orderHandler.Get)
}
