# Fitness Tracker

Backend MVP для учета силовых тренировок через REST API и Telegram-бота.

Основное хранилище проекта — PostgreSQL. In-memory реализация сохранена для быстрых локальных тестов.
HTTP API и Telegram-бот запускаются одним приложением.

## Возможности

- профили пользователей, упражнения, тренировки и подходы;
- статистика по количеству тренировок, объему и среднему RPE;
- RPE от `6` до `10` с шагом `0.5`;
- Telegram-сценарий без ручного ввода внутренних ID;
- PostgreSQL через `pgx`, SQL-миграции и in-memory адаптер;
- Docker Compose для приложения, миграций и PostgreSQL;
- graceful shutdown и повторное подключение Telegram polling;
- Bearer-защита REST API и отдельный токен внутренних Telegram-маршрутов;
- unit-, HTTP- и PostgreSQL integration-тесты;
- CI с race detector и проверкой Docker-сборки.

## Архитектура

```mermaid
flowchart LR
    Client["REST client"] --> Router["chi router"]
    Telegram["Telegram user"] --> Bot["Telegram bot polling"]
    Router --> Handler["HTTP handlers"]
    Handler --> Service["Services"]
    Bot --> Service
    Service --> Repository["Repository interfaces"]
    Repository --> PG["PostgreSQL repositories"]
    Repository --> Memory["In-memory repositories"]
```

Основной поток API: `router → handler → service → repository`.
Бот использует те же сервисы напрямую, поэтому бизнес-правила не дублируются.

## Модель доступа

- `GET /health` доступен без авторизации.
- Обычные REST-маршруты защищены `Authorization: Bearer <API_TOKEN>`.
- Маршруты `/telegram/...` защищены отдельным заголовком `X-Internal-Token`.
- Бот берет Telegram ID только из `update.Message.From.ID`.
- При добавлении подхода проверяется принадлежность тренировки текущему пользователю.

`API_TOKEN` — операторский ключ для локальной разработки и закрытого API. Это не полноценная пользовательская аутентификация для публичного web/mobile-клиента.

## Конфигурация

Приложение читает конфигурацию из переменных окружения и локального файла `.env`.
Системные переменные имеют приоритет над значениями из `.env`.
Файл `.env` игнорируется Git и не должен попадать в репозиторий.

Текущие переменные:

- `APP_ENV`: окружение приложения. По умолчанию `local`.
- `APP_HOST`: host для HTTP-сервера. По умолчанию пустая строка, сервер слушает все интерфейсы.
- `APP_PORT`: порт HTTP-сервера. По умолчанию `8080`.
- `STORAGE_DRIVER`: тип хранилища. Поддерживаются `memory` и `postgres`.
- `API_TOKEN`: Bearer-токен обычного REST API.
- `INTERNAL_API_TOKEN`: токен внутренних маршрутов `/telegram/...`.
- `TELEGRAM_BOT_TOKEN`: секретный токен BotFather.
- `TELEGRAM_MODE`: режим бота, сейчас поддерживается `polling`.

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

## Локальный запуск

```powershell
Copy-Item .env.example .env
go run ./cmd/telegram
```

Сервер запускается на `http://localhost:8080`.

## Запуск в Docker

Заполните в `.env` токены API и новый токен бота:

```env
API_TOKEN=long-random-api-token
INTERNAL_API_TOKEN=long-random-internal-token
TELEGRAM_BOT_TOKEN=token-from-botfather
```

Не публикуйте `.env` и не передавайте токен в сообщениях или логах.

Полный стек запускается одной командой:

```powershell
docker compose up --build
```

Compose последовательно:

- запускает PostgreSQL;
- применяет миграции;
- запускает API и Telegram polling.

Проверка состояния:

```powershell
docker compose ps
docker compose logs -f app
curl.exe http://localhost:8080/health
```

Остановка:

```powershell
docker compose down
```

Данные PostgreSQL сохраняются в volume `postgres-data`.
Для удаления данных используется отдельная команда `docker compose down -v`.

Внутри Docker приложение подключается к `postgres:5432`.
При запуске Go-приложения на Windows используется опубликованный порт `localhost:55432`.

Если Telegram API доступен только через локальный HTTP-прокси Docker Desktop, добавьте в `.env`:

```env
HTTPS_PROXY=http://host.docker.internal:7890
```

Порт должен совпадать с HTTP или mixed port прокси-программы.
Если Telegram временно недоступен, API продолжает работать, а бот повторяет подключение каждые 15 секунд.

## PostgreSQL отдельно

Для запуска только базы:

```powershell
docker compose up -d postgres
```

Пример ручного применения миграций:

```powershell
migrate -path migrations -database "postgres://fitness:fitness@localhost:55432/fitness_tracker?sslmode=disable" up
```

Запуск приложения на Windows с PostgreSQL:

```powershell
$env:STORAGE_DRIVER = "postgres"
$env:DATABASE_URL = "postgres://fitness:fitness@localhost:55432/fitness_tracker?sslmode=disable"
go run ./cmd/telegram
```

Для возврата к in-memory хранилищу:

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
Кроме `/health`, запросы должны содержать:

```http
Authorization: Bearer long-random-api-token
```

Пример PowerShell:

```powershell
curl.exe http://localhost:8080/exercises `
  -H "Authorization: Bearer $env:API_TOKEN"
```

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
- `401 Unauthorized`: отсутствует или неверен API/internal токен.
- `403 Forbidden`: попытка обратиться к тренировке другого пользователя.
- `404 Not Found`: пользователь, упражнение, тренировка или подход не найдены.
- `409 Conflict`: ресурс уже существует.
- `503 Service Unavailable`: обязательный токен не настроен.
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
POSTGRES_TEST_DATABASE_URL="postgres://fitness:fitness@localhost:55432/fitness_tracker?sslmode=disable" go test ./internal/repository/postgres
```

GitHub Actions выполняет форматирование, `go vet`, миграции, `go test -race ./...`, сборку приложения и Docker-образа.

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

## Внутренний Telegram API

Маршруты `/telegram/...` предназначены для доверенных внутренних интеграций и защищены заголовком `X-Internal-Token`.

Example:

```bash
export INTERNAL_API_TOKEN="local-secret"
```

Все запросы должны содержать:

```http
X-Internal-Token: local-secret
```

Supported endpoints:

- `POST /telegram/users`
- `POST /telegram/users/{telegram_id}/workouts`
- `GET /telegram/users/{telegram_id}/workouts`
- `POST /telegram/users/{telegram_id}/sets`
- `GET /telegram/users/{telegram_id}/workouts/{workout_id}/sets`
- `GET /telegram/users/{telegram_id}/stats`

API не принимает `user_id` от Telegram-клиента: backend определяет внутреннего пользователя по `telegram_id`.
Создание и чтение подходов также проверяют принадлежность тренировки.

Create or get Telegram user:

```bash
curl -X POST http://localhost:8080/telegram/users \
  -H "Content-Type: application/json" \
  -H "X-Internal-Token: local-secret" \
  -d '{"telegram_id":123456789,"username":"art"}'
```

Create workout for Telegram user:

```bash
curl -X POST http://localhost:8080/telegram/users/123456789/workouts \
  -H "Content-Type: application/json" \
  -H "X-Internal-Token: local-secret" \
  -d '{"username":"art","description":"Push workout"}'
```

Create set for Telegram user:

```bash
curl -X POST http://localhost:8080/telegram/users/123456789/sets \
  -H "Content-Type: application/json" \
  -H "X-Internal-Token: local-secret" \
  -d '{"workout_id":1,"exercise_id":1,"weight":100,"reps":5,"rpe":8.5}'
```

## Telegram-бот

Бот работает в одном процессе с API и использует long polling.

Основной пользовательский сценарий:

```text
/start
/profile_set 80 180 30
/exercise Bench press
/workout Push day
/addset
/stats
```

После `/addset` пользователь выбирает тренировку и упражнение кнопками, затем отправляет `вес повторения RPE`, например `100 5 8.5`.

Команды:

- `/start` — создать или найти профиль;
- `/profile` — показать профиль;
- `/profile_set <вес> <рост> <возраст>` — обновить профиль;
- `/exercises` — показать упражнения;
- `/exercise <название>` — добавить упражнение;
- `/workout <описание>` — создать тренировку;
- `/workouts` — показать последние тренировки;
- `/addset` — добавить подход с выбором кнопками;
- `/set <workout_id> <exercise_id> <weight> <reps> <rpe>` — совместимый ручной вариант;
- `/cancel` — отменить добавление подхода;
- `/stats` — показать статистику.

## Ограничения текущего MVP

- REST API использует общий операторский токен, а не учетные записи клиентов.
- Telegram работает только через polling; webhook пока не реализован.
- Каталог упражнений общий для всех пользователей.
- Для списков пока нет пагинации.
- Метрики и распределенная трассировка не подключены.
