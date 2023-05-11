package postgres

import (
	"context"
	"fmt"
	"github.com/georgysavva/scany/v2/pgxscan"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
	_ "github.com/lib/pq"
)

type Config struct {
	Host     string
	Port     string
	User     string
	Password string
	DB       string
}

func (config Config) String() string {
	return fmt.Sprintf("postgres://%s:%s@%s:%s/%s?sslmode=disable", config.User, config.Password, config.Host,
		config.Port, config.DB)
}

type Client interface {
	Exec(ctx context.Context, query string, args ...any) (pgconn.CommandTag, error)
	Query(ctx context.Context, query string, args ...any) (pgx.Rows, error)
	QueryRow(ctx context.Context, query string, args ...any) pgx.Row
	Get(ctx context.Context, dest interface{}, query string, args ...interface{}) error
	Select(ctx context.Context, dest interface{}, query string, args ...interface{}) error
}

type client struct {
	pool *pgxpool.Pool
}

func NewClient(ctx context.Context, cfg Config) (Client, error) {
	config, err := pgxpool.ParseConfig(cfg.String())
	if err != nil {
		return nil, err
	}

	config.MinConns = 10

	var pool *pgxpool.Pool
	pool, err = pgxpool.NewWithConfig(ctx, config)
	if err != nil {
		return nil, err
	}
	err = pool.Ping(ctx)
	if err != nil {
		return nil, err
	}

	return &client{
		pool: pool,
	}, nil
}

func (c *client) Exec(ctx context.Context, query string, args ...any) (pgconn.CommandTag, error) {
	return c.pool.Exec(ctx, query, args...)
}

func (c *client) Get(ctx context.Context, dest interface{}, query string, args ...any) error {
	return pgxscan.Get(ctx, c.pool, dest, query, args...)
}

func (c *client) Select(ctx context.Context, dest interface{}, query string, args ...any) error {
	return pgxscan.Select(ctx, c.pool, dest, query, args...)
}

func (c *client) Query(ctx context.Context, query string, args ...any) (pgx.Rows, error) {
	return c.pool.Query(ctx, query, args...)
}

func (c *client) QueryRow(ctx context.Context, query string, args ...any) pgx.Row {
	return c.pool.QueryRow(ctx, query, args...)
}
