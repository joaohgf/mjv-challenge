package adapter

import (
	"log/slog"
	"net/http"

	"github.com/gin-gonic/gin"
	errs "github.com/joaohgf/mjv-challenge/internal/core/error"
	"github.com/joaohgf/mjv-challenge/internal/inbound/http/dto"
)

// Create creates an order and queues it for asynchronous processing.
// @Summary Create an order
// @Tags orders
// @Accept json
// @Produce json
// @Param order body dto.OrderCreate true "Order data"
// @Success 201 {object} dto.OrderResponse
// @Failure 400 {object} errs.RequestError
// @Failure 500 {object} errs.RequestError
// @Router /orders [post]
func (handler *OrderHandler) Create(context *gin.Context) {
	var body dto.OrderCreate
	if err := context.ShouldBindJSON(&body); err != nil {
		slog.Warn("invalid create order request", "error", err)
		context.JSON(http.StatusBadRequest, bindingError[dto.OrderCreate](err))
		return
	}
	created, err := handler.creator.Create(context, handler.toMapper.To(&body))
	if err != nil {
		slog.Error("creating order", "error", err)
		context.JSON(http.StatusInternalServerError, errs.NewRequestError("failed to create order"))
		return
	}
	slog.Debug("order created", "order_id", created.ID)
	context.JSON(http.StatusCreated, handler.fromMapper.From(created))
}
