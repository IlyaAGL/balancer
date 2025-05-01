# Load Balancer with Rate Limiting (Least Connections + Token Bucket)

Этот проект представляет собой HTTP-балансировщик нагрузки, реализующий алгоритм наименьшего количества соединений (Least Connections) с поддержкой rate limiting на основе Token Bucket, health-чеками и конфигурацией через `config.json`.

## Возможности

- Алгоритм балансировки: Least Connections  
- Обработка отказов бэкендов  
- Rate Limiting по IP на основе Token Bucket  
- Health checks с заданным интервалом  
- Конфигурация через JSON или переменные окружения  
- Поддержка Docker и Docker Compose  

## Сборка

### Через Go

```bash
git clone https://github.com/IlyaAGL/balancer.git
cd balancer
go build -o balancer ./cmd/main.go
```

### Через Docker
```bash
docker build -t balancer .
```

### Запуск

#### Локально
```bash
go run ./cmd/main.go --config=config/config.json
```

#### Или
```bash
export BACKEND_CONFIG=config/config.json
go run ./cmd/main.go
```

#### Через Docker Compose
```bash
docker-compose up --build
```
Сервис будет доступен по адресу: http://localhost:8080

## Конфигурация

```json
{
    "port": "8080",
    "healthCheckInterval": "5s",
    "rateLimit": {
      "enabled": true,
      "defaultCapacity": 10,
      "defaultRate": 5
    },
    "backends": [
      {
        "url": "http://backend1:80",
        "weight": 1,
        "name": "backend-1"
      },
      {
        "url": "http://backend2:80",
        "weight": 2,
        "name": "backend-2"
      }
    ]
  }
```

## Пример запроса
```bash
curl http://localhost:8080
```
Если превышен лимит:
```json
{ "error": "rate limit exceeded" }
```