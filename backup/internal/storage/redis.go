package storage

import (
    "context"
    "fmt"
    "time"
)

type Metric struct {
    Timestamp int64   `json:"timestamp"`
    CPU       float64 `json:"cpu"`
    RPS       float64 `json:"rps"`
}

type RedisClient struct {
    addr string
}

func NewRedisClient(addr, password string) *RedisClient {
    fmt.Printf("Клиент Redis подключается к: %s\n", addr)
    return &RedisClient{addr: addr}
}

func (r *RedisClient) StoreMetric(ctx context.Context, metric *Metric) error {
    fmt.Printf("Storing метрики: CPU=%.2f, RPS=%.2f at %d\n", 
        metric.CPU, metric.RPS, metric.Timestamp)
    return nil
}

func (r *RedisClient) GetRecentMetrics(ctx context.Context, limit int) ([]Metric, error) {
    metrics := []Metric{
        {Timestamp: time.Now().Add(-10*time.Second).Unix(), CPU: 45.5, RPS: 120},
        {Timestamp: time.Now().Add(-9*time.Second).Unix(), CPU: 48.2, RPS: 125},
    }
    return metrics, nil
}

func (r *RedisClient) Close() {
    fmt.Println("Клиент Redis остановлен")
}
