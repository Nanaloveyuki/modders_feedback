package store

import (
	"database/sql"

	"modders-feedback/server/internal/domain"
)

type FeedbackFilter struct {
	ModID    int64
	Category string
	Status   string
	Limit    int
	Offset   int
}

const feedbackQuerySelect = `SELECT
 f.id, f.public_id, f.mod_id, f.category, f.category_number, f.title, f.body, f.author,
 COALESCE((SELECT CASE WHEN avatar IS NULL OR length(avatar) = 0 THEN '' ELSE '/api/account/avatar?u=' || username END FROM users WHERE username = f.author), ''),
 f.game_version, f.mod_version, f.mod_list, f.save_link, f.status, f.created_at,
 m.slug, m.name
 FROM feedback f JOIN mods m ON m.id = f.mod_id`

type feedbackQueryScanner interface {
	Scan(...any) error
}

func scanFeedbackQuery(row feedbackQueryScanner) (domain.Feedback, error) {
	var item domain.Feedback
	err := row.Scan(
		&item.ID, &item.PublicID, &item.ModID, &item.Category, &item.CategoryNumber,
		&item.Title, &item.Body, &item.Author, &item.AuthorAvatar,
		&item.GameVersion, &item.ModVersion, &item.ModList, &item.SaveLink,
		&item.Status, &item.CreatedAt, &item.ModSlug, &item.ModName,
	)
	return item, err
}

func queryPageBounds(limit, offset int) (int, int) {
	if limit < 1 || limit > 100 {
		limit = 20
	}
	if offset < 0 {
		offset = 0
	}
	return limit, offset
}

func (s *Store) QueryFeedback(filter FeedbackFilter) ([]domain.Feedback, error) {
	limit, offset := queryPageBounds(filter.Limit, filter.Offset)
	query := feedbackQuerySelect + " WHERE f.mod_id = ?"
	args := make([]any, 1, 5)
	args[0] = filter.ModID
	if filter.Category != "" {
		query += " AND f.category = ?"
		args = append(args, filter.Category)
	}
	if filter.Status != "" {
		query += " AND f.status = ?"
		args = append(args, filter.Status)
	}
	query += " ORDER BY f.created_at DESC, f.id DESC LIMIT ? OFFSET ?"
	args = append(args, limit, offset)
	rows, err := s.db.Query(query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	items := make([]domain.Feedback, 0, limit)
	for rows.Next() {
		item, err := scanFeedbackQuery(rows)
		if err != nil {
			return nil, err
		}
		items = append(items, item)
	}
	if err = rows.Err(); err != nil {
		return nil, err
	}
	// Attachment queries need the store's single connection after this cursor closes.
	if err = rows.Close(); err != nil {
		return nil, err
	}
	if err = s.attach(items); err != nil {
		return nil, err
	}
	return items, nil
}

func (s *Store) FeedbackByUUID(publicID string) (domain.Feedback, error) {
	item, err := scanFeedbackQuery(s.db.QueryRow(feedbackQuerySelect+" WHERE f.public_id = ?", publicID))
	if err != nil {
		return item, err
	}
	if err = s.attachOne(&item); err != nil {
		return domain.Feedback{}, err
	}
	return item, nil
}

func (s *Store) ModByID(id int64) (domain.Mod, error) {
	var item domain.Mod
	err := s.db.QueryRow(
		`SELECT id, slug, name, game_version, mod_version, icon, steam_url, github_url FROM mods WHERE id = ?`,
		id,
	).Scan(&item.ID, &item.Slug, &item.Name, &item.GameVersion, &item.ModVersion, &item.Icon, &item.SteamURL, &item.GitHubURL)
	return item, err
}

func (s *Store) TimelinePage(feedbackID int64, limit, offset int) ([]domain.TimelineEvent, error) {
	limit, offset = queryPageBounds(limit, offset)
	rows, err := s.db.Query(
		`SELECT e.id, e.kind, e.actor,
 COALESCE((SELECT CASE WHEN avatar IS NULL OR length(avatar) = 0 THEN '' ELSE '/api/account/avatar?u=' || username END FROM users WHERE username = e.actor), ''),
 CASE WHEN e.kind = 'comment' AND c.id IS NOT NULL THEN c.body ELSE e.body END,
 CASE WHEN e.kind = 'comment' AND c.id IS NOT NULL THEN c.reply_to ELSE e.reply_to END,
 e.status, e.comment_id, e.created_at
 FROM feedback_events e LEFT JOIN comments c ON c.id = e.comment_id
 WHERE e.feedback_id = ? AND (e.kind <> 'comment' OR c.id IS NOT NULL)
 ORDER BY e.created_at, e.id LIMIT ? OFFSET ?`,
		feedbackID, limit, offset,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	events := make([]domain.TimelineEvent, 0, limit)
	for rows.Next() {
		var item domain.TimelineEvent
		var reply, comment sql.NullInt64
		if err = rows.Scan(&item.ID, &item.Kind, &item.Actor, &item.ActorAvatar, &item.Body, &reply, &item.Status, &comment, &item.CreatedAt); err != nil {
			return nil, err
		}
		if reply.Valid {
			item.ReplyTo = reply.Int64
		}
		if comment.Valid {
			item.CommentID = comment.Int64
		}
		events = append(events, item)
	}
	if err = rows.Err(); err != nil {
		return nil, err
	}
	return events, nil
}
