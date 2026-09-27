package api

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"modders-feedback/server/internal/auth"
	"modders-feedback/server/internal/domain"
	"modders-feedback/server/internal/store"
)

func TestRSSPublishesTextWithoutMedia(t *testing.T) {
	db, err := store.Open(t.TempDir())
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = db.Close() })
	if _, err = db.Register(domain.Registration{Username: "member", Password: "member-password"}); err != nil {
		t.Fatal(err)
	}
	mod, err := db.ModBySlug(domain.DefaultModSlug)
	if err != nil {
		t.Fatal(err)
	}
	created, err := db.CreateFeedback(domain.Feedback{
		ModID:    mod.ID,
		Category: domain.CategoryBug,
		Title:    "Colony food stops",
		Body:     "Steps live here.\n![shot](https://cdn.example/shot.png)\n<video src=\"/clips/repro.mp4\"></video>\n[save](https://files.example/colony.pdf)\nhttps://example.test/clip.webm\nStill readable.",
		Author:   "member",
	})
	if err != nil {
		t.Fatal(err)
	}

	handler := New(db, auth.New("test-secret-at-least-32-characters", false), t.TempDir())
	request := httptest.NewRequest(http.MethodGet, "/rss", nil)
	request.Host = "feedback.test"
	request.Header.Set("Accept-Language", "en")
	response := httptest.NewRecorder()
	handler.Routes().ServeHTTP(response, request)

	if response.Code != http.StatusOK {
		t.Fatalf("status = %d", response.Code)
	}
	if got := response.Header().Get("Content-Type"); !strings.Contains(got, "application/rss+xml") {
		t.Fatalf("content type = %q", got)
	}
	body := response.Body.String()
	link := "http://feedback.test" + domain.FeedbackPath(mod.Slug, created.Category, created.PublicID)
	for _, want := range []string{"鼠族：饥与祸-Bug report-Colony food stops", link, "Steps live here.", "Still readable."} {
		if !strings.Contains(body, want) {
			t.Fatalf("feed missing %q\n%s", want, body)
		}
	}
	for _, banned := range []string{"shot.png", "shot", "repro.mp4", "colony.pdf", "clip.webm", "<video", "</video>", "<img"} {
		if strings.Contains(body, banned) {
			t.Fatalf("feed published media %q\n%s", banned, body)
		}
	}
}

func TestRSSBodyStripsFileMarkup(t *testing.T) {
	body := rssBody("Keep this.\n![photo](https://cdn.example/a.png)\n[log](notes.log)\nSee https://files.example/demo.pdf later.")
	if strings.Contains(body, "photo") || strings.Contains(body, "notes.log") || strings.Contains(body, "demo.pdf") {
		t.Fatalf("media remained: %s", body)
	}
	if !strings.Contains(body, "Keep this.") || !strings.Contains(body, "See") || !strings.Contains(body, "later.") {
		t.Fatalf("text removed: %s", body)
	}
}
