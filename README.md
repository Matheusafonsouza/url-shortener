# Golang URL Shortener

A URL shortening service built with Go, Gin, and PostgreSQL.

## Features

- JSON API + web form UI
- Redirect short codes to original URLs
- Custom alias support (lowercase only)
- Link expiration with default TTL of 30 days (`720h`)
- Click count tracking

## Requirements

- Go 1.22+
- PostgreSQL 14+

## Setup

1. Copy `.env.example` to `.env` and adjust values.
2. Ensure PostgreSQL database exists and `DATABASE_URL` points to it.
3. The app runs versioned migration files automatically on startup using `golang-migrate`.

4. Install dependencies:

```bash
go mod tidy
```

5. Run the server:

```bash
go run ./cmd/server
```

Server starts at `http://localhost:8080` by default.

## Run with Docker Compose

1. Copy docker environment template:

```bash
cp .env.docker.example .env
```

2. Start app + postgres:

```bash
docker compose up --build
```

3. Open:

- Web UI: `http://localhost:8080`
- Health: `http://localhost:8080/health`

### Notes

- `DATABASE_URL` is injected automatically by `docker-compose.yml`.
- Database schema is migrated by the app on startup using migration files in `internal/storage/postgres/migrations`.
- If you need a clean DB re-init, run:

```bash
docker compose down -v
docker compose up --build
```

## API Endpoints

- `POST /api/v1/shorten`
  - Body:
    ```json
    {
      "url": "https://example.com/page",
      "alias": "my-page",
      "ttl_hours": 24
    }
    ```
- `GET /api/v1/urls/:code/stats`
- `GET /health`
- `GET /:code` (redirect)

## Web Endpoints

- `GET /` form page
- `POST /shorten` create from form
