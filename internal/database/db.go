package database

import (
	"context"
	"fmt"
	"net/url"
	"time"

	"csort.ru/analysis-service/internal/repository"
	"github.com/jackc/pgx/v5/pgxpool"
)

type DB struct {
	Pool *pgxpool.Pool
}

type DatabaseConfig struct {
	Host            string
	Port            string
	User            string
	Password        string
	Name            string
	MaxConns        int32
	MinConns        int32
	MaxConnLifetime int64
	MaxConnIdleTime int64
}

func New(cfg *DatabaseConfig) (*DB, error) {
	pool, err := connect(*cfg)
	if err != nil {
		return nil, err
	}
	return &DB{Pool: pool}, nil
}

func (db *DB) Close() {
	if db.Pool != nil {
		db.Pool.Close()
	}
}

func NewQueries(pool *pgxpool.Pool) *repository.Queries {
	return repository.New(pool)
}

func connect(dbCfg DatabaseConfig) (*pgxpool.Pool, error) {
	dsn := fmt.Sprintf("postgres://%s:%s@%s:%s/%s?sslmode=disable",
		dbCfg.User, url.QueryEscape(dbCfg.Password), dbCfg.Host, dbCfg.Port, dbCfg.Name)

	poolCfg, err := pgxpool.ParseConfig(dsn)
	if err != nil {
		return nil, fmt.Errorf("failed to parse config: %w", err)
	}

	poolCfg.MaxConns = dbCfg.MaxConns
	poolCfg.MinConns = dbCfg.MinConns
	poolCfg.MaxConnLifetime = time.Duration(dbCfg.MaxConnLifetime) * time.Second
	poolCfg.MaxConnIdleTime = time.Duration(dbCfg.MaxConnIdleTime) * time.Second

	pool, err := pgxpool.NewWithConfig(context.Background(), poolCfg)
	if err != nil {
		return nil, fmt.Errorf("failed to create new pool: %w", err)
	}

	if err := pool.Ping(context.Background()); err != nil {
		pool.Close()
		return nil, fmt.Errorf("failed to ping database: %w", err)
	}

	return pool, nil
}
