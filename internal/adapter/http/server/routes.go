package server

import (
	"encoding/json"
	"net/http"
)

// setupRoutes - setups http routes
func (a *API) setupRoutes(mux *http.ServeMux) {
	// System Health
	mux.HandleFunc("/health", a.HealthCheck)
	mux.HandleFunc("POST /order", a.routes.order.CreateOrder)

}

// HealthCheck - returns system information.
func (a *API) HealthCheck(w http.ResponseWriter, r *http.Request) {
	response := map[string]any{
		"status": "available",
		"system_info": map[string]string{
			"address": a.addr,
		},
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(response); err != nil {
		a.log.Error(r.Context(), "healthcheck", "failed to encode", err)
		return
	}
}
