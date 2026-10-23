package mcpserver

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/modelcontextprotocol/go-sdk/mcp"
	"modders-feedback/server/internal/domain"
	"modders-feedback/server/internal/store"
)

type tokenTransport struct{ token string }

func (t tokenTransport) RoundTrip(r *http.Request) (*http.Response, error) {
	r = r.Clone(r.Context())
	if t.token != "" {
		r.Header.Set("Authorization", "Bearer "+t.token)
	}
	return http.DefaultTransport.RoundTrip(r)
}

func digest(token string) string {
	hash := sha256.Sum256([]byte(token))
	return hex.EncodeToString(hash[:])
}

func connect(t *testing.T, endpoint, token string) *mcp.ClientSession {
	t.Helper()
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	session, err := mcp.NewClient(&mcp.Implementation{Name: "integration", Version: "1"}, nil).Connect(ctx, &mcp.StreamableClientTransport{Endpoint: endpoint, DisableStandaloneSSE: true, HTTPClient: &http.Client{Transport: tokenTransport{token}}}, nil)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = session.Close() })
	return session
}

func call(t *testing.T, session *mcp.ClientSession, name string, args any, wantError bool) map[string]any {
	t.Helper()
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	result, err := session.CallTool(ctx, &mcp.CallToolParams{Name: name, Arguments: args})
	if wantError {
		if err == nil && !result.IsError {
			t.Fatalf("%s unexpectedly succeeded: %+v", name, result)
		}
		return nil
	}
	if err != nil {
		t.Fatalf("%s: %v", name, err)
	}
	if result.IsError {
		raw, _ := json.Marshal(result.Content)
		t.Fatalf("%s: %s", name, raw)
	}
	raw, err := json.Marshal(result.StructuredContent)
	if err != nil {
		t.Fatal(err)
	}
	var out map[string]any
	if err = json.Unmarshal(raw, &out); err != nil {
		t.Fatal(err)
	}
	return out
}

func fixture(t *testing.T) (*store.Store, domain.Feedback, domain.Feedback) {
	t.Helper()
	db, err := store.Open(t.TempDir())
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = db.Close() })
	for _, username := range []string{"author", "other"} {
		if _, err = db.Register(domain.Registration{Username: username, Password: "member-password"}); err != nil {
			t.Fatal(err)
		}
	}
	if err = db.SeedAdmin("admin", "administrator-password"); err != nil {
		t.Fatal(err)
	}
	mod, err := db.ModBySlug("rhah")
	if err != nil {
		t.Fatal(err)
	}
	attachment, err := db.SaveDraftAttachment("author", "report.txt", "text/plain", []byte("attached report"))
	if err != nil {
		t.Fatal(err)
	}
	own, err := db.CreateFeedback(domain.Feedback{ModID: mod.ID, Category: "bug", Title: "Original report", Body: "Original report body. [report](" + attachment.URL + ")", Author: "author", GameVersion: "1.6", ModVersion: "old", ModList: "Core", SaveLink: "https://example.com/save"})
	if err != nil {
		t.Fatal(err)
	}
	other, err := db.CreateFeedback(domain.Feedback{ModID: mod.ID, Category: "question", Title: "Other question", Body: "Another member question body.", Author: "other"})
	if err != nil {
		t.Fatal(err)
	}
	return db, own, other
}

func serve(t *testing.T, db *store.Store, username, token string) *httptest.Server {
	t.Helper()
	hash := ""
	if token != "" {
		hash = digest(token)
	}
	server, err := New(db, username, hash)
	if err != nil {
		t.Fatal(err)
	}
	httpServer := httptest.NewServer(server.Handler())
	t.Cleanup(httpServer.Close)
	return httpServer
}

func TestMCPMemberPermissionsAndPartialEdits(t *testing.T) {
	db, own, other := fixture(t)
	httpServer := serve(t, db, "author", "member-token")
	client := connect(t, httpServer.URL, "member-token")
	identity := call(t, client, "whoami", map[string]any{}, false)
	if identity["username"] != "author" || identity["role"] != "member" || len(identity) != 2 {
		t.Fatalf("identity exposed unexpected account data: %+v", identity)
	}
	call(t, client, "update_feedback", map[string]any{"publicId": own.PublicID, "title": "Updated report", "saveLink": ""}, false)
	updated, err := db.GetFeedback(own.ID)
	if err != nil {
		t.Fatal(err)
	}
	if updated.Title != "Updated report" || updated.Body != own.Body || updated.GameVersion != own.GameVersion || updated.ModVersion != own.ModVersion || updated.ModList != own.ModList || updated.SaveLink != "" || len(updated.Attachments) != 1 || updated.Attachments[0].ID != own.Attachments[0].ID {
		t.Fatalf("partial edit lost feedback fields or attachment: %+v", updated)
	}
	call(t, client, "update_feedback", map[string]any{"publicId": other.PublicID, "title": "Stolen report"}, true)
	call(t, client, "update_feedback", map[string]any{"publicId": own.PublicID}, true)
	call(t, client, "update_feedback", map[string]any{"publicId": own.PublicID, "body": ""}, true)
	call(t, client, "update_feedback", map[string]any{"publicId": own.PublicID, "title": "Unexpected write", "unexpected": true}, true)
	updated, err = db.GetFeedback(own.ID)
	if err != nil {
		t.Fatal(err)
	}
	if updated.Title != "Updated report" || updated.Body != own.Body {
		t.Fatal("invalid edit changed feedback")
	}
	call(t, client, "set_feedback_status", map[string]any{"publicId": own.PublicID, "status": "withdrawn"}, false)
	call(t, client, "set_feedback_status", map[string]any{"publicId": own.PublicID, "status": "open"}, false)
	call(t, client, "set_feedback_status", map[string]any{"publicId": other.PublicID, "status": "withdrawn"}, true)
	call(t, client, "set_feedback_status", map[string]any{"publicId": own.PublicID, "status": "testing"}, true)
	if _, err = db.UpdateStatus(own.ID, "admin", "testing"); err != nil {
		t.Fatal(err)
	}
	call(t, client, "set_feedback_status", map[string]any{"publicId": own.PublicID, "status": "open"}, true)
	settings, err := db.SiteSettings()
	if err != nil {
		t.Fatal(err)
	}
	settings.Statuses = append(settings.Statuses, domain.StatusLabel{Key: "member_review", LabelZh: "复核", LabelEn: "Review", Light: "#123456", Dark: "#abcdef", Author: true})
	if err = db.ReplaceStatuses(settings.Statuses); err != nil {
		t.Fatal(err)
	}
	if _, err = db.UpdateStatus(own.ID, "admin", "open"); err != nil {
		t.Fatal(err)
	}
	call(t, client, "set_feedback_status", map[string]any{"publicId": own.PublicID, "status": "member_review"}, false)
	call(t, client, "create_mod", map[string]any{"slug": "denied", "name": "Denied", "gameVersion": "1.6", "modVersion": "1", "icon": "rat"}, true)
	call(t, client, "update_mod", map[string]any{"id": own.ModID, "name": "Denied"}, true)
}

func TestMCPAdminModManagementAndLiveRole(t *testing.T) {
	db, own, other := fixture(t)
	httpServer := serve(t, db, "admin", "admin-token")
	client := connect(t, httpServer.URL, "admin-token")
	call(t, client, "update_feedback", map[string]any{"publicId": other.PublicID, "title": "Admin revision"}, false)
	call(t, client, "set_feedback_status", map[string]any{"publicId": own.PublicID, "status": "testing"}, false)
	created := call(t, client, "create_mod", map[string]any{"slug": "new-mod", "name": "New mod", "gameVersion": "1.6", "modVersion": "1", "icon": "rat", "steamUrl": "https://example.com/steam", "githubUrl": "https://example.com/repo"}, false)
	id := created["id"]
	call(t, client, "update_mod", map[string]any{"id": id, "name": "Renamed mod", "slug": "renamed-mod", "githubUrl": ""}, false)
	mod, err := db.ModBySlug("renamed-mod")
	if err != nil {
		t.Fatal(err)
	}
	if mod.Name != "Renamed mod" || mod.GameVersion != "1.6" || mod.ModVersion != "1" || mod.Icon != "rat" || mod.GitHubURL != "" || mod.SteamURL != "https://example.com/steam" || float64(mod.ID) != id {
		t.Fatalf("mod patch lost data: %+v", mod)
	}
	call(t, client, "update_mod", map[string]any{"id": id, "steamUrl": "javascript:alert(1)"}, true)
	call(t, client, "create_mod", map[string]any{"slug": "minimal", "name": "Minimal", "gameVersion": "1.6", "modVersion": "1", "icon": "paw"}, false)
	// Promote the configured member while retaining the same HTTP client. Role must be re-read.
	memberServer := serve(t, db, "author", "member-token")
	member := connect(t, memberServer.URL, "member-token")
	call(t, member, "update_mod", map[string]any{"id": id, "name": "Denied"}, true)
	if err = db.SeedAdmin("author", "administrator-password"); err != nil {
		t.Fatal(err)
	}
	identity := call(t, member, "whoami", map[string]any{}, false)
	if identity["role"] != "admin" {
		t.Fatalf("cached stale role: %+v", identity)
	}
	call(t, member, "update_mod", map[string]any{"id": id, "name": "Promoted owner"}, false)
}

func TestMCPAnonymousReadsAndCredentialRejection(t *testing.T) {
	db, own, _ := fixture(t)
	server := serve(t, db, "author", "valid-token")
	client := connect(t, server.URL, "")
	for _, in := range []struct {
		name string
		args map[string]any
	}{
		{"list_mods", map[string]any{"query": "RHAH", "limit": 1}},
		{"list_feedback", map[string]any{"mod": "rhah", "category": "bug", "status": "open", "limit": 1}},
		{"get_feedback", map[string]any{"publicId": own.PublicID}},
		{"list_feedback_timeline", map[string]any{"publicId": own.PublicID}},
		{"get_catalog", map[string]any{}},
	} {
		call(t, client, in.name, in.args, false)
	}
	for _, args := range []map[string]any{{"mod": "rhah", "limit": 101}, {"mod": "rhah", "offset": -1}, {"mod": "rhah", "category": "bugs"}, {"mod": "rhah", "status": "unknown"}} {
		call(t, client, "list_feedback", args, true)
	}
	body := fmt.Sprintf(`{"jsonrpc":"2.0","id":1,"method":"tools/call","params":{"name":"update_feedback","arguments":{"publicId":%q,"title":"Anonymous edit"}}}`, own.PublicID)
	for _, header := range []string{"", "Bearer invalid-token", "Basic valid-token", "Bearer"} {
		request, _ := http.NewRequest("POST", server.URL, strings.NewReader(body))
		request.Header.Set("Content-Type", "application/json")
		request.Header.Set("Accept", "application/json, text/event-stream")
		if header != "" {
			request.Header.Set("Authorization", header)
		}
		response, err := http.DefaultClient.Do(request)
		if err != nil {
			t.Fatal(err)
		}
		response.Body.Close()
		if response.StatusCode != 401 {
			t.Fatalf("credentials %q: %d, want 401", header, response.StatusCode)
		}
	}
	for _, token := range []string{"invalid-token", "valid-token"} {
		endpoint := server.URL
		if token == "valid-token" {
			endpoint = serve(t, db, "author", "rotated-token").URL
		}
		request, _ := http.NewRequest("POST", endpoint, strings.NewReader(`{"jsonrpc":"2.0","id":1,"method":"tools/list"}`))
		request.Header.Set("Authorization", "Bearer "+token)
		response, err := http.DefaultClient.Do(request)
		if err != nil {
			t.Fatal(err)
		}
		response.Body.Close()
		if response.StatusCode != 401 {
			t.Fatalf("invalid explicit token fell back to public reads: %d", response.StatusCode)
		}
	}
	rotated := serve(t, db, "author", "rotated-token")
	call(t, connect(t, rotated.URL, "rotated-token"), "whoami", map[string]any{}, false)
	request, _ := http.NewRequest("POST", server.URL, strings.NewReader(`{"jsonrpc":"2.0","id":1,"method":"tools/list"}`))
	request.Header.Set("Origin", "https://untrusted.example")
	response, err := http.DefaultClient.Do(request)
	if err != nil {
		t.Fatal(err)
	}
	response.Body.Close()
	if response.StatusCode != 403 {
		t.Fatalf("cross-origin request: %d", response.StatusCode)
	}
	readonly := serve(t, db, "", "")
	call(t, connect(t, readonly.URL, ""), "get_catalog", map[string]any{}, false)
	call(t, connect(t, readonly.URL, ""), "whoami", map[string]any{}, true)
}
