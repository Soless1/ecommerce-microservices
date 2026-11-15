# E-commerce Microservices

## Оглавление

- [О проекте](#о-проекте)
- [Стек](#стек)
- [Сервисы](#сервисы)
- [Проверки в CI](#проверки-в-ci)
- [Установка зависимостей](#установка-зависимостей)
- [Локальный запуск](#локальный-запуск)

## О проекте

Бэкенд интернет-магазина на Go. Сервисы управляют каталогом и остатками, хранят корзины, оформляют заказы и доставляют уведомления о смене статуса. Внешние клиенты могут использовать REST API или gRPC.

## Стек

| Область | Технологии |
| --- | --- |
| Язык и API | Go, Go Workspaces, gRPC, Protocol Buffers, REST API через gRPC-Gateway, EasyP |
| Хранение | PostgreSQL, pgx, sqlc, Goose |
| Кэш | Redis, go-redis, cache-aside, TTL, LRU |
| События и уведомления | Apache Kafka, Transactional Outbox, kafka-go, confluent-kafka-go, Telegram Bot API |
| Логи | Zap |
| Сборка и проверки | Docker, Docker Compose, Go Task, GitHub Actions, GoMock, Testify |

## Сервисы

- **`cart`** хранит корзины в PostgreSQL и обращается к `loms` по gRPC за данными о товарах и при оформлении заказа.
- **`loms`** управляет товарами, остатками и заказами. При оформлении заказа он резервирует товар, при отмене освобождает резерв. Карточки товаров кэшируются в Redis на минуту. При промахе или ошибке Redis сервис читает их из PostgreSQL.
- **`notifications`** читает события из Kafka и доставляет уведомления в Telegram или через HTTP callback. `loms` записывает событие в PostgreSQL в одной транзакции с изменением заказа. Затем фоновый обработчик Outbox публикует его в Kafka.

## Проверки в CI

[GitHub Actions](.github/workflows/ci.yaml) запускает генерацию кода, модульные тесты, линтер и интеграционные тесты при открытии или обновлении pull request и при пуше в `main`. Интеграционные тесты поднимают сервисы вместе с PostgreSQL, Kafka и Redis и проверяют доставку уведомлений через HTTP callback.

## Установка зависимостей

- **Git, Docker и Docker Compose v2** для получения кода и запуска сервисов.
- **Go и Go Task** для генерации кода. Версия Go закреплена в [`go.work`](go.work). Чтобы установить Task, выполни `go install github.com/go-task/task/v3/cmd/task@v3.45.4` и добавь каталог `$(go env GOPATH)/bin` в `PATH`.
- **`jq` и `protoc`** для генерации. На Ubuntu выполни `sudo apt-get update && sudo apt-get install -y jq protobuf-compiler`. На Windows можно использовать WSL с включённой интеграцией Docker Desktop.

PostgreSQL, Kafka и Redis запускаются в Docker. Порты и параметры подключения заданы в [`docker-compose.yml`](docker-compose.yml).

Для доставки уведомлений в Telegram создай бота и заполни `.env` в корне проекта:

```dotenv
TELEGRAM_TOKEN=токен_бота
TELEGRAM_CHAT_ID=идентификатор_чата
```

`TELEGRAM_CHAT_ID` должен быть числом. Файл `.env` исключён из Git.

## Локальный запуск

Из корня репозитория сгенерируй код и запусти бэкенд:

```bash
task generate
task up
```

Для запуска без Telegram используй `task up -- cart cart-2` вместо `task up`. Compose запустит обе реплики `loms`, PostgreSQL, Kafka и Redis как зависимости корзины.

Если нужен веб-интерфейс, установи Node.js с npm и выполни `task frontend` из корня проекта в отдельном терминале. Интерфейс откроется по адресу [http://localhost:5173](http://localhost:5173).

Чтобы остановить бэкенд, выполни `docker compose down`. Данные PostgreSQL сохранятся в томе `postgres-data`.
