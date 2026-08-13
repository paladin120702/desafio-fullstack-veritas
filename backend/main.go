package main

import (
	"context"
	"errors"
	"log"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"
)

func main() {
	port := env("PORT", "8080")
	// DATA_FILE vazio mantem tudo apenas em memoria (escopo minimo do desafio).
	dataFile := env("DATA_FILE", "data/tasks.json")
	corsOrigins := splitAndTrim(env("CORS_ORIGINS", "*"))

	store, err := NewStore(dataFile)
	if err != nil {
		log.Fatalf("falha ao iniciar o armazenamento: %v", err)
	}

	handler := withRecover(withLogging(withCORS(corsOrigins, NewAPI(store).Routes())))

	server := &http.Server{
		Addr:              ":" + port,
		Handler:           handler,
		ReadHeaderTimeout: 5 * time.Second,
		ReadTimeout:       15 * time.Second,
		WriteTimeout:      15 * time.Second,
		IdleTimeout:       60 * time.Second,
	}

	// Encerramento gracioso: para de aceitar conexoes e espera as em andamento.
	shutdown := make(chan os.Signal, 1)
	signal.Notify(shutdown, os.Interrupt, syscall.SIGTERM)

	go func() {
		storage := dataFile
		if storage == "" {
			storage = "somente memoria"
		}
		log.Printf("API do Kanban ouvindo em http://localhost:%s (armazenamento: %s)", port, storage)
		if err := server.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			log.Fatalf("servidor encerrou com erro: %v", err)
		}
	}()

	<-shutdown
	log.Println("encerrando servidor...")

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	if err := server.Shutdown(ctx); err != nil {
		log.Printf("encerramento forcado: %v", err)
	}
	log.Println("servidor encerrado")
}

func env(key, fallback string) string {
	if v, ok := os.LookupEnv(key); ok {
		return v
	}
	return fallback
}

func splitAndTrim(raw string) []string {
	parts := strings.Split(raw, ",")
	out := make([]string, 0, len(parts))
	for _, p := range parts {
		if p = strings.TrimSpace(p); p != "" {
			out = append(out, p)
		}
	}
	return out
}
