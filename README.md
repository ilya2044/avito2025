Makefile:
make build - компиляция сервиса в bin/pr-reviewer
make up - поднятие контейнеров с пересборкой образов
make down - остановка и удаление контейнеров
make clean - полностью очищает окружение
make migrate - применяет миграции к БД
make run - поднимает проект и автоматически выполняет миграцию, после чего выводит логи приложения

Для запуска проекта:
1. make build
2. make up
3. make run

Проблемы и решения:
по ходу выполнения задания столкнулся с такой проблемой:
Изначально в базе можно было создавать пользователей с одинаковым user_id в разных командах. Это приводило к багу: при создании PR по user_id сервер не понимал, к какой команде принадлежит пользователь, и могли возникать некорректные назначения ревьюеров. Также была проблема с добавлением пользователей в команды
Решение:
Добавлено ограничение на уникальность user_id в таблице users
Теперь пользователь может состоять только в одной команде
Если попытаться создать пользователя с существующим user_id или добавить его в другую команду, сервер вернёт ошибку, а действие не выполнится
Добавлена возможность добалять/удалять пользователей из команд

Примеры запросов:

1. Создание команды:
curl -X POST http://localhost:8080/team/add -H "Content-Type: application/json" -d '{
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

2. Добавить нового пользователя:
curl -X POST http://localhost:8080/team/addUser -H "Content-Type: application/json" -d '{
  "team_name": "backend",
  "user": {
    "user_id": "u10",
    "username": "Worker10",
    "is_active": true
  }
}'

2.1 Удалить пользователя:
curl -X POST http://localhost:8080/team/removeUser \
-H "Content-Type: application/json" \
-d '{
  "team_name": "backend",
  "user_id": "u10"
}'

3. Получить команду:
curl http://localhost:8080/team/get?team_name=backend

4. Изменить активность:
curl -X POST http://localhost:8080/users/setIsActive -H "Content-Type: application/json" -d '{
  "user_id":"u3",
  "is_active":true
}'

5. Создать PR:
curl -X POST http://localhost:8080/pullRequest/create -H "Content-Type: application/json" -d '{
  "pull_request_id":"pr1",
  "pull_request_name":"New pr",
  "author_id":"u3"
}'

6. Получить пул-реквесты, в которых worker - reviwer 
curl http://localhost:8080/users/getReview?user_id=u1

7. Переназначить ревьювера:
curl -X POST http://localhost:8080/pullRequest/reassign -H "Content-Type: application/json" -d '{
  "pull_request_id":"pr1",
  "old_user_id":"u2"
}'

8. Merge:
curl -X POST http://localhost:8080/pullRequest/merge -H "Content-Type: application/json" -d '{
  "pull_request_id":"pr1"
}'

