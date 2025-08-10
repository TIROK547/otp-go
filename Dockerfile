FROM golang:1.21-alpine AS build
WORKDIR /app
COPY go.mod go.sum ./
RUN go env -w GOFLAGS=-mod=mod
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -o /otp-go ./cmd/server

FROM alpine:latest
COPY --from=build /otp-go /otp-go
ENV DATABASE_URL=postgres://postgres:postgres@postgres:5432/postgres?sslmode=disable
ENV REDIS_ADDR=redis:6379
ENV JWT_SECRET=replace-me
EXPOSE 8080
CMD ["/otp-go"]
