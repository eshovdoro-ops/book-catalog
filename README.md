# book-catalog

REST API сервер на Go для управления каталогом книг.

## Функционал

- Аутентификация с JWT
- CRUD операции для книг
- Хранение данных в JSON файлах

## Запуск

1. Установите Go и GoLand.
2. Скачайте зависимости: `go mod tidy`
3. Запустите сервер: `cd app && go run main.go`
4. Сервер слушает на :8080

## API

- POST /login: {"username": "admin", "password": "password"} -> JWT token
- GET /books: список книг (требует Bearer token)
- POST /books: добавить книгу
- PUT /books/{id}: обновить книгу
- DELETE /books/{id}: удалить книгу

## Тестирование

Используйте Postman или curl для тестирования API.