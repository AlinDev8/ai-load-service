package main

import (
    "encoding/json"
    "fmt"
    "net/http"
    "log"
    "time"
    "math"
    "sync"
)

var (
    metrics     []float64
    metricsLock sync.RWMutex
    anomalyCount int
    windowSize   = 50
)

func main() {
    metrics = []float64{100, 110, 105, 115, 120, 500, 130, 125, 118, 122}
    
    http.HandleFunc("/health", healthHandler)
    http.HandleFunc("/analyze", analyzeHandler)
    http.HandleFunc("/metric", metricHandler)
    http.HandleFunc("/", rootHandler)
    
    port := ":8080"
    log.Printf("Сервис загрузки ИИ запущен на %s", port)

    server := &http.Server{
        Addr:         port,
        ReadTimeout:  5 * time.Second,
        WriteTimeout: 10 * time.Second,
        IdleTimeout:  15 * time.Second,
    }

    err := server.ListenAndServe()
    if err != nil {
        log.Fatalf("Сервер лежит: %v", err)
    }
}

func healthHandler(w http.ResponseWriter, r *http.Request) {
    w.Header().Set("Content-Type", "application/json")
    json.NewEncoder(w).Encode(map[string]interface{}{
        "status":    "healthy",
        "service":   "ai-load-service",
        "timestamp": time.Now().Format(time.RFC3339),
        "version":   "1.0.0",
    })
}

func analyzeHandler(w http.ResponseWriter, r *http.Request) {
    metricsLock.RLock()
    defer metricsLock.RUnlock()
    
    if len(metrics) == 0 {
        w.Header().Set("Content-Type", "application/json")
        json.NewEncoder(w).Encode(map[string]interface{}{
            "error": "no metrics yet",
            "timestamp": time.Now().Format(time.RFC3339),
        })
        return
    }
    
    var sum float64
    for _, m := range metrics {
        sum += m
    }
    rollingAvg := sum / float64(len(metrics))
    
    var variance float64
    for _, m := range metrics {
        diff := m - rollingAvg
        variance += diff * diff
    }
    variance /= float64(len(metrics))
    stdDev := math.Sqrt(variance)
    
    currentAnomalies := 0
    if stdDev > 0 {
        for _, m := range metrics {
            zScore := math.Abs((m - rollingAvg) / stdDev)
            if zScore > 2.0 {
                currentAnomalies++
            }
        }
    }
    
    response := map[string]interface{}{
        "rolling_average":    rollingAvg,
        "standard_deviation": stdDev,
        "anomaly_count":      currentAnomalies,
        "total_anomalies":    anomalyCount,
        "window_size":        len(metrics),
        "total_metrics":      len(metrics),
        "min_value":          findMin(),
        "max_value":          findMax(),
        "timestamp":          time.Now().Format(time.RFC3339),
        "algorithm":          "z-score (threshold > 2σ)",
    }
    
    w.Header().Set("Content-Type", "application/json")
    json.NewEncoder(w).Encode(response)
    
    log.Printf("Анализ: avg=%.2f, std=%.2f, anomalies=%d", 
        rollingAvg, stdDev, currentAnomalies)
}

func metricHandler(w http.ResponseWriter, r *http.Request) {
    var data map[string]float64
    
    if err := json.NewDecoder(r.Body).Decode(&data); err != nil {
        w.Header().Set("Content-Type", "application/json")
        w.WriteHeader(http.StatusBadRequest)
        json.NewEncoder(w).Encode(map[string]string{
            "error": "Неверный формат JSON",
        })
        return
    }
    
    rps, hasRPS := data["rps"]
    if !hasRPS {
        rps = 100.0
    }
    
    metricsLock.Lock()
    metrics = append(metrics, rps)
    if len(metrics) > windowSize {
        metrics = metrics[1:]
    }
    
    if len(metrics) >= 10 {
        mean, stdDev := calculateStats()
        if stdDev > 0 {
            zScore := math.Abs((rps - mean) / stdDev)
            if zScore > 2.0 {
                anomalyCount++
                log.Printf("Что-то не то: RPS=%.2f, z-score=%.2f", rps, zScore)
            }
        }
    }
    metricsLock.Unlock()
    
    w.Header().Set("Content-Type", "application/json")
    json.NewEncoder(w).Encode(map[string]interface{}{
        "status":    "received",
        "rps":       rps,
        "timestamp": time.Now().Format(time.RFC3339),
        "window":    len(metrics),
    })
    
    log.Printf("Полученные метрики: RPS=%.2f", rps)
}

func rootHandler(w http.ResponseWriter, r *http.Request) {
    fmt.Fprintf(w, "AI Load Service API v1.0\n\n")
    fmt.Fprintf(w, "Endpoints:\n")
    fmt.Fprintf(w, "  GET  /health  - Health check (returns JSON)\n")
    fmt.Fprintf(w, "  GET  /analyze - Analytics with rolling average (returns JSON)\n")
    fmt.Fprintf(w, "  POST /metric  - Submit metrics (accepts JSON)\n")
    fmt.Fprintf(w, "\nFeatures:\n")
    fmt.Fprintf(w, "  • Rolling average with window size %d\n", windowSize)
    fmt.Fprintf(w, "  • Anomaly detection using z-score (threshold > 2σ)\n")
}

func calculateStats() (mean, stdDev float64) {
    if len(metrics) == 0 {
        return 0, 0
    }
    
    var sum float64
    for _, m := range metrics {
        sum += m
    }
    mean = sum / float64(len(metrics))
    
    var variance float64
    for _, m := range metrics {
        diff := m - mean
        variance += diff * diff
    }
    variance /= float64(len(metrics))
    
    return mean, math.Sqrt(variance)
}

func findMin() float64 {
    if len(metrics) == 0 { return 0 }
    min := metrics[0]
    for _, m := range metrics[1:] {
        if m < min { min = m }
    }
    return min
}

func findMax() float64 {
    if len(metrics) == 0 { return 0 }
    max := metrics[0]
    for _, m := range metrics[1:] {
        if m > max { max = m }
    }
    return max
}
