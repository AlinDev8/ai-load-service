package analytics

import (
    "context"
    "fmt"
    "math"
    "sync"
    "time"
    
    "ai-load-service/internal/storage"
)

type Analyzer struct {
    redis          *storage.RedisClient
    metricsWindow  []storage.Metric
    windowSize     int
    mu             sync.RWMutex
    stopChan       chan struct{}
    rollingAvg     float64
    anomalyCount   int
}

func NewAnalyzer(redis *storage.RedisClient) *Analyzer {
    return &Analyzer{
        redis:      redis,
        windowSize: 50,
        metricsWindow: make([]storage.Metric, 0, 50),
        stopChan:   make(chan struct{}),
    }
}

func (a *Analyzer) ProcessMetric(metric storage.Metric) {
    a.mu.Lock()
    defer a.mu.Unlock()
    
    a.metricsWindow = append(a.metricsWindow, metric)
    if len(a.metricsWindow) > a.windowSize {
        a.metricsWindow = a.metricsWindow[1:]
    }
    
    a.calculateRollingAverage()
    
    if a.isAnomaly(metric) {
        a.anomalyCount++
        fmt.Printf("Чёт не то... CPU: %.2f, RPS: %.2f\n", metric.CPU, metric.RPS)
    }
}

func (a *Analyzer) calculateRollingAverage() {
    if len(a.metricsWindow) == 0 {
        a.rollingAvg = 0
        return
    }
    
    var sum float64
    for _, m := range a.metricsWindow {
        sum += m.RPS
    }
    a.rollingAvg = sum / float64(len(a.metricsWindow))
}

func (a *Analyzer) isAnomaly(metric storage.Metric) bool {
    if len(a.metricsWindow) < 10 {
        return false
    }
    
    var sum, sqSum float64
    for _, m := range a.metricsWindow {
        sum += m.RPS
        sqSum += m.RPS * m.RPS
    }
    
    n := float64(len(a.metricsWindow))
    mean := sum / n
    variance := (sqSum / n) - (mean * mean)
    stdDev := math.Sqrt(math.Max(variance, 0))
    
    if stdDev == 0 {
        return false
    }
    
    zScore := math.Abs((metric.RPS - mean) / stdDev)
    
    return zScore > 2.0
}

func (a *Analyzer) GetCurrentAnalysis() map[string]interface{} {
    a.mu.RLock()
    defer a.mu.RUnlock()
    
    return map[string]interface{}{
        "rolling_average":   a.rollingAvg,
        "window_size":       len(a.metricsWindow),
        "anomaly_count":     a.anomalyCount,
        "total_metrics":     len(a.metricsWindow),
        "last_updated":      time.Now().Format(time.RFC3339),
    }
}

func (a *Analyzer) Run(ctx context.Context) {
    fmt.Println("Аналитический анализатор запущен")
    
    for {
        select {
        case <-a.stopChan:
            fmt.Println("Аналитический анализатор остановлен")
            return
        case <-time.After(30 * time.Second):
            a.cleanOldMetrics()
        }
    }
}

func (a *Analyzer) Stop() {
    close(a.stopChan)
}

func (a *Analyzer) cleanOldMetrics() {
    fmt.Println("Очистка старых метрик...")
}
