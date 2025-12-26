package main

import (
    "context"
    "log"
    "net/http"
    "os"
    "os/signal"
    "syscall"
    "time"
    
    "ai-load-service/internal/analytics"
    "ai-load-service/internal/handlers"
    "ai-load-service/internal/storage"
    
    "github.com/go-chi/chi/v5"
    "github.com/go-chi/chi/v5/middleware"
)

func main() {
    // Redis клиент (заглушка пока)
    redisClient := storage.NewRedisClient("localhost:6379", "")
    defer redisClient.Close()
    
    // Инициализация аналитики
    analyzer := analytics.NewAnalyzer(redisClient)
    go analyzer.Run(context.Background())
    
    // Роутер
    r := chi.NewRouter()
    r.Use(middleware.Logger)
    r.Use(middleware.Recoverer)
    
    // Регистрация обработчиков
    h := handlers.NewHandler(redisClient, analyzer)
    r.Post("/metric", h.HandleMetric)
    r.Get("/analyze", h.GetAnalysis)
    r.Get("/health", h.HealthCheck)
    
    // Graceful shutdown
    srv := &http.Server{
        Addr:    ":8080",
        Handler: r,
    }
    
    go func() {
        log.Println("Сервер запущен на :8080")
        if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
            log.Fatal("Сервер не взлетел:", err)
        }
    }()
    
    // Ожидание сигнала завершения
    quit := make(chan os.Signal, 1)
    signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
    <-quit
    
    ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
    defer cancel()
    
    analyzer.Stop()
    if err := srv.Shutdown(ctx); err != nil {
        log.Fatal("Сервер ушёл в разнос:", err)
    }
    log.Println("Сервер остановился")
}
