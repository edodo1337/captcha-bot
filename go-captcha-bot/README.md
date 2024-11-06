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
> token - API токен Telegram.
> ban_timeout - Таймаут для ответа (сек).
> captcha_message - Сообщение капчи.
> user_state_ttl - TTL состояния пользователя (сек).
> cleanup_interval - Интервал очистки (сек).
> vote_kick_timeout - Таймаут голосования (сек).
> min_kick_votes_for - Минимум голосов "за" кик.
> gemini_api_tokens - Токены API Google Gemini.
> yandex_api_tokens - Токены API Yandex.
> yandex_catalog_ids - Каталоги Yandex.
> gpt_client - Клиент GPT.
> prompt_wrap - Шаблон запроса для LLM.
> admins - Администраторы (массив).