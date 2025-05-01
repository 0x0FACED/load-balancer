package client

import (
	"encoding/json"
	"errors"
	"net/http"

	"github.com/0x0FACED/load-balancer/internal/pkg/httpcommon"
)

type ClientHandler struct {
	repository Repository
}

func NewClientHandler(repository Repository) *ClientHandler {
	return &ClientHandler{
		repository: repository,
	}
}

func (h *ClientHandler) RegisterRoutes(mux *http.ServeMux) {
	mux.HandleFunc("POST /client", h.Create)
	mux.HandleFunc("GET /client/{id}", h.Get)
	mux.HandleFunc("PUT /client", h.Update)
	mux.HandleFunc("DELETE /client/{id}", h.Delete)
}

func (h *ClientHandler) Create(w http.ResponseWriter, r *http.Request) {
	var client Client
	if err := json.NewDecoder(r.Body).Decode(&client); err != nil {
		httpcommon.JSONError(w, http.StatusBadRequest, errors.New("invalid request body"))
		return
	}

	if client.ID == "" {
		client.ID = httpcommon.ClientIDFromRequest(r)
	}
	if err := h.repository.Create(r.Context(), client); err != nil {
		httpcommon.JSONError(w, http.StatusInternalServerError, err)
		return
	}

	httpcommon.EmptyResponse(w, http.StatusCreated)
}

func (h *ClientHandler) Get(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	if id == "" {
		httpcommon.JSONError(w, http.StatusBadRequest, errors.New("invalid request"))
		return
	}

	client, err := h.repository.Get(r.Context(), id)
	if err != nil {
		httpcommon.JSONError(w, http.StatusInternalServerError, err)
		return
	}

	if client == nil {
		httpcommon.EmptyResponse(w, http.StatusNotFound)
		return
	}

	httpcommon.JSONResponse(w, http.StatusOK, client)
}

func (h *ClientHandler) Update(w http.ResponseWriter, r *http.Request) {
	var client Client
	if err := json.NewDecoder(r.Body).Decode(&client); err != nil {
		httpcommon.JSONError(w, http.StatusBadRequest, errors.New("invalid request body"))
		return
	}

	if err := h.repository.Update(r.Context(), client); err != nil {
		httpcommon.JSONError(w, http.StatusInternalServerError, err)
		return
	}

	httpcommon.EmptyResponse(w, http.StatusOK)
}

func (h *ClientHandler) Delete(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	if id == "" {
		httpcommon.JSONError(w, http.StatusBadRequest, errors.New("invalid request"))
		return
	}

	if err := h.repository.Delete(r.Context(), id); err != nil {
		httpcommon.JSONError(w, http.StatusInternalServerError, err)
		return
	}

	httpcommon.EmptyResponse(w, http.StatusOK)
}
