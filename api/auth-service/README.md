# auth-service

Handles authentication and account management for Hypertube.

## Responsibilities

- User registration and login
- JWT access + refresh token issuance and rotation
- Token blacklisting on logout (Redis)
- Email verification (6-digit code)
- Password reset via email link

## Endpoints

| Method | Path | Auth | Description |
|--------|------|------|-------------|
| `GET` | `/health` | — | Health check |
| `POST` | `/api/v1/auth/check-availability` | — | Check username/email availability |
| `POST` | `/api/v1/auth/register` | — | Register a new account |
| `POST` | `/api/v1/auth/login` | — | Login, receive token pair |
| `POST` | `/api/v1/auth/logout` | JWT | Invalidate access token |
| `POST` | `/api/v1/auth/refresh` | — | Exchange refresh token for new pair |
| `GET` | `/api/v1/auth/verify` | JWT | Verify token validity |
| `POST` | `/api/v1/auth/send-email-verification` | — | Send verification code by email |
| `POST` | `/api/v1/auth/verify-email` | — | Submit verification code |
| `POST` | `/api/v1/auth/forgot-password` | — | Send password reset email |
| `POST` | `/api/v1/auth/reset-password` | — | Reset password with token |

## Token flow

```
POST /register  or  POST /login
  → { access_token, refresh_token, expires_in }

POST /refresh   (body: { refresh_token })
  → { access_token, refresh_token }    ← rotated pair

POST /logout    (Authorization: Bearer <token>)
  → access_token added to Redis blacklist
```

The gateway validates every access token and forwards `X-User-ID` to downstream services.

## Environment variables

| Variable | Default | Description |
|----------|---------|-------------|
| `PORT` | `8001` | Listening port |
| `DB_HOST` | `localhost` | PostgreSQL host |
| `DB_PORT` | `5432` | PostgreSQL port |
| `DB_NAME` | `hypertube` | Database name |
| `DB_USER` | `postgres` | Database user |
| `DB_PASSWORD` | `password` | Database password |
| `REDIS_HOST` | `redis` | Redis host (token blacklist) |
| `REDIS_PORT` | `6379` | Redis port |
| `JWT_SECRET` | — | **Required.** Secret for signing JWTs |
| `SMTP_HOST` | `smtp.gmail.com` | SMTP server |
| `SMTP_PORT` | `587` | SMTP port |
| `SMTP_USERNAME` | — | SMTP login |
| `SMTP_PASSWORD` | — | SMTP password |
| `FROM_EMAIL` | `noreply@hypertube.app` | Sender address |
| `FROM_NAME` | `Hypertube` | Sender display name |
| `FRONTEND_URL` | `http://localhost:3000` | Base URL used in email links |
| `AUTO_MIGRATE` | `false` | Run GORM AutoMigrate on start |

If `SMTP_USERNAME` or `SMTP_PASSWORD` is empty, emails are printed to stdout instead of being sent (dev mode).

## Layout

```
src/
├── conf/           # DB and Redis initialization
├── handlers/       # HTTP handlers (auth, token, email, password)
├── middleware/     # JWT validation middleware
├── models/         # Users, EmailVerification, PasswordReset
├── services/       # User creation, email sending
├── types/          # Request/response types
└── utils/          # JWT helpers, validation, response, suggestions
templates/
└── email/          # HTML email templates
```

## Development

```bash
docker compose -f docker-compose.dev.yml up auth-service
```

Source files are volume-mounted — Air reloads on every `.go` change.
