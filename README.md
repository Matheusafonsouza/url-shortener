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
2. Create database and run migration:

```sql
\i migrations/0001_create_short_urls.sql
```

3. Install dependencies:

```bash
go mod tidy
```

4. Run the server:

```bash
go run ./cmd/server
```

Server starts at `http://localhost:8080` by default.

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
