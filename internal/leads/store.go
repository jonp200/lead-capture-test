package leads

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"modernc.org/sqlite"
)

// ErrDuplicateEmail indicates the unique email constraint was violated.
var ErrDuplicateEmail = errors.New("email already exists")

// Store owns a SQLite connection pool.
type Store struct{ db *sql.DB }

// Open opens and initializes a persistent database.
func Open(ctx context.Context, path string) (*Store, error) {
	db, err := sql.Open("sqlite", path)
	if err != nil {
		return nil, fmt.Errorf("open database: %w", err)
	}
	db.SetMaxOpenConns(1)
	_, err = db.ExecContext(ctx, `
		PRAGMA busy_timeout = 5000;
		CREATE TABLE IF NOT EXISTS leads (
			id INTEGER PRIMARY KEY,
			email TEXT NOT NULL UNIQUE COLLATE NOCASE,
			data TEXT NOT NULL
		);`)
	if err != nil {
		return nil, errors.Join(fmt.Errorf("initialize database: %w", err), db.Close())
	}
	return &Store{db: db}, nil
}

// Close releases database resources.
func (store *Store) Close() error { return store.db.Close() }

// List returns leads in newest-first insertion order, including an empty array.
func (store *Store) List(ctx context.Context) ([]Lead, error) {
	rows, err := store.db.QueryContext(ctx, "SELECT id, data FROM leads ORDER BY id DESC")
	if err != nil {
		return nil, fmt.Errorf("query leads: %w", err)
	}
	defer rows.Close()
	leads := make([]Lead, 0)
	for rows.Next() {
		var lead Lead
		var id int64
		var data string
		if err = rows.Scan(&id, &data); err != nil {
			return nil, fmt.Errorf("scan lead: %w", err)
		}
		if err = json.Unmarshal([]byte(data), &lead); err != nil {
			return nil, fmt.Errorf("decode lead: %w", err)
		}
		lead.ID = id
		leads = append(leads, lead)
	}
	return leads, rows.Err()
}

// Add inserts a normalized, validated lead and returns its generated metadata.
func (store *Store) Add(ctx context.Context, lead Lead) (Lead, error) {
	lead.ID = 0
	lead.CreatedAt = time.Now().UTC().Format(time.RFC3339)
	data, err := json.Marshal(lead)
	if err != nil {
		return Lead{}, fmt.Errorf("encode lead: %w", err)
	}
	result, err := store.db.ExecContext(ctx, "INSERT INTO leads (email, data) VALUES (?, ?)", lead.Email, string(data))
	if err != nil {
		var sqliteErr *sqlite.Error
		const uniqueConstraint = 2067
		if errors.As(err, &sqliteErr) && sqliteErr.Code() == uniqueConstraint {
			return Lead{}, ErrDuplicateEmail
		}
		return Lead{}, fmt.Errorf("insert lead: %w", err)
	}
	lead.ID, err = result.LastInsertId()
	return lead, err
}
