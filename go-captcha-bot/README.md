# Сборка и запуск

## Локальный запуск
Получаем и помещаем API TOKEN Telegram в файл config.yaml в корневой директории проекта:
```yaml
bot:
  token: <API_TOKEN>
  ban_timeout: <BAN TIMEOUT, default=120>
```

Запускаем скрипт:
`go run cmd/main.go `


## Локальный запуск (Docker)
Собираем образ
`$ sudo docker build -t go-captcha-bot .`

Запускаем контейнер:
`$ sudo docker run --env-file .env captcha-bot`

### Переменные конфигурации
> ban_timeout - (секунды) таймаут на ответ (по умолчанию 120)
> token - API токен бота из @BotFather