### Телеграм бот, аналог бота "Всратослав". Добавляет случайную подпись к изображению по специальной фразе, или с определенным шансом.

## Deployment

### docker compose

**Build** application

```shell
docker compose build
```

**Run** application

```shell
docker compose up -d
```

### All options are loaded from **[.env](.env)**

```dotenv
BOT_DEBUG=false
BOT_TOKEN=YOUR_TOKEN
BOT_ADMIN_LIST=YOUR_ID

POSTGRES_HOST=postgres
POSTGRES_PORT=5432
POSTGRES_USER=postgres
POSTGRES_PASSWORD=postgres
POSTGRES_DB=markoslav
```