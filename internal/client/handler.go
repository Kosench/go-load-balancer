package client

import (
	"net/http"
	"strings"
)

type Handler struct {
	Store ClientStore
}

func NewHandler(store ClientStore) *Handler {
	return &Handler{Store: store}
}

func (h *Handler) RegisterRouter(mux *http.ServeMux) {
	mux.HandleFunc("/client", h.handleClients)
	mux.HandleFunc("/client/", h.handleClientByID)
}

func (h *Handler) handleClients(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodPost:
		h.createClient(w, r)
	case http.MethodGet:
		h.listClient(w, r)
	default:
		http.Error(w, "Method Not Allowed", http.StatusMethodNotAllowed)
	}
}

func (h *Handler) handleClientByID(w http.ResponseWriter, r *http.Request) {
	id := strings.TrimPrefix(r.URL.Path, "/clients/")
	if id == "" {
		http.Error(w, "Client ID required", http.StatusBadRequest)
		return
	}

	switch r.Method {
	case http.MethodGet:
		h.getClient(w, r)
	case http.MethodPut:
		h.updateClient(w, r)
	case http.MethodDelete:
		h.deleteClient(w, r)
	default:
		http.Error(w, "Method Not Allowed", http.StatusMethodNotAllowed)
	}
}
