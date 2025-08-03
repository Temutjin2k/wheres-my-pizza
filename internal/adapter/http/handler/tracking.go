package handler

import (
	"context"
	"encoding/json"
	"net/http"

	"github.com/Temutjin2k/wheres-my-pizza/internal/domain/models"
	"github.com/Temutjin2k/wheres-my-pizza/internal/domain/types"
	"github.com/Temutjin2k/wheres-my-pizza/pkg/logger"
)

type TrackingService interface {
	GetOrderStatus(ctx context.Context, orderNumber string) (models.OrderStatus, error)
	GetTrackingHistory(ctx context.Context, orderNumber string) ([]models.OrderHistory, error)
	ListWorkers(ctx context.Context) ([]models.Worker, error)
}

type Tracking struct {
	service TrackingService
	log     logger.Logger
}

func NewTracking(service TrackingService, log logger.Logger) *Tracking {
	return &Tracking{
		service: service,
		log:     log,
	}
}

func (h *Tracking) GetOrderStatus(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	orderNumber := r.PathValue("order_number")

	h.log.Debug(ctx, types.ActionRequestReceived, "Request has been received", "URL", r.URL.String(), "method", r.Method, "order_number", orderNumber, "host", r.Host)

	orderStatus, err := h.service.GetOrderStatus(ctx, orderNumber)
	if err != nil {
		errorResponse(w, getCode(err), err.Error())
		return
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(orderStatus); err != nil {
		internalErrorResponse(w, err)
	}
}

func (h *Tracking) GetTrackingHistory(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	orderNumber := r.PathValue("order_number")

	h.log.Debug(ctx, types.ActionRequestReceived, "Request has been received", "URL", r.URL.String(), "method", r.Method, "order_number", orderNumber, "host", r.Host)

	historyList, err := h.service.GetTrackingHistory(ctx, orderNumber)
	if err != nil {
		errorResponse(w, getCode(err), err.Error())
		return
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(historyList); err != nil {
		internalErrorResponse(w, err)
	}
}

func (h *Tracking) ListWorkers(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	h.log.Debug(ctx, types.ActionRequestReceived, "Request has been received", "URL", r.URL.String(), "method", r.Method, "host", r.Host)

	workersList, err := h.service.ListWorkers(ctx)
	if err != nil {
		errorResponse(w, getCode(err), err.Error())
		return
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(workersList); err != nil {
		internalErrorResponse(w, err)
	}
}
