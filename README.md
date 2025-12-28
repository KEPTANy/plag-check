# Система проверки на плагиат

Микросервисная система для регистрации пользователей (студентов и преподавателей), загрузки файлов студентами и анализа файлов преподавателями на предмет плагиата.

## Архитектура системы

Система построена на микросервисной архитектуре и состоит из следующих компонентов:

### Микросервисы

1. **user-service** (порт 8081)
   - Регистрация пользователей (студентов и преподавателей)
   - Аутентификация и выдача JWT токенов

2. **file-storage-service** (порт 8082)
   - Загрузка файлов студентами
   - Скачивание файлов студентами и преподавателями
   - Просмотр списка файлов пользователя
   - Поиск файлов по хешу (для преподавателей)

3. **analysis-service** (порт 8083)
   - Проверка на плагиат (поиск файлов с одинаковым хешем)
   - Генерация облака слов из содержимого файла через QuickChart.io API

4. **gateway-api** (порт 8080)
   - API Gateway для маршрутизации запросов к микросервисам
   - Обработка ошибок при недоступности микросервисов
   - Единая точка входа для всех клиентов

### Инфраструктура

- **PostgreSQL** - база данных для хранения пользователей и метаданных файлов
- **Docker** - контейнеризация всех сервисов
- **Docker Compose** - оркестрация и развертывание системы

## Технические сценарии взаимодействия

### Сценарий 1: Регистрация и аутентификация

1. **Регистрация пользователя**
   ```
   Клиент -> Gateway (8080) -> POST /auth/register
   Gateway -> User Service (8081) -> POST /auth/register
   User Service -> PostgreSQL: создание пользователя
   User Service -> Gateway -> Клиент: 201 Created
   ```

2. **Вход в систему**
   ```
   Клиент -> Gateway (8080) -> POST /auth/login
   Gateway -> User Service (8081) -> POST /auth/login
   User Service -> PostgreSQL: проверка учетных данных
   User Service -> Генерация JWT токена
   User Service -> Gateway -> Клиент: { "token": "..." }
   ```

### Сценарий 2: Загрузка файла студентом

1. **Загрузка файла**
   ```
   Клиент -> Gateway (8080) -> POST /files/upload
   Headers: Authorization: Bearer <student_token>
   Gateway -> File Storage Service (8082) -> POST /files/upload
   File Storage Service:
     - Проверка JWT токена
     - Проверка роли
     - Вычисление SHA256 хеша файла
     - Сохранение файла в хранилище
     - Сохранение метаданных в PostgreSQL
   File Storage Service -> Gateway -> Клиент: { "file_info": {...} }
   ```

### Сценарий 3: Проверка на плагиат преподавателем

1. **Получение списка плагиатов**
   ```
   Клиент -> Gateway (8080) -> GET /analysis/plagiarism
   Headers: Authorization: Bearer <teacher_token>
   Gateway -> Analysis Service (8083) -> GET /analysis/plagiarism
   Analysis Service:
     - Проверка JWT токена
     - Проверка роли
     - Запрос к PostgreSQL: поиск хешей с COUNT > 1
     - Для каждого хеша: получение списка файлов
   Analysis Service -> Gateway -> Клиент: { "plagiarism_results": [...] }
   ```

### Сценарий 4: Генерация облака слов

1. **Генерация облака слов из файла**
   ```
   Клиент -> Gateway (8080) -> GET /analysis/wordcloud/{file_id}
   Headers: Authorization: Bearer <teacher_token>
   Gateway -> Analysis Service (8083) -> GET /analysis/wordcloud/{file_id}
   Analysis Service:
     - Проверка JWT токена и роли
     - Получение метаданных файла из PostgreSQL
     - Чтение содержимого файла из хранилища
     - Отправка текста в QuickChart.io API
     - Получение изображения PNG
   Analysis Service -> Gateway -> Клиент: PNG изображение
   ```

### Сценарий 5: Обработка ошибок при недоступности сервиса

1. **Недоступность микросервиса**
   ```
   Клиент -> Gateway (8080) -> GET /files/user/{userid}
   Gateway -> File Storage Service (8082) -> [Сервис недоступен]
   Gateway -> Клиент: 503 Service Unavailable
   Response: { "error": "service unavailable" }
   ```

## API Endpoints

### Authentication

- `POST /auth/register` - Регистрация пользователя
  - Body: `{ "username": "string", "password": "string", "role": "student|teacher" }`
  
- `POST /auth/login` - Вход в систему
  - Body: `{ "username": "string", "password": "string", "duration_min": number }`
  - Response: `{ "token": "string" }`

### File Storage (требует JWT токен)

- `POST /files/upload` - Загрузка файла (только для студентов)
  - Headers: `Authorization: Bearer <token>`
  - Body: `multipart/form-data` с полем `file`
  
- `GET /files/download/{id}` - Скачивание файла (студенты и преподаватели)
  - Headers: `Authorization: Bearer <token>`
  
- `GET /files/user/{userid}` - Список файлов пользователя (студенты и преподаватели)
  - Headers: `Authorization: Bearer <token>`
  
- `GET /files/hash/{hash}` - Список файлов с указанным хешем (только для преподавателей)
  - Headers: `Authorization: Bearer <token>`

### Analysis (требует JWT токен, только для преподавателей)

- `GET /analysis/plagiarism` - Проверка на плагиат
  - Headers: `Authorization: Bearer <token>`
  - Response: `{ "plagiarism_results": [{ "hash": "...", "count": 2, "files": [...] }] }`
  
- `GET /analysis/wordcloud/{id}` - Генерация облака слов из файла
  - Headers: `Authorization: Bearer <token>`
  - Response: PNG изображение

### Health Checks

- `GET /health` - Проверка работоспособности сервиса

## Развертывание

### Требования

- Docker и Docker Compose
- Переменные окружения (создайте `.env` файл):

### Запуск системы

```bash
docker compose up
```

Система будет доступна по адресу `http://localhost:8080` (Gateway API).

### Остановка системы

```bash
docker compose down
```

Для удаления всех данных (включая базу данных):

```bash
docker compose down -v
```

## Структура проекта

```
user-service/          # Сервис управления пользователями
file-storage-service/   # Сервис хранения файлов
analysis-service/       # Сервис анализа (плагиат, облако слов)
gateway-api/           # API Gateway
shared/                # Общий код (JWT, middleware)
docker-compose.yml     # Конфигурация Docker Compose
README.md             # Документация
```

## Алгоритм проверки на плагиат

Система использует простой, но эффективный алгоритм:
1. При загрузке файла вычисляется SHA256 хеш содержимого
2. Файлы с одинаковым хешем считаются идентичными (плагиат)
3. Преподаватель может запросить список всех групп файлов с одинаковыми хешами
4. Для каждой группы возвращается список всех файлов с этим хешем

## Тестирование

Для тестирования API используйте Postman коллекцию (см. `postman_collection.json`)
