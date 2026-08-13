package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"sync"
	"time"
)

// ErrNotFound e devolvido quando o ID informado nao existe no quadro.
var ErrNotFound = errors.New("tarefa nao encontrada")

// Store guarda as tarefas em memoria.
//
// A ordem do slice e a propria ordem do quadro: a posicao de uma tarefa dentro
// da coluna e o indice dela entre as tarefas de mesmo status. Isso evita manter
// contadores de posicao sincronizados e faz o reordenamento ser um simples
// "remove e insere".
//
// Quando filePath nao esta vazio, cada mutacao tambem grava o slice em disco,
// entao a ordem sobrevive a um restart.
type Store struct {
	mu       sync.RWMutex
	tasks    []*Task
	filePath string
	now      func() time.Time
}

// NewStore cria um store em memoria. Se filePath nao for vazio, o arquivo e
// carregado no boot (quando existe) e reescrito a cada mutacao.
func NewStore(filePath string) (*Store, error) {
	s := &Store{
		tasks:    []*Task{},
		filePath: filePath,
		now:      time.Now,
	}
	if filePath == "" {
		return s, nil
	}
	if err := s.load(); err != nil {
		return nil, err
	}
	return s, nil
}

// List devolve todas as tarefas na ordem do quadro, com Position preenchida.
func (s *Store) List() []Task {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.snapshot()
}

// Create adiciona uma tarefa no fim da coluna informada.
func (s *Store) Create(in CreateTaskInput) (Task, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	status := StatusTodo
	if in.Status != nil {
		status = *in.Status
	}

	now := s.now().UTC()
	task := &Task{
		ID:          newID(),
		Title:       in.Title,
		Description: in.Description,
		Status:      status,
		CreatedAt:   now,
		UpdatedAt:   now,
	}

	s.tasks = append(s.tasks, task)
	if err := s.persist(); err != nil {
		s.tasks = s.tasks[:len(s.tasks)-1]
		return Task{}, err
	}
	return *s.withPosition(task), nil
}

// Update aplica uma atualizacao parcial e, se necessario, reposiciona a tarefa.
func (s *Store) Update(id string, in UpdateTaskInput) (Task, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	idx := s.indexOf(id)
	if idx < 0 {
		return Task{}, ErrNotFound
	}

	task := s.tasks[idx]
	backup := *task
	backupOrder := append([]*Task(nil), s.tasks...)

	if in.Title != nil {
		task.Title = *in.Title
	}
	if in.Description != nil {
		task.Description = *in.Description
	}

	targetStatus := task.Status
	if in.Status != nil {
		targetStatus = *in.Status
	}

	// Reposiciona quando muda de coluna ou quando uma posicao explicita chega.
	if in.Status != nil || in.Position != nil {
		position := -1 // -1 = fim da coluna
		if in.Position != nil {
			position = *in.Position
		}
		s.move(idx, targetStatus, position)
	}

	task.UpdatedAt = s.now().UTC()

	if err := s.persist(); err != nil {
		*task = backup
		s.tasks = backupOrder
		return Task{}, err
	}
	return *s.withPosition(task), nil
}

// Delete remove a tarefa do quadro.
func (s *Store) Delete(id string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	idx := s.indexOf(id)
	if idx < 0 {
		return ErrNotFound
	}

	removed := s.tasks[idx]
	s.tasks = append(s.tasks[:idx], s.tasks[idx+1:]...)
	if err := s.persist(); err != nil {
		s.tasks = append(s.tasks, nil)
		copy(s.tasks[idx+1:], s.tasks[idx:])
		s.tasks[idx] = removed
		return err
	}
	return nil
}

// --- internos (exigem lock ja adquirido) ---

func (s *Store) indexOf(id string) int {
	for i, t := range s.tasks {
		if t.ID == id {
			return i
		}
	}
	return -1
}

// move retira a tarefa em from e a reinsere na coluna status.
// position < 0 ou maior que o tamanho da coluna significa "fim da coluna".
func (s *Store) move(from int, status Status, position int) {
	task := s.tasks[from]
	s.tasks = append(s.tasks[:from], s.tasks[from+1:]...)
	task.Status = status

	// Indices globais das tarefas ja presentes na coluna de destino.
	var column []int
	for i, t := range s.tasks {
		if t.Status == status {
			column = append(column, i)
		}
	}

	insertAt := len(s.tasks)
	switch {
	case position >= 0 && position < len(column):
		insertAt = column[position]
	case len(column) > 0:
		insertAt = column[len(column)-1] + 1
	}

	s.tasks = append(s.tasks, nil)
	copy(s.tasks[insertAt+1:], s.tasks[insertAt:])
	s.tasks[insertAt] = task
}

// withPosition devolve a tarefa com o campo Position recalculado.
func (s *Store) withPosition(task *Task) *Task {
	pos := 0
	for _, t := range s.tasks {
		if t.Status != task.Status {
			continue
		}
		if t.ID == task.ID {
			break
		}
		pos++
	}
	task.Position = pos
	return task
}

func (s *Store) snapshot() []Task {
	counters := map[Status]int{}
	out := make([]Task, 0, len(s.tasks))
	for _, t := range s.tasks {
		copied := *t
		copied.Position = counters[t.Status]
		counters[t.Status]++
		out = append(out, copied)
	}
	return out
}

// --- persistencia opcional em arquivo JSON ---

func (s *Store) load() error {
	data, err := os.ReadFile(s.filePath)
	if errors.Is(err, os.ErrNotExist) {
		return nil // primeiro boot: comeca com quadro vazio
	}
	if err != nil {
		return fmt.Errorf("lendo %s: %w", s.filePath, err)
	}
	if len(data) == 0 {
		return nil
	}

	var tasks []*Task
	if err := json.Unmarshal(data, &tasks); err != nil {
		return fmt.Errorf("json invalido em %s: %w", s.filePath, err)
	}

	// Descarta registros corrompidos em vez de derrubar o servidor.
	for _, t := range tasks {
		if t == nil || t.ID == "" || !t.Status.Valid() {
			continue
		}
		s.tasks = append(s.tasks, t)
	}
	return nil
}

// persist grava o estado atual de forma atomica (arquivo temporario + rename).
func (s *Store) persist() error {
	if s.filePath == "" {
		return nil
	}

	dir := filepath.Dir(s.filePath)
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return fmt.Errorf("criando %s: %w", dir, err)
	}

	data, err := json.MarshalIndent(s.snapshot(), "", "  ")
	if err != nil {
		return fmt.Errorf("serializando tarefas: %w", err)
	}

	tmp, err := os.CreateTemp(dir, ".tasks-*.json")
	if err != nil {
		return fmt.Errorf("criando arquivo temporario: %w", err)
	}
	tmpName := tmp.Name()

	if _, err := tmp.Write(data); err != nil {
		tmp.Close()
		os.Remove(tmpName)
		return fmt.Errorf("gravando arquivo temporario: %w", err)
	}
	if err := tmp.Close(); err != nil {
		os.Remove(tmpName)
		return fmt.Errorf("fechando arquivo temporario: %w", err)
	}
	if err := os.Rename(tmpName, s.filePath); err != nil {
		os.Remove(tmpName)
		return fmt.Errorf("substituindo %s: %w", s.filePath, err)
	}
	return nil
}
