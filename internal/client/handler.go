package client

import (
	"encoding/json"
	"github.com/rs/zerolog/log"
	"net/http"
	"strings"
)

type Handler struct {
	Store ClientStore
}

func NewHandler(store ClientStore) *Handler {
	return &Handler{Store: store}
}

func (h *Handler) RegisterRoutes(mux *http.ServeMux) {
	mux.HandleFunc("/clients", h.handleClients)
	mux.HandleFunc("/clients/", h.handleClientByID)
}

func (h *Handler) handleClients(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodPost:
		h.createClient(w, r)
	case http.MethodGet:
		h.listClients(w, r)
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
		h.getClient(w, r, id)
	case http.MethodPut:
		h.updateClient(w, r, id)
	case http.MethodDelete:
		h.deleteClient(w, r, id)
	default:
		http.Error(w, "Method Not Allowed", http.StatusMethodNotAllowed)
	}
}

func (h *Handler) createClient(w http.ResponseWriter, r *http.Request) {
	var c Client
	if err := json.NewDecoder(r.Body).Decode(&c); err != nil {
		http.Error(w, "Invalid JSON", http.StatusBadRequest)
		return
	}
	if c.ID == "" {
		http.Error(w, "client_id is required", http.StatusBadRequest)
		return
	}

	if c.APIKey == "" {
		c.APIKey = GenerateAPIKey()
	}

	if err := h.Store.Create(&c); err != nil {
		http.Error(w, err.Error(), http.StatusConflict)
		return
	}
	log.Info().Str("client_id", c.ID).Msg("Client created")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(c)
}

func (h *Handler) listClients(w http.ResponseWriter, r *http.Request) {
	clients, err := h.Store.List()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	json.NewEncoder(w).Encode(clients)
}

func (h *Handler) getClient(w http.ResponseWriter, r *http.Request, id string) {
	c, err := h.Store.Get(id)
	if err != nil {
		http.Error(w, "client not found", http.StatusNotFound)
		return
	}
	json.NewEncoder(w).Encode(c)
}

func (h *Handler) updateClient(w http.ResponseWriter, r *http.Request, id string) {
	var c Client
	if err := json.NewDecoder(r.Body).Decode(&c); err != nil {
		http.Error(w, "Invalid JSON", http.StatusBadRequest)
		return
	}

	if c.ID == "" {
		c.ID = id // fallback
	}
	if c.ID != id {
		http.Error(w, "client_id in path and body must match", http.StatusBadRequest)
		return
	}
	if err := h.Store.Update(&c); err != nil {
		http.Error(w, err.Error(), http.StatusNotFound)
		return
	}
	json.NewEncoder(w).Encode(c)
}

func (h *Handler) deleteClient(w http.ResponseWriter, r *http.Request, id string) {
	if err := h.Store.Delete(id); err != nil {
		http.Error(w, err.Error(), http.StatusNotFound)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}
