package store

import (
	"database/sql"
	"errors"
	"path/filepath"
	"time"

	"golang.org/x/crypto/bcrypt"
	"modders-feedback/server/internal/domain"
	_ "modernc.org/sqlite"
)

type Store struct {
	db *sql.DB
}

func Open(dataDir string) (*Store, error) {
	db, err := sql.Open("sqlite", filepath.Join(dataDir, "feedback.db"))
	if err != nil {
		return nil, err
	}
	db.SetMaxOpenConns(1)
	store := &Store{db: db}
	if err := store.migrate(); err != nil {
		_ = db.Close()
		return nil, err
	}
	return store, nil
}

func (s *Store) Close() error {
	return s.db.Close()
}

func (s *Store) migrate() error {
	_, err := s.db.Exec(`PRAGMA journal_mode=WAL; PRAGMA foreign_keys=ON;
CREATE TABLE IF NOT EXISTS users (
 id INTEGER PRIMARY KEY, username TEXT NOT NULL UNIQUE, password_hash BLOB NOT NULL, created_at DATETIME NOT NULL
);
CREATE TABLE IF NOT EXISTS feedback (
 id INTEGER PRIMARY KEY AUTOINCREMENT,
 category TEXT NOT NULL CHECK(category IN ('bug','feature','question')),
 title TEXT NOT NULL, body TEXT NOT NULL, author TEXT NOT NULL,
 game_version TEXT NOT NULL DEFAULT '', mod_version TEXT NOT NULL DEFAULT '',
 mod_list TEXT NOT NULL DEFAULT '', save_link TEXT NOT NULL DEFAULT '',
 status TEXT NOT NULL DEFAULT 'open' CHECK(status IN ('open','in_progress','resolved','closed')),
 created_at DATETIME NOT NULL
);
CREATE INDEX IF NOT EXISTS feedback_created_at ON feedback(created_at DESC);`)
	return err
}

func (s *Store) SeedAdmin(username, password string) error {
	if username == "" && password == "" {
		return nil
	}
	if username == "" || len(password) < 12 {
		return errors.New("set ADMIN_USERNAME and ADMIN_PASSWORD (at least 12 characters) together")
	}
	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return err
	}
	_, err = s.db.Exec(
		`INSERT INTO users(username,password_hash,created_at) VALUES(?,?,?) ON CONFLICT(username) DO NOTHING`,
		username, hash, time.Now().UTC(),
	)
	return err
}

func (s *Store) Authenticate(username, password string) (bool, error) {
	var hash []byte
	err := s.db.QueryRow("SELECT password_hash FROM users WHERE username = ?", username).Scan(&hash)
	if errors.Is(err, sql.ErrNoRows) {
		return false, nil
	}
	if err != nil {
		return false, err
	}
	if bcrypt.CompareHashAndPassword(hash, []byte(password)) != nil {
		return false, nil
	}
	return true, nil
}

func (s *Store) ListFeedback(category string) ([]domain.Feedback, error) {
	query := `SELECT id,category,title,body,author,game_version,mod_version,mod_list,save_link,status,created_at FROM feedback`
	args := []any{}
	if category != "" {
		query += " WHERE category = ?"
		args = append(args, category)
	}
	query += " ORDER BY created_at DESC LIMIT 200"
	rows, err := s.db.Query(query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	items := make([]domain.Feedback, 0)
	for rows.Next() {
		var item domain.Feedback
		if err := rows.Scan(
			&item.ID, &item.Category, &item.Title, &item.Body, &item.Author,
			&item.GameVersion, &item.ModVersion, &item.ModList, &item.SaveLink,
			&item.Status, &item.CreatedAt,
		); err != nil {
			return nil, err
		}
		items = append(items, item)
	}
	return items, rows.Err()
}

func (s *Store) CreateFeedback(item domain.Feedback) (domain.Feedback, error) {
	item.CreatedAt = time.Now().UTC()
	item.Status = domain.StatusOpen
	result, err := s.db.Exec(
		`INSERT INTO feedback(category,title,body,author,game_version,mod_version,mod_list,save_link,created_at) VALUES(?,?,?,?,?,?,?,?,?)`,
		item.Category, item.Title, item.Body, item.Author, item.GameVersion, item.ModVersion, item.ModList, item.SaveLink, item.CreatedAt,
	)
	if err != nil {
		return domain.Feedback{}, err
	}
	item.ID, err = result.LastInsertId()
	if err != nil {
		return domain.Feedback{}, err
	}
	return item, nil
}

func (s *Store) UpdateStatus(id, status string) (bool, error) {
	result, err := s.db.Exec("UPDATE feedback SET status = ? WHERE id = ?", status, id)
	if err != nil {
		return false, err
	}
	count, err := result.RowsAffected()
	if err != nil {
		return false, err
	}
	return count > 0, nil
}
