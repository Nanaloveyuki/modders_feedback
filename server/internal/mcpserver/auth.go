package mcpserver

import (
	"context"
	"crypto/sha256"
	"crypto/subtle"
	"errors"
	"net/http"
	"strings"

	"modders-feedback/server/internal/domain"
)

type actorKey struct{}

func (s *Server) authenticate(r *http.Request) (domain.User, error) {
	headers := r.Header.Values("Authorization")
	if len(headers) != 1 || len(s.tokenHash) == 0 {
		return domain.User{}, errors.New("Invalid MCP credentials")
	}
	fields := strings.Fields(headers[0])
	if len(fields) != 2 || !strings.EqualFold(fields[0], "Bearer") {
		return domain.User{}, errors.New("Invalid MCP credentials")
	}
	digest := sha256.Sum256([]byte(fields[1]))
	if subtle.ConstantTimeCompare(digest[:], s.tokenHash) != 1 {
		return domain.User{}, errors.New("Invalid MCP credentials")
	}
	actor, found, err := s.db.User(s.username)
	if err != nil || !found {
		return domain.User{}, errors.New("MCP account is unavailable")
	}
	return actor, nil
}

func actorFrom(ctx context.Context) (domain.User, error) {
	actor, ok := ctx.Value(actorKey{}).(domain.User)
	if !ok {
		return domain.User{}, errors.New("A dedicated MCP Bearer token is required")
	}
	return actor, nil
}

func privateTool(name string) bool {
	switch name {
	case "whoami", "update_feedback", "set_feedback_status", "create_mod", "update_mod":
		return true
	default:
		return false
	}
}

func unauthorized(w http.ResponseWriter) {
	w.Header().Set("WWW-Authenticate", `Bearer realm="modders-feedback-mcp"`)
	http.Error(w, "A valid dedicated MCP Bearer token is required", http.StatusUnauthorized)
}
