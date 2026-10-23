package store

import (
	"database/sql"
	"errors"
	"fmt"
	"reflect"
	"testing"
	"time"

	"modders-feedback/server/internal/domain"
)

func openQueryTestStore(t *testing.T) (*Store, domain.Mod) {
	t.Helper()
	s, err := Open(t.TempDir())
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = s.Close() })
	if _, err := s.Register(domain.Registration{Username: "member", Password: "member-password"}); err != nil {
		t.Fatal(err)
	}
	mod, err := s.ModBySlug(domain.DefaultModSlug)
	if err != nil {
		t.Fatal(err)
	}
	return s, mod
}

func TestQueryFeedbackFiltersBeforePaginationAndLoadsIdentity(t *testing.T) {
	s, mod := openQueryTestStore(t)
	other, err := s.CreateMod(domain.ModInput{Slug: "other", Name: "Other mod", Icon: "squirrel"})
	if err != nil {
		t.Fatal(err)
	}
	var matching []domain.Feedback
	for i := range 10 {
		input := domain.Feedback{
			ModID: mod.ID, Category: domain.CategoryBug, Title: fmt.Sprintf("Feedback %d", i),
			Body: "Full body", Author: "member", GameVersion: "1.6", ModVersion: "2.0",
			ModList: "mod-a, mod-b", SaveLink: "https://example.com/save",
		}
		if i == 2 || i == 7 {
			input.ModID = other.ID
		}
		if i == 3 || i == 8 {
			input.Category = domain.CategoryQuestion
		}
		item, err := s.CreateFeedback(input)
		if err != nil {
			t.Fatal(err)
		}
		status := domain.StatusOpen
		if i == 1 || i == 6 {
			status = domain.StatusClosed
		}
		if _, err = s.db.Exec(`UPDATE feedback SET status = ?, created_at = ? WHERE id = ?`, status, time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC), item.ID); err != nil {
			t.Fatal(err)
		}
		if input.ModID == mod.ID && input.Category == domain.CategoryBug && status == domain.StatusOpen {
			matching = append(matching, item)
		}
	}
	attachment, err := s.SaveFeedbackAttachment(matching[0].ID, "log.txt", "text/plain", []byte("query attachment"))
	if err != nil {
		t.Fatal(err)
	}
	var gotIDs []int64
	for offset := 0; offset < len(matching); offset += 2 {
		page, err := s.QueryFeedback(FeedbackFilter{ModID: mod.ID, Category: domain.CategoryBug, Status: domain.StatusOpen, Limit: 2, Offset: offset})
		if err != nil {
			t.Fatal(err)
		}
		if len(page) != 2 {
			t.Fatalf("offset %d: got %d rows, want 2", offset, len(page))
		}
		for _, item := range page {
			gotIDs = append(gotIDs, item.ID)
			if item.ModSlug != mod.Slug || item.ModName != mod.Name || item.Category != domain.CategoryBug || item.Status != domain.StatusOpen {
				t.Fatalf("wrong identity or filters: %+v", item)
			}
			if item.Body != "Full body" || item.GameVersion != "1.6" || item.ModVersion != "2.0" || item.ModList != "mod-a, mod-b" || item.SaveLink != "https://example.com/save" || item.AuthorAvatar != "" {
				t.Fatalf("missing full feedback fields: %+v", item)
			}
			if item.Attachments == nil {
				t.Fatal("attachments must be a non-nil list")
			}
			if item.ID == matching[0].ID && (len(item.Attachments) != 1 || item.Attachments[0].ID != attachment.ID) {
				t.Fatalf("attachment not loaded: %+v", item.Attachments)
			}
		}
	}
	var wantIDs []int64
	for i := len(matching) - 1; i >= 0; i-- {
		wantIDs = append(wantIDs, matching[i].ID)
	}
	if !reflect.DeepEqual(gotIDs, wantIDs) {
		t.Fatalf("paged IDs = %v, want %v", gotIDs, wantIDs)
	}
	byUUID, err := s.FeedbackByUUID(matching[0].PublicID)
	if err != nil {
		t.Fatal(err)
	}
	if byUUID.ID != matching[0].ID || byUUID.Category != domain.CategoryBug || byUUID.ModID != mod.ID || byUUID.ModSlug != mod.Slug || byUUID.ModName != mod.Name || len(byUUID.Attachments) != 1 || byUUID.Attachments[0].ID != attachment.ID {
		t.Fatalf("UUID identity/attachments = %+v", byUUID)
	}
	byID, err := s.ModByID(mod.ID)
	if err != nil || byID != mod {
		t.Fatalf("mod by ID = %+v, %v; want %+v", byID, err, mod)
	}
	if _, err = s.FeedbackByUUID("missing"); !errors.Is(err, sql.ErrNoRows) {
		t.Fatalf("missing UUID error = %v", err)
	}
	if _, err = s.ModByID(-1); !errors.Is(err, sql.ErrNoRows) {
		t.Fatalf("missing mod error = %v", err)
	}
}

func TestTimelinePageFiltersDeletedCommentsBeforePagination(t *testing.T) {
	s, mod := openQueryTestStore(t)
	feedback, err := s.CreateFeedback(domain.Feedback{ModID: mod.ID, Category: domain.CategoryBug, Title: "Timeline", Body: "Body", Author: "member"})
	if err != nil {
		t.Fatal(err)
	}
	first, err := s.CreateComment(feedback.ID, "member", domain.CommentInput{Body: "Deleted body"})
	if err != nil {
		t.Fatal(err)
	}
	live, err := s.CreateComment(feedback.ID, "member", domain.CommentInput{Body: "Original body"})
	if err != nil {
		t.Fatal(err)
	}
	if _, err = s.db.Exec(`UPDATE comments SET body = ?, reply_to = ? WHERE id = ?`, "Current body", live.ID, live.ID); err != nil {
		t.Fatal(err)
	}
	if _, err = s.db.Exec(`DELETE FROM comments WHERE id = ?`, first.ID); err != nil {
		t.Fatal(err)
	}
	now := time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)
	if _, err = s.db.Exec(`INSERT INTO feedback_events(feedback_id, kind, actor, body, comment_id, created_at) VALUES(?, ?, ?, ?, ?, ?)`, feedback.ID, domain.EventCommentEdited, "missing-user", "Historical edit", first.ID, now); err != nil {
		t.Fatal(err)
	}
	if _, err = s.db.Exec(`INSERT INTO feedback_events(feedback_id, kind, actor, status, created_at) VALUES(?, ?, ?, ?, ?)`, feedback.ID, domain.EventStatus, "missing-user", domain.StatusClosed, now); err != nil {
		t.Fatal(err)
	}
	if _, err = s.db.Exec(`UPDATE feedback_events SET created_at = ? WHERE feedback_id = ?`, now, feedback.ID); err != nil {
		t.Fatal(err)
	}
	page, err := s.TimelinePage(feedback.ID, 2, 0)
	if err != nil {
		t.Fatal(err)
	}
	if len(page) != 2 || page[0].Kind != domain.EventComment || page[0].CommentID != live.ID || page[0].Body != "Current body" || page[0].ReplyTo != live.ID || page[0].ActorAvatar != "" {
		t.Fatalf("first page = %+v", page)
	}
	// Match ListTimeline: historical edit events remain visible after comment deletion.
	if page[1].Kind != domain.EventCommentEdited || page[1].Body != "Historical edit" {
		t.Fatalf("historical edit = %+v", page[1])
	}
	next, err := s.TimelinePage(feedback.ID, 2, 2)
	if err != nil || len(next) != 1 || next[0].Kind != domain.EventStatus || next[0].Status != domain.StatusClosed {
		t.Fatalf("second page = %+v, %v", next, err)
	}
	if page[0].ID >= page[1].ID || page[1].ID >= next[0].ID {
		t.Fatal("equal-time events must be ordered by ID")
	}
}

func TestReadQueriesBoundPages(t *testing.T) {
	s, mod := openQueryTestStore(t)
	now := time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)
	tx, err := s.db.Begin()
	if err != nil {
		t.Fatal(err)
	}
	defer tx.Rollback()
	var feedbackID int64
	for i := range 105 {
		result, err := tx.Exec(`INSERT INTO feedback(public_id, mod_id, category, category_number, title, body, author, created_at) VALUES(?, ?, ?, ?, ?, ?, ?, ?)`, fmt.Sprintf("00000000-0000-4000-8000-%012d", i), mod.ID, domain.CategoryBug, i+1, "Bounded", "Body", "missing-user", now)
		if err != nil {
			t.Fatal(err)
		}
		if i == 0 {
			feedbackID, err = result.LastInsertId()
			if err != nil {
				t.Fatal(err)
			}
		}
		if _, err = tx.Exec(`INSERT INTO feedback_events(feedback_id, kind, actor, status, created_at) VALUES(?, ?, ?, ?, ?)`, feedbackID, domain.EventStatus, "missing-user", domain.StatusOpen, now); err != nil {
			t.Fatal(err)
		}
	}
	if err = tx.Commit(); err != nil {
		t.Fatal(err)
	}
	for _, test := range []struct {
		limit  int
		offset int
		want   int
	}{
		{limit: 0, offset: -1, want: 20},
		{limit: -1, want: 20},
		{limit: 101, want: 20},
		{limit: 1, want: 1},
		{limit: 100, want: 100},
		{limit: 100, offset: 100, want: 5},
		{limit: 100, offset: 105, want: 0},
	} {
		feedback, err := s.QueryFeedback(FeedbackFilter{ModID: mod.ID, Limit: test.limit, Offset: test.offset})
		if err != nil || len(feedback) != test.want {
			t.Fatalf("feedback limit=%d offset=%d: got %d rows, %v; want %d", test.limit, test.offset, len(feedback), err, test.want)
		}
		events, err := s.TimelinePage(feedbackID, test.limit, test.offset)
		if err != nil || len(events) != test.want {
			t.Fatalf("timeline limit=%d offset=%d: got %d rows, %v; want %d", test.limit, test.offset, len(events), err, test.want)
		}
	}
}
