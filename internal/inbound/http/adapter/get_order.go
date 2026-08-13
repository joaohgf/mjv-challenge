package adapter

import (
	"errors"
	"log/slog"
	"net/http"

	errs "github.com/joaohgf/mjv-challenge/internal/core/error"

	"github.com/gin-gonic/gin"
)

// Get returns an order by its identifier.
// @Summary Get an order
// @Tags orders
// @Produce json
// @Param id path string true "Order ID"
// @Success 200 {object} dto.OrderResponse
// @Failure 404 {object} errs.RequestError
// @Failure 500 {object} errs.RequestError
// @Router /orders/{id} [get]
func (handler *OrderHandler) Get(context *gin.Context) {
	id := context.Param("id")
	order, err := handler.getter.Get(context, id)
	if errors.Is(err, errs.ErrNotFound) {
		slog.Info("order not found", "order_id", id)
		context.JSON(http.StatusNotFound, errs.NewRequestError("order not found"))
		return
	}
	if err != nil {
		slog.Error("getting order", "order_id", id, "error", err)
		context.JSON(http.StatusInternalServerError, errs.NewRequestError("failed to get order"))
		return
	}
	slog.Info("order retrieved", "order_id", id)
	context.JSON(http.StatusOK, handler.fromMapper.From(order))
}
