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
	dsn := "file:" + filepath.Join(dataDir, "feedback.db") + "?_txlock=immediate&_pragma=busy_timeout(5000)&_pragma=journal_mode(WAL)&_pragma=foreign_keys(1)&_pragma=synchronous(NORMAL)"
	db, err := sql.Open("sqlite", dsn)
	if err != nil {
		return nil, err
	}
	db.SetMaxOpenConns(1)
	db.SetMaxIdleConns(1)
	db.SetConnMaxLifetime(0)
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
	_, err := s.db.Exec(`
CREATE TABLE IF NOT EXISTS users (
 id INTEGER PRIMARY KEY, username TEXT NOT NULL UNIQUE, password_hash BLOB NOT NULL, created_at DATETIME NOT NULL
);
CREATE TABLE IF NOT EXISTS feedback (
 id INTEGER PRIMARY KEY AUTOINCREMENT,
 category TEXT NOT NULL CHECK(category IN ('bug','feature','question')),
 category_number INTEGER,
 title TEXT NOT NULL, body TEXT NOT NULL, author TEXT NOT NULL,
 game_version TEXT NOT NULL DEFAULT '', mod_version TEXT NOT NULL DEFAULT '',
 mod_list TEXT NOT NULL DEFAULT '', save_link TEXT NOT NULL DEFAULT '',
 status TEXT NOT NULL DEFAULT 'open' CHECK(status IN ('open','in_progress','resolved','closed')),
 created_at DATETIME NOT NULL
);
CREATE INDEX IF NOT EXISTS feedback_created_at ON feedback(created_at DESC);`)
	if err != nil {
		return err
	}
	if err = s.ensureCategoryNumber(); err != nil {
		return err
	}
	_, err = s.db.Exec(`CREATE UNIQUE INDEX IF NOT EXISTS feedback_category_number ON feedback(category, category_number)`)
	return err
}

func (s *Store) ensureCategoryNumber() error {
	var missing int
	if err := s.db.QueryRow(`SELECT COUNT(*) FROM pragma_table_info('feedback') WHERE name = 'category_number'`).Scan(&missing); err != nil {
		return err
	}
	if missing == 0 {
		if _, err := s.db.Exec(`ALTER TABLE feedback ADD COLUMN category_number INTEGER`); err != nil {
			return err
		}
	}
	tx, err := s.db.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()
	if _, err = tx.Exec(`UPDATE feedback AS current SET category_number = (SELECT COUNT(*) FROM feedback AS previous WHERE previous.category = current.category AND previous.id <= current.id) WHERE category_number IS NULL`); err != nil {
		return err
	}
	return tx.Commit()
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
func (s *Store) ListFeedback(category string, limit, offset int) ([]domain.Feedback, error) {
	if limit < 1 || limit > 100 {
		limit = 50
	}
	if offset < 0 {
		offset = 0
	}
	query := `SELECT id,category,category_number,title,body,author,game_version,mod_version,mod_list,save_link,status,created_at FROM feedback`
	args := []any{}
	if category != "" {
		query += " WHERE category = ?"
		args = append(args, category)
	}
	query += " ORDER BY created_at DESC, id DESC LIMIT ? OFFSET ?"
	args = append(args, limit, offset)
	rows, err := s.db.Query(query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	items := make([]domain.Feedback, 0, limit)
	for rows.Next() {
		var item domain.Feedback
		if err := rows.Scan(&item.ID, &item.Category, &item.CategoryNumber, &item.Title, &item.Body, &item.Author, &item.GameVersion, &item.ModVersion, &item.ModList, &item.SaveLink, &item.Status, &item.CreatedAt); err != nil {
			return nil, err
		}
		items = append(items, item)
	}
	return items, rows.Err()
}

func (s *Store) CreateFeedback(item domain.Feedback) (domain.Feedback, error) {
	item.CreatedAt = time.Now().UTC()
	item.Status = domain.StatusOpen
	tx, err := s.db.Begin()
	if err != nil {
		return domain.Feedback{}, err
	}
	defer tx.Rollback()
	if err = tx.QueryRow(`SELECT COALESCE(MAX(category_number), 0) + 1 FROM feedback WHERE category = ?`, item.Category).Scan(&item.CategoryNumber); err != nil {
		return domain.Feedback{}, err
	}
	result, err := tx.Exec(`INSERT INTO feedback(category,category_number,title,body,author,game_version,mod_version,mod_list,save_link,created_at) VALUES(?,?,?,?,?,?,?,?,?,?)`, item.Category, item.CategoryNumber, item.Title, item.Body, item.Author, item.GameVersion, item.ModVersion, item.ModList, item.SaveLink, item.CreatedAt)
	if err != nil {
		return domain.Feedback{}, err
	}
	item.ID, err = result.LastInsertId()
	if err != nil {
		return domain.Feedback{}, err
	}
	if err = tx.Commit(); err != nil {
		return domain.Feedback{}, err
	}
	return item, nil
}

func (s *Store) UpdateStatus(id int64, status string) (bool, error) {
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
