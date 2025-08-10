package db

import (
	"context"
	"errors"
	"time"

	"otp-go/internal/models"

	"github.com/jackc/pgx/v5/pgxpool"
)

type Postgres struct {
	pool *pgxpool.Pool
}

func NewPostgres(ctx context.Context, dsn string) (*Postgres, error) {
	if dsn == "" {
		return nil, errors.New("DATABASE_URL required")
	}
	pool, err := pgxpool.New(ctx, dsn)
	if err != nil {
		return nil, err
	}
	p := &Postgres{pool: pool}
	if err := p.migrate(ctx); err != nil {
		return nil, err
	}
	return p, nil
}

func (p *Postgres) migrate(ctx context.Context) error {
	_, err := p.pool.Exec(ctx, `
CREATE TABLE IF NOT EXISTS users (
  id BIGSERIAL PRIMARY KEY,
  phone TEXT UNIQUE NOT NULL,
  created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
`)
	return err
}

func (p *Postgres) Close(ctx context.Context) error {
	p.pool.Close()
	return nil
}

func (p *Postgres) CreateUserIfNotExists(ctx context.Context, phone string) (*models.User, error) {
	u, err := p.GetUserByPhone(ctx, phone)
	if err == nil && u != nil {
		return u, nil
	}
	var id int64
	var createdAt time.Time
	err = p.pool.QueryRow(ctx, "INSERT INTO users (phone) VALUES ($1) ON CONFLICT (phone) DO UPDATE SET phone = users.phone RETURNING id, created_at", phone).Scan(&id, &createdAt)
	if err != nil {
		return nil, err
	}
	return &models.User{ID: id, Phone: phone, CreatedAt: createdAt}, nil
}

func (p *Postgres) GetUserByPhone(ctx context.Context, phone string) (*models.User, error) {
	var id int64
	var createdAt time.Time
	err := p.pool.QueryRow(ctx, "SELECT id, created_at FROM users WHERE phone=$1", phone).Scan(&id, &createdAt)
	if err != nil {
		return nil, err
	}
	return &models.User{ID: id, Phone: phone, CreatedAt: createdAt}, nil
}

func (p *Postgres) GetUserByID(ctx context.Context, id int64) (*models.User, error) {
	var phone string
	var createdAt time.Time
	err := p.pool.QueryRow(ctx, "SELECT phone, created_at FROM users WHERE id=$1", id).Scan(&phone, &createdAt)
	if err != nil {
		return nil, err
	}
	return &models.User{ID: id, Phone: phone, CreatedAt: createdAt}, nil
}

func (p *Postgres) ListUsers(ctx context.Context, limit, offset int, q string) ([]models.User, int, error) {
	rows, err := p.pool.Query(ctx, "SELECT id, phone, created_at FROM users WHERE phone ILIKE $1 ORDER BY created_at DESC LIMIT $2 OFFSET $3", "%"+q+"%", limit, offset)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()
	users := []models.User{}
	for rows.Next() {
		var u models.User
		if err := rows.Scan(&u.ID, &u.Phone, &u.CreatedAt); err != nil {
			return nil, 0, err
		}
		users = append(users, u)
	}
	var total int
	err = p.pool.QueryRow(ctx, "SELECT COUNT(*) FROM users WHERE phone ILIKE $1", "%"+q+"%").Scan(&total)
	if err != nil {
		return nil, 0, err
	}
	return users, total, nil
}
