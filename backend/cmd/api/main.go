package main

import (
	"context"
	"errors"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/go-chi/chi/v5"

	"vtb-copilot-backend/internal/handler"
	"vtb-copilot-backend/internal/middleware"
)

func main() {
	h, err := handler.New("")
	if err != nil {
		log.Fatal(err)
	}

	r := chi.NewRouter()
	r.Use(middleware.CORS)
	r.Get("/", h.Root)
	r.Route("/api/v1", func(r chi.Router) {
		r.Get("/cases", h.GetDemoCases)
		r.Post("/search", h.SearchChunks)
		r.Post("/copilot/resolve", h.ResolveCopilot)
		r.Post("/sandbox/evaluate", h.EvaluateSandbox)
	})

	server := &http.Server{
		Addr:              ":8000",
		Handler:           r,
		ReadHeaderTimeout: 5 * time.Second,
		ReadTimeout:       30 * time.Second,
		WriteTimeout:      30 * time.Second,
		IdleTimeout:       60 * time.Second,
	}
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	go func() {
		log.Printf("API слушает %s", server.Addr)
		if err := server.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			log.Fatal(err)
		}
	}()

	<-ctx.Done()
	shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	if err := server.Shutdown(shutdownCtx); err != nil {
		log.Printf("Ошибка завершения API: %v", err)
	}
}
