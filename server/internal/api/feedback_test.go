package api

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strconv"
	"strings"
	"testing"

	"modders-feedback/server/internal/auth"
	"modders-feedback/server/internal/domain"
	"modders-feedback/server/internal/store"
)

func TestFeedbackEditingRequiresAuthorOrAdministrator(t *testing.T) {
	db, err := store.Open(t.TempDir())
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = db.Close() })
	sessions := auth.New("test-secret-at-least-32-characters", false)
	routes := New(db, sessions, t.TempDir()).Routes()
	request := func(method, path, body string, cookie *http.Cookie) *httptest.ResponseRecorder {
		t.Helper()
		r := httptest.NewRequest(method, path, strings.NewReader(body))
		r.Header.Set("Content-Type", "application/json")
		if cookie != nil {
			r.AddCookie(cookie)
		}
		w := httptest.NewRecorder()
		routes.ServeHTTP(w, r)
		return w
	}
	register := func(username string) *http.Cookie {
		t.Helper()
		w := request(http.MethodPost, "/api/auth/register", `{"username":"`+username+`","password":"member-password"}`, nil)
		if w.Code != http.StatusCreated {
			t.Fatalf("register %s: status %d: %s", username, w.Code, w.Body.String())
		}
		return w.Result().Cookies()[0]
	}
	author := register("author")
	other := register("other")
	created := request(http.MethodPost, "/api/feedback", `{"category":"bug","title":"Original report","body":"Original feedback body."}`, author)
	if created.Code != http.StatusCreated {
		t.Fatalf("create: status %d: %s", created.Code, created.Body.String())
	}
	var item domain.Feedback
	if err := json.Unmarshal(created.Body.Bytes(), &item); err != nil {
		t.Fatal(err)
	}
	path := "/api/feedback/" + strconv.FormatInt(item.ID, 10)
	update := `{"title":"Edited report","body":"Edited feedback body.","gameVersion":"1.6","modVersion":"updated","modList":"Core","saveLink":""}`
	for _, check := range []struct {
		name   string
		cookie *http.Cookie
		want   int
	}{
		{"anonymous", nil, http.StatusUnauthorized},
		{"another member", other, http.StatusForbidden},
		{"author", author, http.StatusOK},
	} {
		w := request(http.MethodPatch, path, update, check.cookie)
		if w.Code != check.want {
			t.Fatalf("%s edit: status %d, want %d: %s", check.name, w.Code, check.want, w.Body.String())
		}
	}
	public := request(http.MethodGet, "/api/mods/rhah/feedback/bugs/"+item.PublicID, "", nil)
	var edited domain.Feedback
	if err := json.Unmarshal(public.Body.Bytes(), &edited); err != nil {
		t.Fatal(err)
	}
	if public.Code != http.StatusOK || edited.Title != "Edited report" || edited.Body != "Edited feedback body." || edited.ModVersion != "updated" {
		t.Fatalf("public read after editing: status %d: %+v", public.Code, edited)
	}
	if err := db.SeedAdmin("admin", "administrator-password"); err != nil {
		t.Fatal(err)
	}
	admin := httptest.NewRecorder()
	if err := sessions.Issue(admin, "admin"); err != nil {
		t.Fatal(err)
	}
	w := request(http.MethodPatch, path, strings.Replace(update, "Edited report", "Administrator revision", 1), admin.Result().Cookies()[0])
	if w.Code != http.StatusOK {
		t.Fatalf("administrator edit: status %d: %s", w.Code, w.Body.String())
	}
	public = request(http.MethodGet, "/api/mods/rhah/feedback/bugs/"+item.PublicID, "", nil)
	if err := json.Unmarshal(public.Body.Bytes(), &edited); err != nil {
		t.Fatal(err)
	}
	if public.Code != http.StatusOK || edited.Title != "Administrator revision" {
		t.Fatalf("public read after administrator editing: status %d: %+v", public.Code, edited)
	}
}
