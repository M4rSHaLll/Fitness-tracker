# Fitness Tracker

API для хранения пользователей, упражнений, тренировок, подходов и расчета базовой статистики.

Текущая версия использует in-memory хранилище. Оно подходит для локальной разработки и тестов, но данные теряются после перезапуска приложения.

## Конфигурация

Приложение читает конфигурацию из переменных окружения. Файл `.env.example` содержит пример значений, но `.env` не загружается приложением автоматически.

Текущие переменные:

- `APP_ENV`: окружение приложения. По умолчанию `local`.
- `APP_HOST`: host для HTTP-сервера. По умолчанию пустая строка, сервер слушает все интерфейсы.
- `APP_PORT`: порт HTTP-сервера. По умолчанию `8080`.
- `STORAGE_DRIVER`: тип хранилища. Поддерживаются `memory` и `postgres`.

Переменные PostgreSQL:

- `DATABASE_URL`: полная строка подключения к БД.
- `DB_HOST`: host БД.
- `DB_PORT`: порт БД.
- `DB_USER`: пользователь БД.
- `DB_PASSWORD`: пароль БД.
- `DB_NAME`: имя БД.
- `DB_SSLMODE`: SSL-режим. По умолчанию `disable`.

Пример для PowerShell:

```powershell
$env:APP_PORT = "8081"
$env:STORAGE_DRIVER = "memory"
go run ./cmd/telegram
```

## Запуск

```bash
go run ./cmd/telegram
```

Сервер запускается на `http://localhost:8080`.

## PostgreSQL

Для локального запуска PostgreSQL:

```bash
docker compose up -d postgres
```

Перед запуском приложения с `STORAGE_DRIVER=postgres` нужно применить миграции из папки `migrations`.

Пример с `golang-migrate`:

```bash
migrate -path migrations -database "postgres://fitness:fitness@localhost:5432/fitness_tracker?sslmode=disable" up
```

Запуск приложения с PostgreSQL в PowerShell:

```powershell
$env:STORAGE_DRIVER = "postgres"
$env:DATABASE_URL = "postgres://fitness:fitness@localhost:5432/fitness_tracker?sslmode=disable"
go run ./cmd/telegram
```

Для возврата к тестовому in-memory хранилищу:

```powershell
$env:STORAGE_DRIVER = "memory"
go run ./cmd/telegram
```

## Модель данных

### User

```json
{
  "id": 1,
  "telegram_id": 123456789,
  "username": "art",
  "created_at": "2026-06-08T12:00:00Z",
  "updated_at": "2026-06-08T12:00:00Z",
  "weight": 80,
  "height": 180,
  "age": 30
}
```

Правила:

- `telegram_id` должен быть уникальным.
- `username` обязателен.
- `weight`: от `1` до `500`.
- `height`: от `1` до `300`.
- `age`: от `1` до `150`.

### Exercise

```json
{
  "id": 1,
  "name": "Bench press",
  "created_at": "2026-06-08T12:00:00Z",
  "updated_at": "2026-06-08T12:00:00Z"
}
```

Правила:

- `name` обязателен.
- `name` должен быть уникальным без учета регистра.

### Workout

```json
{
  "id": 1,
  "user_id": 1,
  "description": "Push workout",
  "created_at": "2026-06-08T12:00:00Z"
}
```

Правила:

- `user_id` должен ссылаться на существующего пользователя.
- `description` обязателен.

### Set

```json
{
  "id": 1,
  "exercise_id": 1,
  "workout_id": 1,
  "reps": 5,
  "weight": 100,
  "rpe": 8.5,
  "created_at": "2026-06-08T12:00:00Z"
}
```

Правила:

- `workout_id` должен ссылаться на существующую тренировку.
- `exercise_id` должен ссылаться на существующее упражнение.
- `reps`: от `0` до `1000`.
- `weight`: от `1` до `10000`.
- `rpe`: от `6` до `10` с шагом `0.5`.

## API

Все ответы возвращаются в JSON.

### Health

#### Get health

`GET /health`

Success: `200 OK`

Response:

```json
{
  "status": "ok",
  "storage": "postgres"
}
```

### Users

#### Create user

`POST /users`

Request:

```json
{
  "telegram_id": 123456789,
  "username": "art"
}
```

Success: `201 Created`

Response: `User`

#### Get user

`GET /users/{id}`

Success: `200 OK`

Response: `User`

#### Update user profile

`PATCH /users/{id}`

Request:

```json
{
  "weight": 80,
  "height": 180,
  "age": 30
}
```

Success: `200 OK`

Response:

```json
{
  "message": "user updated successfully"
}
```

### Exercises

#### Create exercise

`POST /exercises`

Request:

```json
{
  "name": "Bench press"
}
```

Success: `201 Created`

Response: `Exercise`

#### Get exercises

`GET /exercises`

Success: `200 OK`

Response:

```json
[
  {
    "id": 1,
    "name": "Bench press",
    "created_at": "2026-06-08T12:00:00Z",
    "updated_at": "2026-06-08T12:00:00Z"
  }
]
```

#### Get exercise

`GET /exercises/{id}`

Success: `200 OK`

Response: `Exercise`

#### Delete exercise

`DELETE /exercises/{id}`

Success: `200 OK`

Response:

```json
{
  "message": "exercise deleted successfully"
}
```

### Workouts

#### Create workout

`POST /workouts`

Request:

```json
{
  "user_id": 1,
  "description": "Push workout"
}
```

Success: `201 Created`

Response: `Workout`

#### Get user workouts

`GET /users/{id}/workouts`

Success: `200 OK`

Response:

```json
[
  {
    "id": 1,
    "user_id": 1,
    "description": "Push workout",
    "created_at": "2026-06-08T12:00:00Z"
  }
]
```

Если у пользователя нет тренировок, возвращается `[]`.

#### Delete workout

`DELETE /workouts/{id}`

Success: `200 OK`

Response:

```json
{
  "message": "workout deleted successfully"
}
```

### Sets

#### Create set

`POST /sets`

Request:

```json
{
  "workout_id": 1,
  "exercise_id": 1,
  "weight": 100,
  "reps": 5,
  "rpe": 8.5
}
```

Success: `201 Created`

Response: `Set`

#### Get workout sets

`GET /workouts/{id}/sets`

Success: `200 OK`

Response:

```json
[
  {
    "id": 1,
    "exercise_id": 1,
    "workout_id": 1,
    "reps": 5,
    "weight": 100,
    "rpe": 8.5,
    "created_at": "2026-06-08T12:00:00Z"
  }
]
```

Если у тренировки нет подходов, возвращается `[]`.

#### Delete set

`DELETE /sets/{id}`

Success: `200 OK`

Response:

```json
{
  "message": "set deleted successfully"
}
```

### Stats

#### Get user stats

`GET /users/{id}/stats`

Success: `200 OK`

Response:

```json
{
  "total_workouts": 1,
  "total_volume": 500,
  "average_rpe": 8.5
}
```

## Ошибки

Общий формат:

```json
{
  "error": "Error message",
  "code": "machine_readable_code"
}
```

Коды:

- `400 Bad Request`: невалидный JSON, неизвестное поле, невалидные значения.
- `404 Not Found`: пользователь, упражнение, тренировка или подход не найдены.
- `409 Conflict`: ресурс уже существует.
- `500 Internal Server Error`: внутренняя ошибка приложения.

## Проверка

```bash
go test ./...
```

Через Makefile:

```bash
make test
```

PostgreSQL integration-тесты запускаются только если задан `POSTGRES_TEST_DATABASE_URL`:

```bash
POSTGRES_TEST_DATABASE_URL="postgres://fitness:fitness@localhost:5432/fitness_tracker?sslmode=disable" go test ./internal/repository/postgres
```

## Makefile

```bash
make run
make run-postgres
make test
make build
make db-up
make db-down
make migrate-up
make migrate-down
```
