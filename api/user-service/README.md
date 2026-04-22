# user-service

Manages user profiles and online presence for Hypertube.

## Responsibilities

- Serve public and private user profiles
- Accept profile updates (name, avatar)
- Delete accounts
- Report online status via Redis presence

## Endpoints

| Method | Path | Auth | Description |
|--------|------|------|-------------|
| `GET` | `/health` | — | Health check |
| `GET` | `/api/v1/users/profile/:id` | — | Get a user's public profile |
| `GET` | `/api/v1/users/:id/online-status` | — | Get online status |
| `GET` | `/api/v1/users/profile` | JWT | Get own profile |
| `PUT` | `/api/v1/users/profile/:id` | JWT | Update own profile |
| `DELETE` | `/api/v1/users/profile/:id` | JWT | Delete own account |

Auth is enforced by the API Gateway, which forwards `X-User-ID` on protected routes.

## Environment variables

| Variable | Default | Description |
|----------|---------|-------------|
| `PORT` | `8002` | Listening port |
| `DB_HOST` | `localhost` | PostgreSQL host |
| `DB_PORT` | `5432` | PostgreSQL port |
| `DB_NAME` | `hypertube` | Database name |
| `DB_USER` | `postgres` | Database user |
| `DB_PASSWORD` | `password` | Database password |
| `REDIS_HOST` | `redis` | Redis host |
| `REDIS_PORT` | `6379` | Redis port |
| `AUTO_MIGRATE` | `false` | Run GORM AutoMigrate on start |

## Layout

```
src/
├── conf/           # DB and Redis initialization
├── handlers/       # HTTP handlers
├── middleware/     # Auth middleware (reads X-User-ID from gateway)
├── models/         # User GORM model
├── services/       # Presence service (Redis)
└── utils/          # Response helpers, validation
```

## Development

Hot reload via [Air](https://github.com/air-verse/air):

```bash
docker compose -f docker-compose.dev.yml up user-service
```

Source files are volume-mounted — changes reload automatically.
