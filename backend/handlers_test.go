package main

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"
)

func newTestServer(t *testing.T, dataFile string) http.Handler {
	t.Helper()
	store, err := NewStore(dataFile)
	if err != nil {
		t.Fatalf("NewStore: %v", err)
	}
	return withCORS(nil, NewAPI(store).Routes())
}

func do(t *testing.T, h http.Handler, method, path string, body any) *httptest.ResponseRecorder {
	t.Helper()

	var reader *bytes.Reader
	if body != nil {
		raw, err := json.Marshal(body)
		if err != nil {
			t.Fatalf("marshal body: %v", err)
		}
		reader = bytes.NewReader(raw)
	} else {
		reader = bytes.NewReader(nil)
	}

	req := httptest.NewRequest(method, path, reader)
	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)
	return rec
}

func decodeTask(t *testing.T, rec *httptest.ResponseRecorder) Task {
	t.Helper()
	var task Task
	if err := json.Unmarshal(rec.Body.Bytes(), &task); err != nil {
		t.Fatalf("decode task: %v (body=%s)", err, rec.Body.String())
	}
	return task
}

func listTasks(t *testing.T, h http.Handler) []Task {
	t.Helper()
	rec := do(t, h, http.MethodGet, "/api/tasks", nil)
	if rec.Code != http.StatusOK {
		t.Fatalf("GET /api/tasks = %d, quer 200", rec.Code)
	}
	var payload struct {
		Tasks []Task `json:"tasks"`
	}
	if err := json.Unmarshal(rec.Body.Bytes(), &payload); err != nil {
		t.Fatalf("decode list: %v", err)
	}
	return payload.Tasks
}

func TestCreateTask(t *testing.T) {
	h := newTestServer(t, "")

	rec := do(t, h, http.MethodPost, "/api/tasks", map[string]any{
		"title":       "  Escrever README  ",
		"description": "com instrucoes de execucao",
	})
	if rec.Code != http.StatusCreated {
		t.Fatalf("POST = %d, quer 201 (body=%s)", rec.Code, rec.Body.String())
	}

	task := decodeTask(t, rec)
	if task.ID == "" {
		t.Error("tarefa criada sem ID")
	}
	if task.Title != "Escrever README" {
		t.Errorf("titulo = %q, quer %q (espacos devem ser removidos)", task.Title, "Escrever README")
	}
	if task.Status != StatusTodo {
		t.Errorf("status = %q, quer %q (padrao)", task.Status, StatusTodo)
	}
	if task.CreatedAt.IsZero() {
		t.Error("createdAt nao preenchido")
	}
}

func TestCreateTaskValidation(t *testing.T) {
	h := newTestServer(t, "")

	cases := []struct {
		name  string
		body  map[string]any
		field string
	}{
		{"titulo vazio", map[string]any{"title": "   "}, "title"},
		{"titulo ausente", map[string]any{"description": "sem titulo"}, "title"},
		{"status invalido", map[string]any{"title": "ok", "status": "arquivada"}, "status"},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			rec := do(t, h, http.MethodPost, "/api/tasks", tc.body)
			if rec.Code != http.StatusBadRequest {
				t.Fatalf("POST = %d, quer 400 (body=%s)", rec.Code, rec.Body.String())
			}

			var payload struct {
				Error struct {
					Code   string            `json:"code"`
					Fields map[string]string `json:"fields"`
				} `json:"error"`
			}
			if err := json.Unmarshal(rec.Body.Bytes(), &payload); err != nil {
				t.Fatalf("decode erro: %v", err)
			}
			if payload.Error.Code != "validation_error" {
				t.Errorf("code = %q, quer validation_error", payload.Error.Code)
			}
			if _, ok := payload.Error.Fields[tc.field]; !ok {
				t.Errorf("faltou o campo %q em fields (%v)", tc.field, payload.Error.Fields)
			}
		})
	}

	if got := len(listTasks(t, h)); got != 0 {
		t.Errorf("tarefas armazenadas = %d, quer 0 apos requisicoes invalidas", got)
	}
}

func TestUpdateTaskMovesBetweenColumns(t *testing.T) {
	h := newTestServer(t, "")

	created := decodeTask(t, do(t, h, http.MethodPost, "/api/tasks", map[string]any{"title": "Mover"}))

	rec := do(t, h, http.MethodPut, "/api/tasks/"+created.ID, map[string]any{
		"status":      "doing",
		"description": "em andamento",
	})
	if rec.Code != http.StatusOK {
		t.Fatalf("PUT = %d, quer 200 (body=%s)", rec.Code, rec.Body.String())
	}

	updated := decodeTask(t, rec)
	if updated.Status != StatusDoing {
		t.Errorf("status = %q, quer doing", updated.Status)
	}
	if updated.Title != "Mover" {
		t.Errorf("titulo = %q, quer Mover (campo ausente deve ser preservado)", updated.Title)
	}
	if updated.Description != "em andamento" {
		t.Errorf("descricao = %q, quer 'em andamento'", updated.Description)
	}
	if !updated.UpdatedAt.After(updated.CreatedAt) && updated.UpdatedAt.Before(updated.CreatedAt) {
		t.Error("updatedAt deveria ser >= createdAt")
	}
}

func TestUpdateTaskReordersInsideColumn(t *testing.T) {
	h := newTestServer(t, "")

	a := decodeTask(t, do(t, h, http.MethodPost, "/api/tasks", map[string]any{"title": "A"}))
	decodeTask(t, do(t, h, http.MethodPost, "/api/tasks", map[string]any{"title": "B"}))
	decodeTask(t, do(t, h, http.MethodPost, "/api/tasks", map[string]any{"title": "C"}))

	// Leva A para o fim da coluna.
	rec := do(t, h, http.MethodPut, "/api/tasks/"+a.ID, map[string]any{"position": 2})
	if rec.Code != http.StatusOK {
		t.Fatalf("PUT = %d, quer 200 (body=%s)", rec.Code, rec.Body.String())
	}

	tasks := listTasks(t, h)
	got := []string{tasks[0].Title, tasks[1].Title, tasks[2].Title}
	want := []string{"B", "C", "A"}
	for i := range want {
		if got[i] != want[i] {
			t.Fatalf("ordem = %v, quer %v", got, want)
		}
		if tasks[i].Position != i {
			t.Errorf("position de %s = %d, quer %d", tasks[i].Title, tasks[i].Position, i)
		}
	}
}

func TestUpdateAndDeleteUnknownID(t *testing.T) {
	h := newTestServer(t, "")

	rec := do(t, h, http.MethodPut, "/api/tasks/nao-existe", map[string]any{"title": "x"})
	if rec.Code != http.StatusNotFound {
		t.Errorf("PUT inexistente = %d, quer 404", rec.Code)
	}

	rec = do(t, h, http.MethodDelete, "/api/tasks/nao-existe", nil)
	if rec.Code != http.StatusNotFound {
		t.Errorf("DELETE inexistente = %d, quer 404", rec.Code)
	}
}

func TestDeleteTask(t *testing.T) {
	h := newTestServer(t, "")

	created := decodeTask(t, do(t, h, http.MethodPost, "/api/tasks", map[string]any{"title": "Apagar"}))

	rec := do(t, h, http.MethodDelete, "/api/tasks/"+created.ID, nil)
	if rec.Code != http.StatusNoContent {
		t.Fatalf("DELETE = %d, quer 204", rec.Code)
	}
	if got := len(listTasks(t, h)); got != 0 {
		t.Errorf("tarefas restantes = %d, quer 0", got)
	}
}

func TestPersistenceSurvivesRestart(t *testing.T) {
	dataFile := filepath.Join(t.TempDir(), "tasks.json")

	h := newTestServer(t, dataFile)
	created := decodeTask(t, do(t, h, http.MethodPost, "/api/tasks", map[string]any{"title": "Persistir"}))
	do(t, h, http.MethodPut, "/api/tasks/"+created.ID, map[string]any{"status": "done"})

	if _, err := os.Stat(dataFile); err != nil {
		t.Fatalf("arquivo de dados nao foi criado: %v", err)
	}

	// Simula um restart do processo lendo o mesmo arquivo.
	reloaded := newTestServer(t, dataFile)
	tasks := listTasks(t, reloaded)
	if len(tasks) != 1 {
		t.Fatalf("tarefas apos restart = %d, quer 1", len(tasks))
	}
	if tasks[0].Title != "Persistir" || tasks[0].Status != StatusDone {
		t.Errorf("tarefa recarregada = %+v, quer titulo Persistir e status done", tasks[0])
	}
}

func TestCORSPreflight(t *testing.T) {
	h := newTestServer(t, "")

	req := httptest.NewRequest(http.MethodOptions, "/api/tasks", nil)
	req.Header.Set("Origin", "http://localhost:5173")
	req.Header.Set("Access-Control-Request-Method", "POST")
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)

	if rec.Code != http.StatusNoContent {
		t.Errorf("OPTIONS = %d, quer 204", rec.Code)
	}
	if got := rec.Header().Get("Access-Control-Allow-Origin"); got != "*" {
		t.Errorf("Allow-Origin = %q, quer *", got)
	}
	if got := rec.Header().Get("Access-Control-Allow-Methods"); got == "" {
		t.Error("Allow-Methods ausente na resposta de preflight")
	}
}

func TestRejectsInvalidJSON(t *testing.T) {
	h := newTestServer(t, "")

	req := httptest.NewRequest(http.MethodPost, "/api/tasks", bytes.NewReader([]byte("{nao e json")))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Errorf("POST json invalido = %d, quer 400", rec.Code)
	}
}
