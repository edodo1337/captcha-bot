# Сборка и запуск

## Локальный запуск
Создаем виртуальное окружение python3.10 любым  удобным способом, например:
`$ python3.10 -m venv venv`

Устанавливаем зависимости:
`$ pip install -r requirements.txt`

Получаем и экспортируем переменную окружения с Telegram токеном:
`export TG_API_TOKEN=<полученный токен>`

Запускаем скрипт:
`python3 main.py`


## Локальный запуск (Docker)
Собираем образ
`$ sudo docker build -t captcha-bot .`

Получаем Telegram токен и создаем .env файл:
> TG_API_TOKEN=<полученный токен>

Запускаем контейнер:
`$ sudo docker run --env-file .env captcha-bot`