# avito2025

## Запуск проекта

```bash
docker-compose up -d
```
## Проблемы и решения
### По ходу выполнения задания столкнулся с проблемой:
Изначально в базе можно было создавать пользователей с одинаковым user_id в разных командах. Это приводило к багу: при создании PR по user_id сервер не понимал, к какой команде принадлежит пользователь, и могли возникать некорректные назначения ревьюеров. Также была проблема с добавлением пользователей в команды
### Решение:
- Добавлено ограничение на уникальность user_id в таблице users
- Теперь пользователь может состоять только в одной команде
- Если попытаться создать пользователя с существующим user_id или добавить его в другую команду, сервер вернёт ошибку, а действие не выполнится
- Добавлена возможность добавлять и удалять пользователей из команд

## Примеры запросов:
### 1. Создание команды
```bash
curl -X POST http://localhost:8080/team/add \
-H "Content-Type: application/json" \
-d '{
  "team_name":"backend",
  "members":[
    {"user_id":"u1","username":"Worker1","is_active":true},
    {"user_id":"u2","username":"Worker2","is_active":true},
    {"user_id":"u3","username":"Worker3","is_active":true},
    {"user_id":"u4","username":"Worker4","is_active":true},
    {"user_id":"u5","username":"Worker5","is_active":true},
    {"user_id":"u6","username":"Worker6","is_active":false}
  ]
}'
```
### 2. Добавить нового пользователя
```bash
curl -X POST http://localhost:8080/team/addUser \
-H "Content-Type: application/json" \
-d '{
  "team_name": "backend",
  "user": {
    "user_id": "u10",
    "username": "Worker10",
    "is_active": true
  }
}'

```
### 2.1. Удалить пользователя
```bash
curl -X POST http://localhost:8080/team/removeUser \
-H "Content-Type: application/json" \
-d '{
  "team_name": "backend",
  "user_id": "u10"
}'

```
### 3. Получить команду
```bash
curl http://localhost:8080/team/get?team_name=backend

```
### 4. Изменить активность пользователя
```bash
curl -X POST http://localhost:8080/users/setIsActive \
-H "Content-Type: application/json" \
-d '{
  "user_id":"u3",
  "is_active":true
}'

```
### 5. Создать PR
```bash
curl -X POST http://localhost:8080/pullRequest/create \
-H "Content-Type: application/json" \
-d '{
  "pull_request_id":"pr1",
  "pull_request_name":"New pr",
  "author_id":"u3"
}'

```
### 6. Получить PR, где пользователь является ревьюером
```bash
curl http://localhost:8080/users/getReview?user_id=u1

```
### 7. Переназначит ревьювера
```bash
curl -X POST http://localhost:8080/pullRequest/reassign \
-H "Content-Type: application/json" \
-d '{
  "pull_request_id":"pr1",
  "old_user_id":"u2"
}'

```
### 8. Merge
```bash
curl -X POST http://localhost:8080/pullRequest/merge \
-H "Content-Type: application/json" \
-d '{
  "pull_request_id":"pr1"
}'

```
