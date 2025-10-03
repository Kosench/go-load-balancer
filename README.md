# go-load-balancer

Простой HTTP load balancer на Go с round-robin балансировкой и rate limiting.

## Возможности

- Round-robin распределение запросов между бэкендами
- Health checks для автоматического отключения недоступных бэкендов
- Rate limiting на основе API ключей
- Graceful shutdown
- Управление клиентами через REST API

## Установка и запуск

```bash
# Клонировать репозиторий
git clone <repo-url>
cd go-load-balancer

# Скачать зависимости
go mod download

# Запустить
go run cmd/loadbalancer/main.go
```

## Конфигурация

Создайте файл `app.env` в корне проекта (или используйте переменные окружения):

```env
LISTEN_ADDRESS=:8080
BACKENDS=localhost:9001,localhost:9002,localhost:9003
RATE_LIMIT_CAPACITY=5
RATE_LIMIT_REFILL_RATE=1
```

Если файл не найден, используются значения по умолчанию.

## Использование

### Создание клиента

```bash
curl -X POST http://localhost:8080/clients \
  -H "Content-Type: application/json" \
  -d '{
    "client_id": "user1",
    "capacity": 10,
    "rate_per_sec": 2,
    "api_key": "secret-key-123"
  }'
```

### Отправка запроса

```bash
curl http://localhost:8080/ \
  -H "X-API-Key: secret-key-123"
```

### Управление клиентами

```bash
# Получить список клиентов
curl http://localhost:8080/clients

# Получить клиента по ID
curl http://localhost:8080/clients/user1

# Удалить клиента
curl -X DELETE http://localhost:8080/clients/user1
```

## Структура проекта

```
.
├── cmd/loadbalancer/     # Точка входа
├── internal/
│   ├── backend/          # Управление бэкендами
│   ├── balancer/         # Стратегии балансировки
│   ├── client/           # Управление клиентами
│   ├── config/           # Конфигурация
│   ├── health/           # Health checks
│   └── server/           # HTTP сервер и middleware
└── app.env              # Конфигурационный файл
```

## Тесты

```bash
go test ./...
```
