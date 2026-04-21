# ITS Reg

REST API для создания и управления Telegram-ботами на основе конечных автоматов. 
Платформа позволяет определять ботов, сценарии их поведения и запускать обработку сообщений через заданные сценарии.

## Требования

- Go 1.25+
- PostgreSQL
- Docker / Docker Compose (опционально)

## Структура проекта

```
api/openapi/              - спецификация OpenAPI
cmd/http/                 - точка входа приложения
config/                   - файлы конфигурации
internal/
  ├── api/v3/             - HTTP API (OpenAPI/Chi)
  ├── app/                - бизнес-логика
  │   ├── command/        - обработчики команд
  │   ├── dispatcher/     - диспетчеры команд и событий
  │   ├── dto/            - DTO прикладного слоя
  │   ├── eventhandler/   - обработчики событий
  │   ├── port/           - интерфейсы портов
  │   └── query/          - обработчики запросов
  ├── config/             - конфигурация
  ├── domain/             - модели доменов
  └── infra/              - реализация портов
      ├── postgres/       - хранилище БД
      ├── inmemory/       - шина событий
      ├── telegram/       - интеграция Telegram
      ├── jwt/            - генерация токенов
      └── mutexlimiter/   - rate limiting
migrations/               - миграции БД
pkg/                      - общие пакеты
```

## Установка и запуск

### Docker Compose

```bash
docker-compose up
```

### Локально

1. Подготовить .env:
   ```bash 
   cp .env.example .env # отредактировать при необходимости
   ```

2. Запустить через Docker Compose:
   ```bash
   docker compose --env-file .env up -d
   ```
   
### Локально через go run

1. Подготовить .env:
   ```bash 
   cp .env.example .env # отредактировать при необходимости
   ```
   
2. Поднять PostgreSQL и применить миграции:
   ```bash
   docker compose --env-file .env up -d postgres
   docker compose --env-file .env run --rm migrate
   ```

3. Запустить приложение:
   ```bash
   go run cmd/http/main.go -config config/local.yaml
   ```
   
## Конфигурация

Файл конфигурации (YAML) указывается флагом `-config`. 
Переменные окружения с префиксом `IR_` переопределяют значения из файла.

```yaml
http:
  port: 8400
  cors_allow_origins: ["http://localhost:3000"]
  cors_max_age: 300

postgres:
  uri: "postgres://user:password@localhost:5432/itsreg"

logging:
  level: debug

jwt:
  secret: "your-secret-key"
  access_ttl: 24h

rate_limiter:
  capacity: 10
  rate: 25.0

proxy:
  url: ""
```

Пример с переменными окружения:
```bash
IR_POSTGRES_URI="postgres://..." \
IR_JWT_SECRET="key" \
go run cmd/http/main.go -config config/local.yaml
```

## API

REST API на `/api/v3`. Спецификация OpenAPI в `api/openapi/itsreg.swagger.yaml`.

Основные эндпоинты:
- `/bots` - управление ботами
- `/scripts` - управление сценариями  
- `/runs` - управление запусками

## Разработка

### Генерация кода

```bash
make openapi-gen
```

Требует `oapi-codegen`.

### Тесты

```bash
go test ./...
```

Интеграционные тесты используют `testcontainers`, поэтому требуют Docker.
Для запуска только модульных тестов:
```bash
go test -short ./...
```

### Миграции БД

Запускаются автоматически при старте. Для добавления:
```bash
migrate create -ext sql -dir migrations -seq migration_name
```

Требуют установленного [golang-migrate](https://github.com/golang-migrate/migrate).

## Архитектура

Реализована на основе CQRS и event sourcing:
- **Commands** - изменение состояния
- **Queries** - чтение состояния  
- **Event handlers** - реакция на события
- **Ports** - интерфейсы инфраструктуры

Логирование через `slog`. Уровень в конфигурации.
