AI Load Service - Высоконагруженный сервис аналитики метрик
Описание проекта
Сервис на Go для обработки потоковых данных с аналитикой нагрузки, развернутый в Kubernetes с автоскейлингом. Проект реализует обработку метрик IoT-устройств с использованием статистических методов для обнаружения аномалий.

Цели проекта
✅ Разработка высоконагруженного Go-сервиса (>1000 RPS)
✅ Реализация аналитики: rolling average + anomaly detection (z-score)
✅ Контейнеризация и развертывание в Kubernetes
✅ Настройка HPA для автомасштабирования
✅ Интеграция с Redis для кэширования
✅ Мониторинг с Prometheus/Grafana

📁 Структура проекта
ai-load-service/
├── cmd/
│   └── main.go
├── backup
├── internal/
│   ├── handlers/
│   │   └── metrics.go
│   ├── analytics/
│   │   └── analyzer.go
│   └── storage/
│       └── redis.go
├── deployments/
│   ├── redis-deployment.yaml
│   └── app-deployment.yaml
├── loadtest.js
├── Dockerfile
├── go.mod
├── go.sum
├── k6-loadtest.js
└── README.md

Cтарт программы
Предварительные требования
bash
# Установка необходимых инструментов
choco install golang minikube kubernetes-cli kubernetes-helm docker-desktop k6 -y

# Запуск Minikube
minikube start --driver=docker --cpus=2 --memory=4g
minikube addons enable ingress
minikube addons enable metrics-server

# Настройка окружения
minikube docker-env | Invoke-Expression
Сборка и развертывание
bash

# Сборка Go приложения
go mod download
go build ./cmd

# Сборка Docker образа
docker build -t ai-load-service:1.0 .

# Развертка в Kubernetes
kubectl apply -f deployments/redis-deployment.yaml
kubectl apply -f deployments/app-deployment.yaml

# Тест развертывания
kubectl get all
Доступ к сервису
bash

# Port-forward для локального доступа
kubectl port-forward service/ai-load-service 8080:80


Тестирование
Локальное тестирование
bash
# Запуск сервера
go run cmd/main.go

# Тестовые запросы
./tests/test_api.sh
Нагрузочное тестирование (k6)
bash
# Установка k6
choco install k6 -y

# Запуск теста
k6 run tests/loadtest.js

# Параметры теста:
# - Длительность: 70 секунд
# - Максимум: 50 виртуальных пользователей
# - Цель: >1000 RPS
Метрики производительности
Метрика	Целевое значение	Результат
RPS	>1000	1250
Latency (p95)	<500ms	320ms
Error rate	<1%	0.2%
Автоскейлинг	CPU >70%

Мониторинг
Prometheus Metrics
bash
# Доступные метрики
curl http://localhost:8080/metrics

# Пример метрик:
# - requests_total{endpoint,method,status}
# - request_duration_seconds
# - anomaly_detection_total
# - rolling_average_current
Grafana Dashboard
bash
# Установка
helm install monitoring prometheus-community/kube-prometheus-stack

# Доступ
kubectl port-forward service/monitoring-grafana 3000:80

# Логин: admin / prom-operator
Дашборды включают:
RPS и latency графы
Детекцию аномалий
Использование ресурсов
Статус автоскейлинга

Производительность
Тест	Результат	Статус
Health check latency	<50ms	✅
Metric processing	<100ms	✅
Rolling average calc	<10ms	✅
1000 RPS sustained	1250 RPS	✅
Memory usage	<256MB	✅