package adapter_test

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/joaohgf/mjv-challenge/internal/core/domain"
	errs "github.com/joaohgf/mjv-challenge/internal/core/error"
	"github.com/joaohgf/mjv-challenge/internal/inbound/http/adapter"
	"github.com/joaohgf/mjv-challenge/internal/inbound/http/mapper"
)

type creatorStub struct {
	err   error
	calls *int
}

type getterStub struct {
	err error
}

func (stub creatorStub) Create(_ context.Context, order *domain.Order) (*domain.Order, error) {
	if stub.calls != nil {
		*stub.calls++
	}
	if stub.err != nil {
		return nil, stub.err
	}
	order.ID = "order-1"
	return order, nil
}

func (stub getterStub) Get(_ context.Context, id string) (*domain.Order, error) {
	return &domain.Order{ID: id}, stub.err
}

func TestOrderRouteCreatesOrder(t *testing.T) {
	gin.SetMode(gin.TestMode)
	orderMapper := new(mapper.Order)
	handler := adapter.NewOrderHandler(orderMapper, orderMapper, creatorStub{}, getterStub{})
	engine := gin.New()
	engine.POST("/orders", handler.Create)
	response := httptest.NewRecorder()
	request := httptest.NewRequest(http.MethodPost, "/orders", strings.NewReader(`{"product_name":"book","quantity":2}`))
	request.Header.Set("Content-Type", "application/json")

	engine.ServeHTTP(response, request)

	if response.Code != http.StatusCreated {
		t.Fatalf("expected status %d, got %d: %s", http.StatusCreated, response.Code, response.Body)
	}
	if !strings.Contains(response.Body.String(), `"order_id":"order-1"`) {
		t.Fatalf("expected created order, got %s", response.Body)
	}
}

func TestOrderRouteGetsOrder(t *testing.T) {
	gin.SetMode(gin.TestMode)
	orderMapper := new(mapper.Order)
	handler := adapter.NewOrderHandler(orderMapper, orderMapper, creatorStub{}, getterStub{})
	engine := gin.New()
	engine.GET("/orders/:id", handler.Get)
	response := httptest.NewRecorder()

	engine.ServeHTTP(response, httptest.NewRequest(http.MethodGet, "/orders/order-1", nil))

	if response.Code != http.StatusOK || !strings.Contains(response.Body.String(), `"order_id":"order-1"`) {
		t.Fatalf("expected found order, got status=%d body=%s", response.Code, response.Body.String())
	}
}

func TestOrderRouteReturnsNotFound(t *testing.T) {
	gin.SetMode(gin.TestMode)
	orderMapper := new(mapper.Order)
	handler := adapter.NewOrderHandler(orderMapper, orderMapper, creatorStub{}, getterStub{err: errs.ErrNotFound})
	engine := gin.New()
	engine.GET("/orders/:id", handler.Get)
	response := httptest.NewRecorder()

	engine.ServeHTTP(response, httptest.NewRequest(http.MethodGet, "/orders/order-1", nil))

	if response.Code != http.StatusNotFound {
		t.Fatalf("expected status %d, got %d", http.StatusNotFound, response.Code)
	}
}

func TestOrderRouteRejectsInvalidCreateBody(t *testing.T) {
	gin.SetMode(gin.TestMode)
	orderMapper := new(mapper.Order)
	handler := adapter.NewOrderHandler(orderMapper, orderMapper, creatorStub{}, getterStub{})
	engine := gin.New()
	engine.POST("/orders", handler.Create)
	response := httptest.NewRecorder()

	engine.ServeHTTP(response, httptest.NewRequest(http.MethodPost, "/orders", strings.NewReader(`invalid`)))

	var result errs.RequestError
	if err := json.Unmarshal(response.Body.Bytes(), &result); err != nil {
		t.Fatal(err)
	}
	if response.Code != http.StatusBadRequest || result.Err != "invalid request body" || result.Errors != nil {
		t.Fatalf("expected malformed body response, got status=%d body=%#v", response.Code, result)
	}
}

func TestOrderRoutePresentsInvalidOrderFields(t *testing.T) {
	for _, test := range []struct {
		body    string
		field   string
		message string
	}{
		{`{"quantity":2}`, "product_name", "field is required"},
		{`{"product_name":"book","quantity":0}`, "quantity", "must be greater than zero"},
		{`{"product_name":"book","quantity":-1}`, "quantity", "must be greater than zero"},
	} {
		t.Run(test.field, func(t *testing.T) {
			calls := 0
			gin.SetMode(gin.TestMode)
			orderMapper := new(mapper.Order)
			handler := adapter.NewOrderHandler(orderMapper, orderMapper, creatorStub{calls: &calls}, getterStub{})
			engine := gin.New()
			engine.POST("/orders", handler.Create)
			response := httptest.NewRecorder()
			request := httptest.NewRequest(http.MethodPost, "/orders", strings.NewReader(test.body))
			request.Header.Set("Content-Type", "application/json")

			engine.ServeHTTP(response, request)

			var result errs.RequestError
			if err := json.Unmarshal(response.Body.Bytes(), &result); err != nil {
				t.Fatal(err)
			}
			if response.Code != http.StatusBadRequest || calls != 0 || result.Err != "invalid request data" || len(result.Errors) != 1 || result.Errors[0].Field != test.field || result.Errors[0].Message != test.message {
				t.Fatalf("expected presented validation error, got status=%d calls=%d body=%#v", response.Code, calls, result)
			}
		})
	}
}

func TestOrderRouteReturnsInternalErrorForCreatorFailure(t *testing.T) {
	gin.SetMode(gin.TestMode)
	orderMapper := new(mapper.Order)
	handler := adapter.NewOrderHandler(orderMapper, orderMapper, creatorStub{err: errors.New("broker unavailable")}, getterStub{})
	engine := gin.New()
	engine.POST("/orders", handler.Create)
	response := httptest.NewRecorder()
	request := httptest.NewRequest(http.MethodPost, "/orders", strings.NewReader(`{"product_name":"book","quantity":2}`))
	request.Header.Set("Content-Type", "application/json")

	engine.ServeHTTP(response, request)

	if response.Code != http.StatusInternalServerError {
		t.Fatalf("expected status %d, got %d", http.StatusInternalServerError, response.Code)
	}
}

func TestOrderRouteReturnsInternalErrorForGetFailure(t *testing.T) {
	gin.SetMode(gin.TestMode)
	orderMapper := new(mapper.Order)
	handler := adapter.NewOrderHandler(orderMapper, orderMapper, creatorStub{}, getterStub{err: errors.New("database unavailable")})
	engine := gin.New()
	engine.GET("/orders/:id", handler.Get)
	response := httptest.NewRecorder()

	engine.ServeHTTP(response, httptest.NewRequest(http.MethodGet, "/orders/order-1", nil))

	if response.Code != http.StatusInternalServerError {
		t.Fatalf("expected status %d, got %d", http.StatusInternalServerError, response.Code)
	}
}
