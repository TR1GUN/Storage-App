package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"storage_app/handlers"
	"storage_app/storage"
	"syscall"
	"time"
)

// StartServer — запускаем HTTP сервер
func StartServer(addr string, sm *storage.StorageManager) error {
	// Инициализация хранилища
	if err := sm.Startup(); err != nil {
		return err
	}
	// Инициализация роутов
	mux := http.NewServeMux()
	handlers.Routes(mux, sm)

	srv := &http.Server{
		Addr:    addr,
		Handler: mux,
	}

	// Канал для сигналов ОС
	stop := make(chan os.Signal, 1)
	signal.Notify(stop, os.Interrupt, syscall.SIGTERM)

	// Запуск сервера в горутине
	go func() {
		log.Printf("Server starting on %s", addr)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("Server error: %v", err)
		}
	}()

	// Ожидание сигнала остановки
	<-stop
	log.Println("Shutting down server...")

	// Graceful shutdown с таймаутом 5 секунд
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := srv.Shutdown(ctx); err != nil {
		log.Printf("Server forced to shutdown: %v", err)
	}

	if err := sm.Shutdown(); err != nil {
		log.Printf("Storage shutdown error: %v", err)
	}

	log.Println("Server stopped")
	return nil
}

func main() {
	addr := ":8000"
	if err := StartServer(addr, storage.NewStorageManager()); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}

//func main() {
//	sm := storage.NewStorageManager()
//
//	if err := sm.Startup(); err != nil {
//		log.Fatalf("Startup failed: %v", err)
//	}
//
//	mux := http.NewServeMux()
//	handlers.Routes(mux, sm)
//
//	log.Println("Server starting on :8000")
//	if err := http.ListenAndServe(":8000", mux); err != nil {
//		log.Fatalf("Server error: %v", err)
//	}
//}
