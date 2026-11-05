# Slot Booking Service

Веб-сервис для управления расписанием и бронирования временных слотов.

## Быстрый запуск через Docker Compose

### Требования

- Docker;
- Docker Compose.

### Запуск

1. Создайте файл с переменными окружения:

   ```bash
   cp .env.example .env
   ```

2. Соберите и запустите сервисы:

   ```bash
   docker compose up --build -d
   ```

### Полезные команды

```bash
# Состояние контейнеров
docker compose ps

# Логи приложения
docker compose logs -f app

# Логи всех сервисов
docker compose logs -f

# Остановка сервисов без удаления данных
docker compose down
```

Приложение будет доступно по адресу `http://localhost:8080`.
