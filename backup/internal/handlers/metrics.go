package handlers

import (
    "encoding/json"
    "fmt"
    "net/http"
    "time"
    
    "ai-load-service/internal/analytics"
    "ai-load-service/internal/storage"
)

type Handler struct {
    redis    *storage.RedisClient
    analyzer *analytics.Analyzer
}

func NewHandler(redis *storage.RedisClient, analyzer *analytics.Analyzer) *Handler {
    return &Handler{redis: redis, analyzer: analyzer}
}

type MetricRequest struct {
    Timestamp int64   `json:"timestamp"`
    CPU       float64 `json:"cpu"`
    RPS       float64 `json:"rps"`
}

type MetricResponse struct {
    ID            string  `json:"id"`
    Status        string  `json:"status"`
    RollingAvg    float64 `json:"rolling_avg,omitempty"`
    IsAnomaly     bool    `json:"is_anomaly,omitempty"`
    ReceivedAt    string  `json:"received_at"`
}

func (h *Handler) HandleMetric(w http.ResponseWriter, r *http.Request) {
    var req MetricRequest
    if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
        http.Error(w, `{"error": "Invalid JSON"}`, http.StatusBadRequest)
        return
    }
    
    // Если timestamp не указан, используем текущее время
    if req.Timestamp == 0 {
        req.Timestamp = time.Now().Unix()
    }
    
    metric := storage.Metric{
        Timestamp: req.Timestamp,
        CPU:       req.CPU,
        RPS:       req.RPS,
    }
    
    // Сохраняем в Redis
    ctx := r.Context()
    if err := h.redis.StoreMetric(ctx, &metric); err != nil {
        http.Error(w, `{"error": "Storage error"}`, http.StatusInternalServerError)
        return
    }
    
    // Обрабатываем в анализаторе
    h.analyzer.ProcessMetric(metric)
    
    // Получаем текущий анализ для ответа
    analysis := h.analyzer.GetCurrentAnalysis()
    rollingAvg, _ := analysis["rolling_average"].(float64)
    
    resp := MetricResponse{
        ID:         fmt.Sprintf("metric_%d", req.Timestamp),
        Status:     "processed",
        RollingAvg: rollingAvg,
        ReceivedAt: time.Now().Format(time.RFC3339),
    }
    
    w.Header().Set("Content-Type", "application/json")
    json.NewEncoder(w).Encode(resp)
}

func (h *Handler) GetAnalysis(w http.ResponseWriter, r *http.Request) {
    analysis := h.analyzer.GetCurrentAnalysis()
    w.Header().Set("Content-Type", "application/json")
    json.NewEncoder(w).Encode(analysis)
}

func (h *Handler) HealthCheck(w http.ResponseWriter, r *http.Request) {
    w.Header().Set("Content-Type", "application/json")
    json.NewEncoder(w).Encode(map[string]interface{}{
        "status":   "healthy",
        "service":  "ai-load-service",
        "version":  "1.0.0",
        "datetime": time.Now().Format(time.RFC3339),
    })
}
