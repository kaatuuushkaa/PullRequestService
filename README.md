# Сервис назначения ревьюеров для Pull Request’ов
Внутри команды требуется единый микросервис, который автоматически назначает ревьюеров на Pull Request’ы (PR), а также позволяет управлять командами и участниками. Взаимодействие происходит исключительно через HTTP API.

## Запуск
```bash
docker-compose up
```

Сервис доступен на порту 8080.

## Стек проекта
- Язык: Go
- Фреймворк для HTTP: Echo
- База данных: PostgreSQL
- Миграции: через `golang-migrate`(автоматически подтягиваются при запуске проекта)
- Документация API: OpenApi + `oapi-codegen`

## Архитектура
```
.
├── cmd/                       # Точка входа приложения
├── domain/                    # Доменные сущности
├── internal/
│   ├── app/                   # Инициализация приложения
│   ├── db/                    # Подключение к PostgreSQL
│   ├── handlers/              # HTTP-слой сервиса
│   ├── pullRequestService/    # Бизнес-логика Pull Requests
│   ├── statsService/          # Бизнес-логика статистики
│   └── teamService/           # Бизнес-логика команд
│   └── userService/           # Бизнес-логика пользователей
│   └── web/                   # Код, сгенерированный OpenAPI генератором
│   │   ├── pullRequests/
│   │   ├── stats/
│   │   ├── teams/
│   │   ├── users/
├── loadtest/                  # Нагрузочный тест на Go
├── migrations/                # SQL миграции для PostgreSQL
└── openapi/                   # OpenAPI спецификация (openapi.yaml)
```
## Примечание
В качестве `user_id` используется тип `string` (например: `u1`, `u2`), что полностью соответствует примеру в `openapi.yml`. При необходимости система допускает последующий переход на `int` или `UUID`.

Для повышения производительности и ускорения выборок, связанных с активностью пользователей, статусами PR и операциями по ревью, были созданы следующие индексы: `idx_users_team_active`, `idx_pull_requests_status`, `idx_pr_reviewers_reviewer`, `idx_pr_reviewers_pr`.

---
# Бонусные фичи
## Статистика
`GET /stats`
Возвращает среднюю статистику по pull request-аь и назначенным ревьюверам.
1. `assignments_by_pr` - сколько ревьюверов назначено на кажждый PR.
2. `assignments_by_user` - сколько PR назначено каждому пользователю.
3. `merged_prs_count` - общее количество замерженных PR.
4. `open_prs_count` - общее количество открытых PR.
<img width="1245" height="769" alt="image" src="https://github.com/user-attachments/assets/051fc0b5-6c77-4995-a016-524338fd99e3" />


## Linter 
В проекте настроен golangci-lint для автоматической проверки кода на ошибки и нарушения стиля.
Конфиг находится в .golangci.yaml и включает в себя:
- errcheck - чтобы не пропустить ошибки
- gosimple - выявить лишний код
- govet - выявить ошибки с затенением
- ineffassign - найти неээфективные присваивания переменных
- staticcheck - статический анализ кода
- unused - найти неиспользуемые переменные, фугкции, импорты, чтобы поддерживать чистоту кода
- misspell - проверка опечатки в строках и комментариях
- gofmt - следит за стилем

## LoadTest
Для нагрузочного тестирования добавлен loadtest/test.go.
Запускаем после `docker-compose up` с помощью `make load-test`
- один раз создается команда `payments` с 5 активными пользователями (`POST /team/add`)
- в несколько потоков создаются PR (`POST /pullRequest/create`)
- периодически вызываются `GET /users/getReview` и `GET /stats`
<img width="713" height="466" alt="image" src="https://github.com/user-attachments/assets/b3d425ac-0645-4029-a750-afbee2f9b23e" />

## Mассовая деактивация пользователей команды
Операция деактивирует указанных пользователей и корректно обновляет открые PR
Примеры запросов:
1. Успешная массовая деактивация
Request
```
curl -X POST http://localhost:8080/users/deactivate \
  -H "Content-Type: application/json" \
  -d '{
    "team_name": "backend",
    "user_ids": ["u2", "u3", "u5"]
  }'
```

Response (200)
```
{
  "team_name": "backend",
  "deactivated_count": 3,
  "affected_pr_count": 4,
  "reassigned_reviewers_count": 3
}
```
2. Частичная деактивация (например, часть id не существует)
Request
```
curl -X POST http://localhost:8080/users/deactivate \
  -H "Content-Type: application/json" \
  -d '{
    "team_name": "backend",
    "user_ids": ["u1", "u999", "u2"]
  }'
  ```
Response(200)
```
{
  "team_name": "backend",
  "deactivated_count": 2,
  "affected_pr_count": 1,
  "reassigned_reviewers_count": 1
}
```
3. Команда не найдена
Request
```
curl -X POST http://localhost:8080/users/deactivate \
  -H "Content-Type: application/json" \
  -d '{
    "team_name": "unknown_team",
    "user_ids": ["u1", "u2"]
  }'
  ```
Response(404)
```
{
  "error": {
    "code": "NOT_FOUND",
    "message": "team not found"
  }
}
```
4. Пустой список пользователей
Request
```
curl -X POST http://localhost:8080/users/deactivate \
  -H "Content-Type: application/json" \
  -d '{
    "team_name": "backend",
    "user_ids": []
  }'
  ```
Response(400)
```
{
  "error": {
    "code": "BAD_REQUEST",
    "message": "team_name and user_ids are required"
  }
}
```
5. Успешно, но никто не был ревьюером
Request
```
curl -X POST http://localhost:8080/users/deactivate \
  -H "Content-Type: application/json" \
  -d '{
    "team_name": "backend",
    "user_ids": ["u7"]
  }'
  ```
Response(200)
```
{
  "team_name": "backend",
  "deactivated_count": 1,
  "affected_pr_count": 0,
  "reassigned_reviewers_count": 0
}
```
