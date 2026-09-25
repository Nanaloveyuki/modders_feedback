package store

import (
	"database/sql"
	"errors"
	"path/filepath"
	"strings"
	"time"

	"golang.org/x/crypto/bcrypt"
	"modders-feedback/server/internal/domain"
	_ "modernc.org/sqlite"
)

type Store struct {
	db      *sql.DB
	dataDir string
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
	store := &Store{db: db, dataDir: dataDir}
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
 status TEXT NOT NULL DEFAULT 'open' CHECK(status IN ('open','in_progress','resolved','closed','withdrawn')),
 created_at DATETIME NOT NULL
);
CREATE TABLE IF NOT EXISTS site_settings (
 key TEXT PRIMARY KEY,
 value TEXT NOT NULL
);
CREATE TABLE IF NOT EXISTS mods (
 id INTEGER PRIMARY KEY AUTOINCREMENT,
 slug TEXT NOT NULL UNIQUE,
 name TEXT NOT NULL,
 game_version TEXT NOT NULL DEFAULT '',
 mod_version TEXT NOT NULL DEFAULT '',
 icon TEXT NOT NULL DEFAULT 'squirrel',
 sort_order INTEGER NOT NULL DEFAULT 0
);
CREATE INDEX IF NOT EXISTS feedback_created_at ON feedback(created_at DESC);`)
	if err != nil {
		return err
	}
	if err = s.ensureUsers(); err != nil {
		return err
	}
	if err = s.ensureCategoryNumber(); err != nil {
		return err
	}
	if err = s.ensureMods(); err != nil {
		return err
	}
	if err = s.ensureModLinks(); err != nil {
		return err
	}
	if err = s.ensureWithdrawnStatus(); err != nil {
		return err
	}
	_, err = s.db.Exec(`CREATE UNIQUE INDEX IF NOT EXISTS feedback_category_number ON feedback(mod_id, category, category_number)`)
	return err
}

func (s *Store) ensureWithdrawnStatus() error {
	var definition string
	if err := s.db.QueryRow(`SELECT sql FROM sqlite_master WHERE type = 'table' AND name = 'feedback'`).Scan(&definition); err != nil {
		return err
	}
	if strings.Contains(definition, "'withdrawn'") {
		return nil
	}
	tx, err := s.db.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()
	statements := []string{
		`CREATE TABLE feedback_status_migration (
 id INTEGER PRIMARY KEY AUTOINCREMENT,
 mod_id INTEGER REFERENCES mods(id),
 category TEXT NOT NULL CHECK(category IN ('bug','feature','question')),
 category_number INTEGER,
 title TEXT NOT NULL, body TEXT NOT NULL, author TEXT NOT NULL,
 game_version TEXT NOT NULL DEFAULT '', mod_version TEXT NOT NULL DEFAULT '',
 mod_list TEXT NOT NULL DEFAULT '', save_link TEXT NOT NULL DEFAULT '',
 status TEXT NOT NULL DEFAULT 'open' CHECK(status IN ('open','in_progress','resolved','closed','withdrawn')),
 created_at DATETIME NOT NULL
)`,
		`INSERT INTO feedback_status_migration(id,mod_id,category,category_number,title,body,author,game_version,mod_version,mod_list,save_link,status,created_at)
 SELECT id,mod_id,category,category_number,title,body,author,game_version,mod_version,mod_list,save_link,status,created_at FROM feedback`,
		`DROP TABLE feedback`,
		`ALTER TABLE feedback_status_migration RENAME TO feedback`,
		`CREATE INDEX IF NOT EXISTS feedback_created_at ON feedback(created_at DESC)`,
		`CREATE UNIQUE INDEX IF NOT EXISTS feedback_category_number ON feedback(mod_id, category, category_number)`,
	}
	for _, statement := range statements {
		if _, err = tx.Exec(statement); err != nil {
			return err
		}
	}
	return tx.Commit()
}

func (s *Store) ensureMods() error {
	var column int
	if err := s.db.QueryRow(`SELECT COUNT(*) FROM pragma_table_info('feedback') WHERE name = 'mod_id'`).Scan(&column); err != nil {
		return err
	}
	if column == 0 {
		if _, err := s.db.Exec(`ALTER TABLE feedback ADD COLUMN mod_id INTEGER REFERENCES mods(id)`); err != nil {
			return err
		}
		if _, err := s.db.Exec(`DROP INDEX IF EXISTS feedback_category_number`); err != nil {
			return err
		}
	}
	tx, err := s.db.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()
	var count int
	if err = tx.QueryRow(`SELECT COUNT(*) FROM mods`).Scan(&count); err != nil {
		return err
	}
	if count == 0 {
		gameVersion, modVersion, icon := "RIMWORLD 1.6", "DEV BUILD", "squirrel"
		rows, err := tx.Query(`SELECT key, value FROM site_settings`)
		if err != nil {
			return err
		}
		for rows.Next() {
			var key, value string
			if err = rows.Scan(&key, &value); err != nil {
				rows.Close()
				return err
			}
			switch key {
			case "game_version":
				gameVersion = value
			case "mod_version":
				modVersion = value
			case "icon":
				if domain.ValidIcon(value) {
					icon = value
				}
			}
		}
		if err = rows.Err(); err != nil {
			rows.Close()
			return err
		}
		rows.Close()
		if _, err = tx.Exec(
			`INSERT INTO mods(slug, name, game_version, mod_version, icon, sort_order) VALUES(?,?,?,?,?,0)`,
			domain.DefaultModSlug, "鼠族：饥与祸", gameVersion, modVersion, icon,
		); err != nil {
			return err
		}
	}
	if _, err = tx.Exec(`UPDATE feedback SET mod_id = (SELECT id FROM mods ORDER BY sort_order, id LIMIT 1) WHERE mod_id IS NULL`); err != nil {
		return err
	}
	if _, err = tx.Exec(`UPDATE feedback AS current SET category_number = (
		SELECT COUNT(*) FROM feedback AS previous
		WHERE previous.mod_id = current.mod_id AND previous.category = current.category AND previous.id <= current.id
	) WHERE category_number IS NULL`); err != nil {
		return err
	}
	return tx.Commit()
}
func (s *Store) ensureModLinks() error {
	columns := []string{"steam_url", "github_url"}
	for _, name := range columns {
		var present int
		if err := s.db.QueryRow(`SELECT COUNT(*) FROM pragma_table_info('mods') WHERE name = ?`, name).Scan(&present); err != nil {
			return err
		}
		if present == 0 {
			if _, err := s.db.Exec(`ALTER TABLE mods ADD COLUMN ` + name + ` TEXT NOT NULL DEFAULT ''`); err != nil {
				return err
			}
		}
	}
	_, err := s.db.Exec(
		`UPDATE mods SET steam_url = ?, github_url = ? WHERE slug = ? AND steam_url = '' AND github_url = ''`,
		domain.DefaultSteamURL, domain.DefaultGitHubURL, domain.DefaultModSlug,
	)
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

func (s *Store) ensureUsers() error {
	columns := map[string]string{
		"role":   "TEXT NOT NULL DEFAULT 'member'",
		"email":  "TEXT NOT NULL DEFAULT ''",
		"qq":     "TEXT NOT NULL DEFAULT ''",
		"avatar": "BLOB",
	}
	for name, definition := range columns {
		var present int
		if err := s.db.QueryRow(`SELECT COUNT(*) FROM pragma_table_info('users') WHERE name = ?`, name).Scan(&present); err != nil {
			return err
		}
		if present == 0 {
			if _, err := s.db.Exec(`ALTER TABLE users ADD COLUMN ` + name + ` ` + definition); err != nil {
				return err
			}
		}
	}
	_, err := s.db.Exec(`CREATE UNIQUE INDEX IF NOT EXISTS users_email ON users(email) WHERE email != ''`)
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
		`INSERT INTO users(username,password_hash,role,email,qq,created_at) VALUES(?,?,?,?,?,?) ON CONFLICT(username) DO UPDATE SET role = 'admin'`,
		username, hash, "admin", "", "", time.Now().UTC(),
	)
	return err
}

func (s *Store) Authenticate(username, password string) (domain.User, bool, error) {
	user, hash, err := s.userByUsername(username)
	if errors.Is(err, sql.ErrNoRows) {
		_ = bcrypt.CompareHashAndPassword([]byte("$2a$10$N9qo8uLOickgx2ZMRZoMyeIjZAgcfl7p92ldGxad68LJZdL17lhWy"), []byte(password))
		return domain.User{}, false, nil
	}
	if err != nil {
		return domain.User{}, false, err
	}
	if bcrypt.CompareHashAndPassword(hash, []byte(password)) != nil {
		return domain.User{}, false, nil
	}
	return user, true, nil
}

func (s *Store) User(username string) (domain.User, bool, error) {
	user, _, err := s.userByUsername(username)
	if errors.Is(err, sql.ErrNoRows) {
		return domain.User{}, false, nil
	}
	if err != nil {
		return domain.User{}, false, err
	}
	return user, true, nil
}

func (s *Store) Register(input domain.Registration) (domain.User, error) {
	hash, err := bcrypt.GenerateFromPassword([]byte(input.Password), bcrypt.DefaultCost)
	if err != nil {
		return domain.User{}, err
	}
	now := time.Now().UTC()
	result, err := s.db.Exec(
		`INSERT INTO users(username,password_hash,role,email,qq,created_at) VALUES(?,?,?,?,?,?)`,
		input.Username, hash, "member", input.Email, input.QQ, now,
	)
	if err != nil {
		if strings.Contains(strings.ToLower(err.Error()), "unique") {
			return domain.User{}, domain.ErrUsernameTaken
		}
		return domain.User{}, err
	}
	id, err := result.LastInsertId()
	if err != nil || id < 1 {
		return domain.User{}, err
	}
	return domain.User{Username: input.Username, Role: "member"}, nil
}

func (s *Store) UpdateProfile(currentUsername string, input domain.ProfileUpdate) (domain.User, error) {
	result, err := s.db.Exec(
		`UPDATE users SET username = ?, email = ?, qq = ? WHERE username = ? AND NOT EXISTS (SELECT 1 FROM users AS other WHERE other.username = ? AND other.username != ?)`,
		input.Username, input.Email, input.QQ, currentUsername, input.Username, currentUsername,
	)
	if err != nil {
		if strings.Contains(strings.ToLower(err.Error()), "unique") {
			return domain.User{}, domain.ErrEmailTaken
		}
		return domain.User{}, err
	}
	changed, err := result.RowsAffected()
	if err != nil {
		return domain.User{}, err
	}
	if changed == 0 {
		return domain.User{}, domain.ErrUsernameTaken
	}
	if currentUsername != input.Username {
		if _, err := s.db.Exec(`UPDATE feedback SET author = ? WHERE author = ?`, input.Username, currentUsername); err != nil {
			return domain.User{}, err
		}
	}
	user, _, err := s.userByUsername(input.Username)
	return user, err
}

func (s *Store) UpdatePassword(username, currentPassword, nextPassword string) error {
	_, hash, err := s.userByUsername(username)
	if err != nil {
		return err
	}
	if bcrypt.CompareHashAndPassword(hash, []byte(currentPassword)) != nil {
		return domain.ErrBadPassword
	}
	next, err := bcrypt.GenerateFromPassword([]byte(nextPassword), bcrypt.DefaultCost)
	if err != nil {
		return err
	}
	_, err = s.db.Exec(`UPDATE users SET password_hash = ? WHERE username = ?`, next, username)
	return err
}

func (s *Store) SetAvatar(username string, png []byte) error {
	_, err := s.db.Exec(`UPDATE users SET avatar = ? WHERE username = ?`, png, username)
	return err
}

func (s *Store) Avatar(username string) ([]byte, bool, error) {
	var png []byte
	err := s.db.QueryRow(`SELECT avatar FROM users WHERE username = ?`, username).Scan(&png)
	if errors.Is(err, sql.ErrNoRows) || len(png) == 0 {
		return nil, false, nil
	}
	if err != nil {
		return nil, false, err
	}
	return png, true, nil
}

func (s *Store) userByUsername(username string) (domain.User, []byte, error) {
	var user domain.User
	var hash []byte
	err := s.db.QueryRow(`SELECT username, role, email, qq, CASE WHEN avatar IS NULL OR length(avatar) = 0 THEN '' ELSE '/api/account/avatar?u=' || username END, password_hash FROM users WHERE username = ?`, username).Scan(&user.Username, &user.Role, &user.Email, &user.QQ, &user.AvatarURL, &hash)
	if user.Role != "admin" {
		user.Role = "member"
	}
	return user, hash, err
}
func (s *Store) ListFeedback(modID int64, category string, limit, offset int) ([]domain.Feedback, error) {
	if limit < 1 || limit > 100 {
		limit = 50
	}
	if offset < 0 {
		offset = 0
	}
	query := `SELECT id,mod_id,category,category_number,title,body,author,game_version,mod_version,mod_list,save_link,status,created_at FROM feedback WHERE mod_id = ?`
	args := []any{modID}
	if category != "" {
		query += " AND category = ?"
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
		if err := rows.Scan(&item.ID, &item.ModID, &item.Category, &item.CategoryNumber, &item.Title, &item.Body, &item.Author, &item.GameVersion, &item.ModVersion, &item.ModList, &item.SaveLink, &item.Status, &item.CreatedAt); err != nil {
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
	if err = tx.QueryRow(`SELECT COALESCE(MAX(category_number), 0) + 1 FROM feedback WHERE mod_id = ? AND category = ?`, item.ModID, item.Category).Scan(&item.CategoryNumber); err != nil {
		return domain.Feedback{}, err
	}
	result, err := tx.Exec(`INSERT INTO feedback(mod_id,category,category_number,title,body,author,game_version,mod_version,mod_list,save_link,created_at) VALUES(?,?,?,?,?,?,?,?,?,?,?)`, item.ModID, item.Category, item.CategoryNumber, item.Title, item.Body, item.Author, item.GameVersion, item.ModVersion, item.ModList, item.SaveLink, item.CreatedAt)
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

func (s *Store) DeleteFeedback(id int64) (bool, error) {
	result, err := s.db.Exec("DELETE FROM feedback WHERE id = ?", id)
	if err != nil {
		return false, err
	}
	count, err := result.RowsAffected()
	if err != nil {
		return false, err
	}
	return count > 0, nil
}

func (s *Store) UpdateFeedback(id int64, update domain.FeedbackUpdate) (domain.Feedback, bool, error) {
	result, err := s.db.Exec(
		`UPDATE feedback SET title = ?, body = ?, game_version = ?, mod_version = ?, mod_list = ?, save_link = ? WHERE id = ?`,
		update.Title, update.Body, update.GameVersion, update.ModVersion, update.ModList, update.SaveLink, id,
	)
	if err != nil {
		return domain.Feedback{}, false, err
	}
	count, err := result.RowsAffected()
	if err != nil {
		return domain.Feedback{}, false, err
	}
	if count == 0 {
		return domain.Feedback{}, false, nil
	}
	item, err := s.GetFeedback(id)
	return item, true, err
}

func (s *Store) GetFeedback(id int64) (domain.Feedback, error) {
	var item domain.Feedback
	err := s.db.QueryRow(
		`SELECT id,mod_id,category,category_number,title,body,author,game_version,mod_version,mod_list,save_link,status,created_at FROM feedback WHERE id = ?`,
		id,
	).Scan(&item.ID, &item.ModID, &item.Category, &item.CategoryNumber, &item.Title, &item.Body, &item.Author, &item.GameVersion, &item.ModVersion, &item.ModList, &item.SaveLink, &item.Status, &item.CreatedAt)
	return item, err
}

func (s *Store) SiteSettings() (domain.SiteSettings, error) {
	settings := domain.SiteSettings{ModVersion: "DEV BUILD", GameVersion: "RIMWORLD 1.6", Icon: "squirrel"}
	rows, err := s.db.Query(`SELECT key, value FROM site_settings`)
	if err != nil {
		return settings, err
	}
	defer rows.Close()
	for rows.Next() {
		var key, value string
		if err := rows.Scan(&key, &value); err != nil {
			return settings, err
		}
		switch key {
		case "mod_version":
			settings.ModVersion = value
		case "game_version":
			settings.GameVersion = value
		case "icon":
			settings.Icon = value
		}
	}
	return settings, rows.Err()
}

func (s *Store) UpdateSiteSettings(settings domain.SiteSettings) error {
	tx, err := s.db.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()
	pairs := [][2]string{
		{"mod_version", settings.ModVersion},
		{"game_version", settings.GameVersion},
		{"icon", settings.Icon},
	}
	for _, pair := range pairs {
		if _, err = tx.Exec(`INSERT INTO site_settings(key, value) VALUES(?, ?) ON CONFLICT(key) DO UPDATE SET value = excluded.value`, pair[0], pair[1]); err != nil {
			return err
		}
	}
	return tx.Commit()
}

func (s *Store) ListMods() ([]domain.Mod, error) {
	rows, err := s.db.Query(`SELECT id, slug, name, game_version, mod_version, icon, steam_url, github_url FROM mods ORDER BY sort_order, id`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	mods := make([]domain.Mod, 0)
	for rows.Next() {
		var item domain.Mod
		if err := rows.Scan(&item.ID, &item.Slug, &item.Name, &item.GameVersion, &item.ModVersion, &item.Icon, &item.SteamURL, &item.GitHubURL); err != nil {
			return nil, err
		}
		mods = append(mods, item)
	}
	return mods, rows.Err()
}

func (s *Store) ModBySlug(slug string) (domain.Mod, error) {
	var item domain.Mod
	err := s.db.QueryRow(
		`SELECT id, slug, name, game_version, mod_version, icon, steam_url, github_url FROM mods WHERE slug = ?`,
		slug,
	).Scan(&item.ID, &item.Slug, &item.Name, &item.GameVersion, &item.ModVersion, &item.Icon, &item.SteamURL, &item.GitHubURL)
	return item, err
}

func (s *Store) CreateMod(input domain.ModInput) (domain.Mod, error) {
	result, err := s.db.Exec(
		`INSERT INTO mods(slug, name, game_version, mod_version, icon, steam_url, github_url, sort_order) VALUES(?,?,?,?,?,?,?,(SELECT COALESCE(MAX(sort_order), 0) + 1 FROM mods))`,
		input.Slug, input.Name, input.GameVersion, input.ModVersion, input.Icon, input.SteamURL, input.GitHubURL,
	)
	if err != nil {
		return domain.Mod{}, err
	}
	id, err := result.LastInsertId()
	if err != nil {
		return domain.Mod{}, err
	}
	return domain.Mod{ID: id, Slug: input.Slug, Name: input.Name, GameVersion: input.GameVersion, ModVersion: input.ModVersion, Icon: input.Icon, SteamURL: input.SteamURL, GitHubURL: input.GitHubURL}, nil
}

func (s *Store) UpdateMod(id int64, input domain.ModInput) (domain.Mod, bool, error) {
	result, err := s.db.Exec(
		`UPDATE mods SET slug = ?, name = ?, game_version = ?, mod_version = ?, icon = ?, steam_url = ?, github_url = ? WHERE id = ?`,
		input.Slug, input.Name, input.GameVersion, input.ModVersion, input.Icon, input.SteamURL, input.GitHubURL, id,
	)
	if err != nil {
		return domain.Mod{}, false, err
	}
	count, err := result.RowsAffected()
	if err != nil {
		return domain.Mod{}, false, err
	}
	if count == 0 {
		return domain.Mod{}, false, nil
	}
	return domain.Mod{ID: id, Slug: input.Slug, Name: input.Name, GameVersion: input.GameVersion, ModVersion: input.ModVersion, Icon: input.Icon, SteamURL: input.SteamURL, GitHubURL: input.GitHubURL}, true, nil
}

func (s *Store) DeleteMod(id int64) (bool, error) {
	tx, err := s.db.Begin()
	if err != nil {
		return false, err
	}
	defer tx.Rollback()
	var exists int
	if err = tx.QueryRow(`SELECT COUNT(*) FROM mods WHERE id = ?`, id).Scan(&exists); err != nil {
		return false, err
	}
	if exists == 0 {
		return false, nil
	}
	var count int
	if err = tx.QueryRow(`SELECT COUNT(*) FROM mods`).Scan(&count); err != nil {
		return false, err
	}
	if count < 2 {
		return false, domain.ErrLastMod
	}
	if _, err = tx.Exec(`DELETE FROM feedback WHERE mod_id = ?`, id); err != nil {
		return false, err
	}
	result, err := tx.Exec(`DELETE FROM mods WHERE id = ?`, id)
	if err != nil {
		return false, err
	}
	affected, err := result.RowsAffected()
	if err != nil {
		return false, err
	}
	if affected == 0 {
		return false, nil
	}
	if err = tx.Commit(); err != nil {
		return false, err
	}
	return true, nil
}
