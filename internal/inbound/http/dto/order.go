package dto

import "time"

type (
	// OrderCreate is the JSON body accepted when an order is submitted.
	OrderCreate struct {
		ProductName string `json:"product_name" binding:"required"`
		Quantity    int    `json:"quantity" binding:"gte=1"`
	}
	// OrderResponse is the JSON representation returned by the order endpoints.
	OrderResponse struct {
		ID          string     `json:"id"`
		ProductName string     `json:"product_name"`
		Status      string     `json:"status"`
		Quantity    int        `json:"quantity"`
		CreatedAt   time.Time  `json:"created_at"`
		UpdatedAt   *time.Time `json:"updated_at"`
	}
)

// NewOrderResponse creates an empty HTTP response DTO for a mapper to fill.
func NewOrderResponse() *OrderResponse {
	target := new(OrderResponse)
	return target
}
