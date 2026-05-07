# user-service

Manages user profiles, avatar uploads, language preferences, and online presence for Hypertube.

## Responsibilities

- Serve public and private user profiles
- Accept profile updates (first name, last name, language preference)
- Handle avatar uploads with MIME type validation (jpeg, png, gif, webp — 5 MB max)
- Delete accounts
- Report online/offline status via Redis presence

## Endpoints

| Method   | Path                              | Auth | Description                        |
|----------|-----------------------------------|------|------------------------------------|
| `GET`    | `/health`                         | —    | Health check                       |
| `GET`    | `/api/v1/users/profile/:id`       | —    | Get a user's public profile        |
| `GET`    | `/api/v1/users/:id/online-status` | —    | Get online status                  |
| `GET`    | `/api/v1/users/avatars/*filename` | —    | Serve an uploaded avatar file      |
| `GET`    | `/api/v1/users/profile`           | JWT  | Get own profile                    |
| `PUT`    | `/api/v1/users/profile/:id`       | JWT  | Update own profile                 |
| `DELETE` | `/api/v1/users/profile/:id`       | JWT  | Delete own account                 |
| `POST`   | `/api/v1/users/avatar`            | JWT  | Upload avatar (`multipart/form-data`, field `avatar`) |

Auth is enforced by the API Gateway, which validates the JWT and forwards `X-User-ID` on protected routes.

### `PUT /api/v1/users/profile/:id` — request body

```json
{
  "first_name": "Jean",
  "last_name":  "Dupont",
  "language":   "fr"
}
```

All fields are optional. `language` accepts `fr` or `en`.

### `POST /api/v1/users/avatar` — response

```json
{
  "data": {
    "avatar_url": "/api/v1/users/avatars/42_1746612345678901234.jpg"
  }
}
```

## Data model

```
users
├── id            SERIAL PK
├── username      VARCHAR(50) UNIQUE NOT NULL
├── email         VARCHAR(255) UNIQUE NOT NULL
├── password_hash TEXT NOT NULL
├── first_name    VARCHAR(100)
├── last_name     VARCHAR(100)
├── avatar_url    TEXT
├── language      VARCHAR(10) DEFAULT 'fr'
├── email_verified BOOLEAN DEFAULT FALSE
├── created_at    TIMESTAMP
└── updated_at    TIMESTAMP
```

## Environment variables

| Variable      | Default      | Description                          |
|---------------|--------------|--------------------------------------|
| `DB_HOST`     | `localhost`  | PostgreSQL host                      |
| `DB_PORT`     | `5432`       | PostgreSQL port                      |
| `DB_NAME`     | `hypertube`  | Database name                        |
| `DB_USER`     | `postgres`   | Database user                        |
| `DB_PASSWORD` | `password`   | Database password                    |
| `REDIS_HOST`  | `redis`      | Redis host                           |
| `REDIS_PORT`  | `6379`       | Redis port                           |
| `AVATAR_DIR`  | `/data/avatars` | Directory where avatars are stored |
| `AUTO_MIGRATE`| `false`      | Run GORM AutoMigrate on start        |

The service listens on port **8002**.

## Layout

```
src/
├── conf/        # DB and Redis initialization
├── handlers/    # HTTP handlers
│   ├── avatar_upload.go      # POST /avatar — MIME validation + storage
│   ├── GetOwnProfileHandler.go
│   ├── online_status.go
│   ├── profile_delete.go
│   ├── profile_get.go
│   ├── profile_update.go
│   └── types.go              # Request structs
├── middleware/  # Auth middleware (reads X-User-ID forwarded by gateway)
├── models/      # User GORM model
├── services/    # Presence service (Redis — online/offline tracking)
└── utils/       # Response helpers, error helpers, validation
```

## Development

Hot reload via [Air](https://github.com/air-verse/air) — source files are volume-mounted:

```bash
docker compose -f docker-compose.dev.yml up user-service
```

Run tests:

```bash
cd src
go test -v ./...
```
