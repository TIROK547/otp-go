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
