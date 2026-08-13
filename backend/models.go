package main

import (
	"crypto/rand"
	"encoding/hex"
	"strings"
	"time"
)

// Status representa uma das tres colunas fixas do quadro.
type Status string

const (
	StatusTodo  Status = "todo"  // A Fazer
	StatusDoing Status = "doing" // Em Progresso
	StatusDone  Status = "done"  // Concluidas
)

// Statuses mantem a ordem canonica das colunas.
var Statuses = []Status{StatusTodo, StatusDoing, StatusDone}

// Valid informa se o status recebido e uma das colunas conhecidas.
func (s Status) Valid() bool {
	for _, known := range Statuses {
		if s == known {
			return true
		}
	}
	return false
}

const (
	maxTitleLen       = 120
	maxDescriptionLen = 1000
)

// Task e a unidade de trabalho exibida no quadro.
type Task struct {
	ID          string    `json:"id"`
	Title       string    `json:"title"`
	Description string    `json:"description"`
	Status      Status    `json:"status"`
	Position    int       `json:"position"`
	CreatedAt   time.Time `json:"createdAt"`
	UpdatedAt   time.Time `json:"updatedAt"`
}

// CreateTaskInput e o corpo aceito no POST /api/tasks.
type CreateTaskInput struct {
	Title       string  `json:"title"`
	Description string  `json:"description"`
	Status      *Status `json:"status"`
}

// UpdateTaskInput e o corpo aceito no PUT /api/tasks/{id}.
// Campos ausentes (nil) preservam o valor atual da tarefa.
type UpdateTaskInput struct {
	Title       *string `json:"title"`
	Description *string `json:"description"`
	Status      *Status `json:"status"`
	Position    *int    `json:"position"`
}

// FieldErrors mapeia campo -> motivo da recusa.
type FieldErrors map[string]string

// Validate aplica as regras de negocio da criacao.
func (in *CreateTaskInput) Validate() FieldErrors {
	errs := FieldErrors{}

	in.Title = strings.TrimSpace(in.Title)
	in.Description = strings.TrimSpace(in.Description)

	if in.Title == "" {
		errs["title"] = "titulo e obrigatorio"
	} else if len([]rune(in.Title)) > maxTitleLen {
		errs["title"] = "titulo deve ter no maximo 120 caracteres"
	}

	if len([]rune(in.Description)) > maxDescriptionLen {
		errs["description"] = "descricao deve ter no maximo 1000 caracteres"
	}

	if in.Status != nil && !in.Status.Valid() {
		errs["status"] = "status deve ser todo, doing ou done"
	}

	if len(errs) == 0 {
		return nil
	}
	return errs
}

// Validate aplica as regras de negocio da atualizacao parcial.
func (in *UpdateTaskInput) Validate() FieldErrors {
	errs := FieldErrors{}

	if in.Title != nil {
		trimmed := strings.TrimSpace(*in.Title)
		in.Title = &trimmed
		if trimmed == "" {
			errs["title"] = "titulo e obrigatorio"
		} else if len([]rune(trimmed)) > maxTitleLen {
			errs["title"] = "titulo deve ter no maximo 120 caracteres"
		}
	}

	if in.Description != nil {
		trimmed := strings.TrimSpace(*in.Description)
		in.Description = &trimmed
		if len([]rune(trimmed)) > maxDescriptionLen {
			errs["description"] = "descricao deve ter no maximo 1000 caracteres"
		}
	}

	if in.Status != nil && !in.Status.Valid() {
		errs["status"] = "status deve ser todo, doing ou done"
	}

	if in.Position != nil && *in.Position < 0 {
		errs["position"] = "position nao pode ser negativa"
	}

	if len(errs) == 0 {
		return nil
	}
	return errs
}

// newID gera um identificador hexadecimal de 16 caracteres.
func newID() string {
	buf := make([]byte, 8)
	if _, err := rand.Read(buf); err != nil {
		// crypto/rand so falha em ambiente sem fonte de entropia;
		// nesse caso o timestamp em nanossegundos ainda garante unicidade local.
		return hex.EncodeToString([]byte(time.Now().Format(time.RFC3339Nano)))
	}
	return hex.EncodeToString(buf)
}
