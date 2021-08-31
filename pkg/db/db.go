package db

import (
	"context"
	"errors"
	"fmt"
	"time"

	"database/sql"
	_ "github.com/jackc/pgx/v4/stdlib"

	"url_shortener/pkg/config"

	log "github.com/sirupsen/logrus"
)

// ShortenerDB database interface for URL shortener service
type ShortenerDB interface {
	Add(ctx context.Context, row Row) error
	GetOriginalURL(ctx context.Context, shortURL string) (string, error)
	Close() error
}

type Row struct {
	OriginalURL string
	ShortURL    string
}

type NoRowError struct{}

func (e *NoRowError) Error() string {
	return "row doesn't exist"
}

type DB struct {
	cfg config.DBConfig

	db *sql.DB
}

func New(ctx context.Context, cfg config.DBConfig) (ShortenerDB, error) {
	sdb := &DB{cfg: cfg}
	var err error

	sdb.db, err = sql.Open("pgx", cfg.ConnectURL())
	if err != nil {
		return nil, fmt.Errorf("db: cannot open database: %w", err)
	}

	var connErr error
	for i := 1; i <= cfg.ConnTriesCnt; i++ {
		log.Printf("trying to connect to database #%d", i)
		if connErr = sdb.db.PingContext(ctx); connErr != nil {
			<-time.After(time.Duration(cfg.ConnTryTime) * time.Second)
		} else {
			log.Print("database connection established")
			break
		}
	}
	if connErr != nil {
		return nil, fmt.Errorf("db: cannot connect to database (%d tries by %d seconds): %w", cfg.ConnTriesCnt, cfg.ConnTryTime, err)
	}

	sdb.db.SetMaxOpenConns(cfg.MaxOpenConns)
	sdb.db.SetMaxIdleConns(cfg.MaxIdleConns)

	return sdb, nil
}

func (d *DB) Add(ctx context.Context, row Row) error {
	if err := d.add(ctx, row); err != nil {
		return fmt.Errorf("db: cannot add row: original_url=%s, short_url=%s: %w", row.OriginalURL, row.ShortURL, err)
	}
	return nil
}

func (d *DB) add(ctx context.Context, row Row) error {
	tx, err := d.db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("cannot create transaction: %w", err)
	}
	defer func() { _ = tx.Rollback() }()

	_, err = tx.ExecContext(
		ctx,
		"INSERT INTO url_db(original_url, short_url) VALUES ($1, $2) ON CONFLICT (original_url) DO NOTHING;",
		row.OriginalURL,
		row.ShortURL,
	)
	if err != nil {
		return fmt.Errorf("cannot exec query: %w", err)
	}

	if err = tx.Commit(); err != nil {
		return fmt.Errorf("cannot commit transaction: %w", err)
	}

	return nil
}

func (d *DB) GetOriginalURL(ctx context.Context, shortURL string) (string, error) {
	originalURL, err := d.getOriginalURL(ctx, shortURL)
	if err != nil {
		return "", fmt.Errorf("db: cannot get original_url by short_url=%s: %w", shortURL, err)
	}
	return originalURL, nil
}

func (d *DB) getOriginalURL(ctx context.Context, shortURL string) (original string, err error) {
	row := d.db.QueryRowContext(
		ctx,
		"SELECT original_url FROM url_db WHERE short_url = $1",
		shortURL,
	)
	if err = row.Scan(&original); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return "", &NoRowError{}
		}
		return "", fmt.Errorf("cannot scan row: %w", err)
	}
	return
}

func (d *DB) Close() error {
	return d.db.Close()
}
