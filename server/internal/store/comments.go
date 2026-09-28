package store

import (
	"database/sql"
	"errors"
	"time"

	"modders-feedback/server/internal/domain"
)

func (s *Store) ensureComments() error {
	_, err := s.db.Exec(`
CREATE TABLE IF NOT EXISTS comments (
 id INTEGER PRIMARY KEY AUTOINCREMENT,
 feedback_id INTEGER NOT NULL REFERENCES feedback(id) ON DELETE CASCADE,
 author TEXT NOT NULL,
 body TEXT NOT NULL,
 reply_to INTEGER,
 created_at DATETIME NOT NULL,
 updated_at DATETIME NOT NULL
);
CREATE INDEX IF NOT EXISTS comments_feedback ON comments(feedback_id, created_at, id);
CREATE TABLE IF NOT EXISTS feedback_events (
 id INTEGER PRIMARY KEY AUTOINCREMENT,
 feedback_id INTEGER NOT NULL REFERENCES feedback(id) ON DELETE CASCADE,
 kind TEXT NOT NULL CHECK(kind IN ('comment','status','edited','comment_edited','deleted')),
 actor TEXT NOT NULL,
 body TEXT NOT NULL DEFAULT '',
 reply_to INTEGER,
 status TEXT NOT NULL DEFAULT '',
 comment_id INTEGER,
 created_at DATETIME NOT NULL
);
CREATE INDEX IF NOT EXISTS feedback_events_order ON feedback_events(feedback_id, created_at, id);`)
	return err
}

func (s *Store) ensureCommentID() error {
	var present int
	if err := s.db.QueryRow(`SELECT COUNT(*) FROM pragma_table_info('feedback_events') WHERE name = 'comment_id'`).Scan(&present); err != nil {
		return err
	}
	if present == 0 {
		if _, err := s.db.Exec(`ALTER TABLE feedback_events ADD COLUMN comment_id INTEGER`); err != nil {
			return err
		}
	}
	return nil
}

func (s *Store) CreateComment(feedbackID int64, author string, input domain.CommentInput) (domain.Comment, error) {
	reply := sql.NullInt64{}
	if input.ReplyTo > 0 {
		var parent int64
		err := s.db.QueryRow(`SELECT id FROM comments WHERE id = ? AND feedback_id = ?`, input.ReplyTo, feedbackID).Scan(&parent)
		if errors.Is(err, sql.ErrNoRows) {
			return domain.Comment{}, domain.ErrCommentTarget
		}
		if err != nil {
			return domain.Comment{}, err
		}
		reply = sql.NullInt64{Int64: parent, Valid: true}
	}
	now := time.Now().UTC()
	tx, err := s.db.Begin()
	if err != nil {
		return domain.Comment{}, err
	}
	defer tx.Rollback()
	result, err := tx.Exec(
		`INSERT INTO comments(feedback_id, author, body, reply_to, created_at, updated_at) VALUES(?,?,?,?,?,?)`,
		feedbackID, author, input.Body, reply, now, now,
	)
	if err != nil {
		return domain.Comment{}, err
	}
	id, err := result.LastInsertId()
	if err != nil {
		return domain.Comment{}, err
	}
	if err = s.bindDraftAttachments(tx, feedbackID, author, input.Body); err != nil {
		return domain.Comment{}, err
	}
	if err = insertEvent(tx, feedbackID, domain.EventComment, author, input.Body, reply, "", sql.NullInt64{Int64: id, Valid: true}); err != nil {
		return domain.Comment{}, err
	}
	if err = tx.Commit(); err != nil {
		return domain.Comment{}, err
	}
	return s.GetComment(id)
}

func (s *Store) UpdateComment(id int64, body string) (domain.Comment, bool, error) {
	current, err := s.GetComment(id)
	if errors.Is(err, sql.ErrNoRows) {
		return domain.Comment{}, false, nil
	}
	if err != nil {
		return domain.Comment{}, false, err
	}
	now := time.Now().UTC()
	tx, err := s.db.Begin()
	if err != nil {
		return domain.Comment{}, false, err
	}
	defer tx.Rollback()
	result, err := tx.Exec(`UPDATE comments SET body = ?, updated_at = ? WHERE id = ?`, body, now, id)
	if err != nil {
		return domain.Comment{}, false, err
	}
	count, err := result.RowsAffected()
	if err != nil {
		return domain.Comment{}, false, err
	}
	if count == 0 {
		return domain.Comment{}, false, nil
	}
	var reply sql.NullInt64
	if current.ReplyTo > 0 {
		reply = sql.NullInt64{Int64: current.ReplyTo, Valid: true}
	}
	if err = s.bindDraftAttachments(tx, current.FeedbackID, current.Author, body); err != nil {
		return domain.Comment{}, false, err
	}
	if err = insertEvent(tx, current.FeedbackID, domain.EventCommentEdited, current.Author, body, reply, "", sql.NullInt64{Int64: id, Valid: true}); err != nil {
		return domain.Comment{}, false, err
	}
	if err = tx.Commit(); err != nil {
		return domain.Comment{}, false, err
	}
	if err = s.PruneCommentAttachments(current.FeedbackID, id, body); err != nil {
		return domain.Comment{}, false, err
	}
	item, err := s.GetComment(id)
	return item, true, err
}

func (s *Store) DeleteComment(id int64, actor string) (bool, error) {
	current, err := s.GetComment(id)
	if errors.Is(err, sql.ErrNoRows) {
		return false, nil
	}
	if err != nil {
		return false, err
	}
	tx, err := s.db.Begin()
	if err != nil {
		return false, err
	}
	defer tx.Rollback()
	if _, err = tx.Exec(`UPDATE comments SET reply_to = NULL WHERE reply_to = ?`, id); err != nil {
		return false, err
	}
	result, err := tx.Exec(`DELETE FROM comments WHERE id = ?`, id)
	if err != nil {
		return false, err
	}
	count, err := result.RowsAffected()
	if err != nil {
		return false, err
	}
	if count == 0 {
		return false, nil
	}
	if err = insertEvent(tx, current.FeedbackID, domain.EventDeleted, actor, "", sql.NullInt64{}, "", sql.NullInt64{Int64: id, Valid: true}); err != nil {
		return false, err
	}
	if err = tx.Commit(); err != nil {
		return false, err
	}
	if err = s.PruneCommentAttachments(current.FeedbackID, 0, ""); err != nil {
		return false, err
	}
	return true, nil
}

func (s *Store) GetComment(id int64) (domain.Comment, error) {
	var item domain.Comment
	var reply sql.NullInt64
	err := s.db.QueryRow(
		`SELECT id, feedback_id, author, (SELECT CASE WHEN avatar IS NULL OR length(avatar) = 0 THEN '' ELSE '/api/account/avatar?u=' || username END FROM users WHERE username = comments.author), body, reply_to, created_at, updated_at FROM comments WHERE id = ?`,
		id,
	).Scan(&item.ID, &item.FeedbackID, &item.Author, &item.AuthorAvatar, &item.Body, &reply, &item.CreatedAt, &item.UpdatedAt)
	if reply.Valid {
		item.ReplyTo = reply.Int64
	}
	return item, err
}

func (s *Store) ListTimeline(feedbackID int64) (domain.Timeline, error) {
	rows, err := s.db.Query(
		`SELECT e.id, e.kind, e.actor, (SELECT CASE WHEN avatar IS NULL OR length(avatar) = 0 THEN '' ELSE '/api/account/avatar?u=' || username END FROM users WHERE username = e.actor), CASE WHEN e.kind = 'comment' AND c.id IS NOT NULL THEN c.body ELSE e.body END, CASE WHEN e.kind = 'comment' AND c.id IS NOT NULL THEN c.reply_to ELSE e.reply_to END, e.status, e.comment_id, e.created_at, c.id FROM feedback_events e LEFT JOIN comments c ON c.id = e.comment_id WHERE e.feedback_id = ? ORDER BY e.created_at, e.id`,
		feedbackID,
	)
	if err != nil {
		return domain.Timeline{}, err
	}
	defer rows.Close()
	events := []domain.TimelineEvent{}
	for rows.Next() {
		var item domain.TimelineEvent
		var reply, comment, live sql.NullInt64
		if err = rows.Scan(&item.ID, &item.Kind, &item.Actor, &item.ActorAvatar, &item.Body, &reply, &item.Status, &comment, &item.CreatedAt, &live); err != nil {
			return domain.Timeline{}, err
		}
		if item.Kind == domain.EventComment && !live.Valid {
			continue
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
		return domain.Timeline{}, err
	}
	return domain.Timeline{Events: events}, nil
}

func (s *Store) PruneCommentAttachments(feedbackID, commentID int64, body string) error {
	item, err := s.GetFeedback(feedbackID)
	if err != nil {
		return err
	}
	rows, err := s.db.Query(`SELECT id, body FROM comments WHERE feedback_id = ?`, feedbackID)
	if err != nil {
		return err
	}
	defer rows.Close()
	keepBody := item.Body + "\n" + body
	for rows.Next() {
		var id int64
		var text string
		if err = rows.Scan(&id, &text); err != nil {
			return err
		}
		if id == commentID {
			continue
		}
		keepBody += "\n" + text
	}
	if err = rows.Err(); err != nil {
		return err
	}
	return s.PruneUnreferenced(feedbackID, keepBody)
}

func (s *Store) deleteOrphanAttachments(tx *sql.Tx, feedbackID int64, bodies []string) error {
	joined := ""
	for _, body := range bodies {
		joined += "\n" + body
	}
	keep := map[string]struct{}{}
	for _, id := range attachmentIDs(joined) {
		keep[id] = struct{}{}
	}
	rows, err := tx.Query(`SELECT id FROM attachments WHERE feedback_id = ?`, feedbackID)
	if err != nil {
		return err
	}
	defer rows.Close()
	var drop []string
	for rows.Next() {
		var id string
		if err = rows.Scan(&id); err != nil {
			return err
		}
		if _, found := keep[id]; !found {
			drop = append(drop, id)
		}
	}
	if err = rows.Err(); err != nil {
		return err
	}
	for _, id := range drop {
		if err = s.files.Delete(id); err != nil {
			return err
		}
		if _, err = tx.Exec(`DELETE FROM attachments WHERE id = ?`, id); err != nil {
			return err
		}
	}
	return nil
}

func insertEvent(tx *sql.Tx, feedbackID int64, kind, actor, body string, reply sql.NullInt64, status string, commentID sql.NullInt64) error {
	_, err := tx.Exec(
		`INSERT INTO feedback_events(feedback_id, kind, actor, body, reply_to, status, comment_id, created_at) VALUES(?,?,?,?,?,?,?,?)`,
		feedbackID, kind, actor, body, reply, status, commentID, time.Now().UTC(),
	)
	return err
}
