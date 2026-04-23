# auth-service

Handles authentication and account management for Hypertube.

## Responsibilities

- User registration and login
- JWT access + refresh token issuance and rotation
- Token blacklisting on logout (Redis)
- Email verification (6-digit code, sent automatically on registration)
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
| `POST` | `/api/v1/auth/send-email-verification` | — | Re-send verification code |
| `POST` | `/api/v1/auth/verify-email` | — | Submit verification code |
| `POST` | `/api/v1/auth/forgot-password` | — | Send password reset email |
| `POST` | `/api/v1/auth/reset-password` | — | Reset password with token |

### POST /api/v1/auth/register

```json
// Request
{
  "username":   "jdoe",
  "email":      "jane@example.com",
  "password":   "min8chars",
  "first_name": "Jane",
  "last_name":  "Doe"
}

// 201 Created
{
  "success": true,
  "data": {
    "message":       "User registered successfully",
    "user":          { "id": 1, "username": "jdoe", "email": "jane@example.com" },
    "access_token":  "<jwt>",
    "refresh_token": "<jwt>",
    "token_type":    "Bearer",
    "expires_in":    21600
  }
}
```

A verification code is sent to the email address immediately after registration. The account is usable before verification, but downstream services may gate features on `email_verified`.

Errors: `409` username taken (+ suggestions) · `409` email taken · `400` invalid payload

### POST /api/v1/auth/login

```json
// Request — login accepts username or email
{ "login": "jdoe", "password": "min8chars" }

// 200 OK — same shape as /register
```

Errors: `401` invalid credentials

### POST /api/v1/auth/logout

Header: `Authorization: Bearer <access_token>`

```json
// 200 OK
{ "success": true, "data": { "message": "logged out successfully" } }
```

Adds the token to the Redis blacklist with a TTL equal to its remaining lifetime.

### POST /api/v1/auth/refresh

```json
// Request
{ "refresh_token": "<jwt>" }

// 200 OK — rotated pair
{
  "success": true,
  "data": {
    "access_token":  "<jwt>",
    "refresh_token": "<jwt>",
    "token_type":    "Bearer",
    "expires_in":    21600
  }
}
```

Errors: `401` invalid or wrong-scope token

### GET /api/v1/auth/verify

Header: `Authorization: Bearer <access_token>`

```json
// 200 OK
{ "success": true, "data": { "valid": true, "user_id": "1" } }
```

Errors: `401` missing/invalid/expired token

### POST /api/v1/auth/check-availability

```json
// Request — at least one field required
{ "username": "jdoe", "email": "jane@example.com" }

// 200 OK
{ "status": "success", "available": true }

// 409 Conflict — username taken
{ "status": "error", "available": false, "message": "username déjà utilisé", "suggestions": ["jdoe42", "jdoe_"] }

// 409 Conflict — email taken
{ "status": "error", "available": false, "message": "Email déjà utilisé" }
```

### POST /api/v1/auth/send-email-verification

Resends a code if the account is not yet verified. The code expires after 15 minutes.

```json
// Request
{ "email": "jane@example.com" }

// 200 OK
{ "success": true, "data": { "message": "Verification code sent successfully" } }
```

Errors: `404` user not found · `400` already verified

### POST /api/v1/auth/verify-email

```json
// Request
{ "email": "jane@example.com", "verification_code": "482931" }

// 200 OK
{ "success": true, "data": { "message": "Email verified successfully" } }
```

Errors: `400` invalid code · `400` expired code

### POST /api/v1/auth/forgot-password

Always responds with success to prevent user enumeration.

```json
// Request
{ "email": "jane@example.com" }

// 200 OK
{ "success": true, "data": { "message": "If the email exists, a password reset link will be sent" } }
```

### POST /api/v1/auth/reset-password

Token is valid for 1 hour and single-use.

```json
// Request
{ "token": "<reset_token>", "new_password": "newmin8chars" }

// 200 OK
{ "success": true, "data": { "message": "Password reset successful" } }
```

Errors: `400` invalid/expired token · `400` same as current password

---

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

---

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
| `JWT_ACCESS_TTL` | `6h` | Access token lifetime |
| `JWT_REFRESH_TTL` | `168h` | Refresh token lifetime |
| `JWT_REFRESH_SECRET` | `JWT_SECRET` | Separate secret for refresh tokens (optional) |
| `SMTP_HOST` | `smtp.gmail.com` | SMTP server |
| `SMTP_PORT` | `587` | SMTP port |
| `SMTP_USERNAME` | — | SMTP login |
| `SMTP_PASSWORD` | — | SMTP password |
| `FROM_EMAIL` | `noreply@hypertube.app` | Sender address |
| `FROM_NAME` | `Hypertube` | Sender display name |
| `FRONTEND_URL` | `http://localhost:3000` | Base URL used in email links |
| `AUTO_MIGRATE` | `false` | Run GORM AutoMigrate on start |

If `SMTP_USERNAME` or `SMTP_PASSWORD` is empty, emails are printed to stdout instead of being sent (dev mode).

---

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
