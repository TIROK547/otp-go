# OTP GO

Backend service in Golang that provides OTP-based login & registration, rate limiting, user management, OpenAPI docs, and Docker support.

## Features
- OTP generation (printed to app console)
- OTP stored in Redis with 2 minute expiration
- Rate limiting: max 3 OTP requests per phone within 10 minutes
- Registration/login via phone+OTP
- JWT token issued on success
- Users persisted in PostgreSQL (phone, created_at)
- REST endpoints and OpenAPI spec
- Dockerized, docker-compose included (app + postgres + redis)

## Why Postgres + Redis
Redis is used for ephemeral data (OTP + rate counters) because it provides TTL and atomic counters with low latency. Postgres is used for persistent user storage so user data survives restarts and is easy to query (search, pagination). This separation keeps transient and persistent concerns in the appropriate systems.

## Run locally (no Docker)
1. Install PostgreSQL and Redis locally
2. Create a Postgres database and set `DATABASE_URL` env var
3. Set `REDIS_ADDR` and `JWT_SECRET`
4. `go build ./cmd/server` then run the binary

Example env:
```bash
export DATABASE_URL=postgres://postgres:postgres@localhost:5432/postgres?sslmode=disable
export REDIS_ADDR=localhost:6379
export JWT_SECRET=supersecret
```
## Run:
./otp-go

## Run with Docker (recommended)
1. Build and run:
```bash
docker compose up --build
```
2. App will be available at `http://localhost:8080`

## APIs

1. Generate OTP
POST /v1/otp/send
Content-Type: application/json
Body: {"phone": "+441234567890"}
Response: { "status": "ok", "message": "OTP generated (printed to console)" }
Console prints:
OTP for +441234567890 : 123456

2. Verify OTP (login/register)
POST /v1/otp/verify
Content-Type: application/json
Body: {"phone":"+441234567890", "otp":"123456"}
Response: {"token":"<jwt>", "user": {"id":1,"phone":"+441234567890","created_at":"2025-08-10T...Z"}}

3. Get user
GET /v1/users/{id}
Response: {"id":1,"phone":"+441234567890","created_at":"..."}

4. List users
GET /v1/users?q=+44&limit=10&offset=0
Response: {"users":[...],"total":123}


## Rate limiting behaviour
- Each time `/v1/otp/send` is called for a phone, an internal Redis counter `rl:{phone}` increments and expires in 10 minutes.
- If the counter > 3, requests return HTTP 429.

## Notes and extensions
- You can wire up real SMS by replacing the console print with an SMS provider call.
- JWT secret should be strong in production.
- Add HTTPS and auth middleware for protected endpoints.

## Example curl flow
```bash
curl -s -X POST http://localhost:8080/v1/otp/send -d '{"phone":"+441234567890"}' -H "Content-Type: application/json"

check console for OTP 6-digit
curl -s -X POST http://localhost:8080/v1/otp/verify -d '{"phone":"+441234567890","otp":"123456"}' -H "Content-Type: application/json"
```