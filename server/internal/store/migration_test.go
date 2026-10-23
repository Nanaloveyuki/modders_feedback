package store

import (
	"fmt"
	"strings"
	"testing"

	"modders-feedback/server/internal/domain"
)

func TestReopeningPreservesFeedbackAttachmentsAndTimeline(t *testing.T) {
	for _, constrained := range []bool{false, true} {
		t.Run(fmt.Sprintf("legacy_status_constraint=%t", constrained), func(t *testing.T) {
			testReopeningPreservesFeedback(t, constrained)
		})
	}
}

func testReopeningPreservesFeedback(t *testing.T, constrained bool) {
	t.Helper()
	dir := t.TempDir()
	db, err := Open(dir)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() {
		if db != nil {
			_ = db.Close()
		}
	})
	if _, err := db.Register(domain.Registration{Username: "member", Password: "member-password"}); err != nil {
		t.Fatal(err)
	}
	mod, err := db.ModBySlug(domain.DefaultModSlug)
	if err != nil {
		t.Fatal(err)
	}
	item, err := db.CreateFeedback(domain.Feedback{ModID: mod.ID, Category: domain.CategoryBug, Title: "Persistent feedback", Body: "Feedback must survive restarts.", Author: "member"})
	if err != nil {
		t.Fatal(err)
	}
	attachment, err := db.SaveFeedbackAttachment(item.ID, "log.txt", "text/plain", []byte("preserved attachment"))
	if err != nil {
		t.Fatal(err)
	}
	if _, _, err := db.UpdateFeedback(item.ID, "member", domain.FeedbackUpdate{Title: item.Title, Body: item.Body + "\n[log](" + attachment.URL + ")"}); err != nil {
		t.Fatal(err)
	}
	if _, err := db.CreateComment(item.ID, "member", domain.CommentInput{Body: "Persistent comment"}); err != nil {
		t.Fatal(err)
	}
	if constrained {
		var definition string
		if err := db.db.QueryRow(`SELECT sql FROM sqlite_master WHERE name = 'feedback'`).Scan(&definition); err != nil {
			t.Fatal(err)
		}
		definition = strings.Replace(definition, "status TEXT NOT NULL DEFAULT 'open'", "status TEXT NOT NULL DEFAULT 'open' CHECK(status IN ('open','in_progress','testing','fixed_unreleased','resolved','closed','withdrawn'))", 1)
		if _, err := db.db.Exec(`PRAGMA writable_schema = ON`); err != nil {
			t.Fatal(err)
		}
		if _, err := db.db.Exec(`UPDATE sqlite_master SET sql = ? WHERE name = 'feedback'`, definition); err != nil {
			t.Fatal(err)
		}
		if _, err := db.db.Exec(`PRAGMA writable_schema = OFF`); err != nil {
			t.Fatal(err)
		}
	}
	if err := db.Close(); err != nil {
		t.Fatal(err)
	}
	db, err = Open(dir)
	if err != nil {
		t.Fatal(err)
	}
	got, err := db.GetFeedbackByPublicID(mod.ID, item.Category, item.PublicID)
	if err != nil {
		t.Fatal(err)
	}
	if got.Title != item.Title || len(got.Attachments) != 1 || got.Attachments[0].ID != attachment.ID {
		t.Fatalf("feedback after restart: %+v", got)
	}
	raw, _, err := db.OpenAttachment(attachment.ID)
	if err != nil || string(raw) != "preserved attachment" {
		t.Fatalf("attachment after restart: %q, %v", raw, err)
	}
	timeline, err := db.ListTimeline(item.ID)
	if err != nil {
		t.Fatal(err)
	}
	for _, event := range timeline.Events {
		if event.Body == "Persistent comment" {
			return
		}
	}
	t.Fatal("comment missing after restart")
}
