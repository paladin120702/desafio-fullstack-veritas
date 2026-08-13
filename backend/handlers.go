package main

import (
	"encoding/json"
	"errors"
	"io"
	"log"
	"net/http"
	"strings"
)

const maxBodyBytes = 1 << 20 // 1 MiB

// API agrupa os handlers HTTP em torno do store.
type API struct {
	store *Store
}

// NewAPI cria os handlers.
func NewAPI(store *Store) *API {
	return &API{store: store}
}

// Routes monta o mux com todas as rotas da aplicacao.
func (a *API) Routes() http.Handler {
	mux := http.NewServeMux()

	mux.HandleFunc("GET /api/health", a.handleHealth)
	mux.HandleFunc("GET /api/tasks", a.handleList)
	mux.HandleFunc("POST /api/tasks", a.handleCreate)
	mux.HandleFunc("PUT /api/tasks/{id}", a.handleUpdate)
	mux.HandleFunc("DELETE /api/tasks/{id}", a.handleDelete)

	return mux
}

// GET /api/health
func (a *API) handleHealth(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, http.StatusOK, map[string]string{"status": "ok"})
}

// GET /api/tasks
func (a *API) handleList(w http.ResponseWriter, r *http.Request) {
	tasks := a.store.List()
	writeJSON(w, http.StatusOK, map[string]any{"tasks": tasks})
}

// POST /api/tasks
func (a *API) handleCreate(w http.ResponseWriter, r *http.Request) {
	var in CreateTaskInput
	if !decodeJSON(w, r, &in) {
		return
	}

	if fieldErrs := in.Validate(); fieldErrs != nil {
		writeValidationError(w, fieldErrs)
		return
	}

	task, err := a.store.Create(in)
	if err != nil {
		writeServerError(w, "nao foi possivel salvar a tarefa", err)
		return
	}
	writeJSON(w, http.StatusCreated, task)
}

// PUT /api/tasks/{id}
func (a *API) handleUpdate(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")

	var in UpdateTaskInput
	if !decodeJSON(w, r, &in) {
		return
	}

	if fieldErrs := in.Validate(); fieldErrs != nil {
		writeValidationError(w, fieldErrs)
		return
	}

	task, err := a.store.Update(id, in)
	switch {
	case errors.Is(err, ErrNotFound):
		writeError(w, http.StatusNotFound, "not_found", "tarefa nao encontrada", nil)
	case err != nil:
		writeServerError(w, "nao foi possivel atualizar a tarefa", err)
	default:
		writeJSON(w, http.StatusOK, task)
	}
}

// DELETE /api/tasks/{id}
func (a *API) handleDelete(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")

	err := a.store.Delete(id)
	switch {
	case errors.Is(err, ErrNotFound):
		writeError(w, http.StatusNotFound, "not_found", "tarefa nao encontrada", nil)
	case err != nil:
		writeServerError(w, "nao foi possivel excluir a tarefa", err)
	default:
		w.WriteHeader(http.StatusNoContent)
	}
}

// --- helpers de resposta ---

type errorBody struct {
	Code    string      `json:"code"`
	Message string      `json:"message"`
	Fields  FieldErrors `json:"fields,omitempty"`
}

func writeJSON(w http.ResponseWriter, status int, payload any) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(status)
	if err := json.NewEncoder(w).Encode(payload); err != nil {
		log.Printf("erro ao escrever resposta: %v", err)
	}
}

func writeError(w http.ResponseWriter, status int, code, message string, fields FieldErrors) {
	writeJSON(w, status, map[string]errorBody{
		"error": {Code: code, Message: message, Fields: fields},
	})
}

func writeValidationError(w http.ResponseWriter, fields FieldErrors) {
	writeError(w, http.StatusBadRequest, "validation_error", "dados invalidos", fields)
}

func writeServerError(w http.ResponseWriter, message string, err error) {
	log.Printf("erro interno: %v", err)
	writeError(w, http.StatusInternalServerError, "internal_error", message, nil)
}

// decodeJSON le o corpo da requisicao e responde 400 quando ele nao e valido.
// Devolve false quando a resposta de erro ja foi escrita.
func decodeJSON(w http.ResponseWriter, r *http.Request, dst any) bool {
	if ct := r.Header.Get("Content-Type"); ct != "" && !strings.HasPrefix(ct, "application/json") {
		writeError(w, http.StatusUnsupportedMediaType, "unsupported_media_type",
			"Content-Type deve ser application/json", nil)
		return false
	}

	r.Body = http.MaxBytesReader(w, r.Body, maxBodyBytes)
	dec := json.NewDecoder(r.Body)
	dec.DisallowUnknownFields()

	if err := dec.Decode(dst); err != nil {
		writeError(w, http.StatusBadRequest, "invalid_json", "corpo da requisicao invalido: "+err.Error(), nil)
		return false
	}
	if err := dec.Decode(&struct{}{}); err != io.EOF {
		writeError(w, http.StatusBadRequest, "invalid_json", "corpo deve conter um unico objeto JSON", nil)
		return false
	}
	return true
}
