package mcpserver

import (
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/modelcontextprotocol/go-sdk/mcp"
	"modders-feedback/server/internal/board"
	"modders-feedback/server/internal/store"
)

const maxRequestBytes = 1 << 20

type Server struct {
	db        *store.Store
	protocol  *mcp.Server
	board     *board.Service
	username  string
	tokenHash []byte
}

func New(db *store.Store, username, tokenSHA256 string) (*Server, error) {
	var digest []byte
	if username != "" || tokenSHA256 != "" {
		var err error
		digest, err = hex.DecodeString(tokenSHA256)
		if username == "" || err != nil || len(digest) != sha256.Size {
			return nil, errors.New("MCP_USERNAME and a 64-character hexadecimal MCP_TOKEN_SHA256 must be configured together")
		}
	}
	s := &Server{db: db, board: board.New(db), username: username, tokenHash: digest, protocol: mcp.NewServer(&mcp.Implementation{Name: "modders-feedback", Version: "1.0.0"}, nil)}
	s.registerReads()
	s.registerWrites()
	return s, nil
}

func (s *Server) Handler() http.Handler {
	transport := mcp.NewStreamableHTTPHandler(func(*http.Request) *mcp.Server { return s.protocol }, &mcp.StreamableHTTPOptions{Stateless: true, JSONResponse: true})
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Cache-Control", "no-store")
		if origin := r.Header.Get("Origin"); origin != "" {
			u, err := url.Parse(origin)
			if err != nil || u.User != nil || u.Host == "" || !strings.EqualFold(u.Host, r.Host) || (u.Scheme != "http" && u.Scheme != "https") || u.Path != "" || u.RawQuery != "" || u.Fragment != "" {
				http.Error(w, "Origin is not allowed", http.StatusForbidden)
				return
			}
		}
		ctx := r.Context()
		if len(r.Header.Values("Authorization")) > 0 {
			actor, err := s.authenticate(r)
			if err != nil {
				unauthorized(w)
				return
			}
			ctx = context.WithValue(ctx, actorKey{}, actor)
		}
		if r.Method == http.MethodPost {
			r.Body = http.MaxBytesReader(w, r.Body, maxRequestBytes)
			body, err := io.ReadAll(r.Body)
			if err != nil {
				http.Error(w, "Request body is too large or unreadable", http.StatusRequestEntityTooLarge)
				return
			}
			if len(bytes.TrimSpace(body)) == 0 || bytes.TrimSpace(body)[0] != '{' {
				http.Error(w, "A single JSON-RPC object is required", http.StatusBadRequest)
				return
			}
			method, name, err := decodeEnvelope(body)
			if err != nil {
				http.Error(w, "Invalid JSON-RPC request", http.StatusBadRequest)
				return
			}
			if method == "tools/call" && privateTool(name) {
				if _, err := actorFrom(ctx); err != nil {
					unauthorized(w)
					return
				}
			}
			r.Body = io.NopCloser(bytes.NewReader(body))
		}
		ctx, cancel := context.WithTimeout(ctx, 15*time.Second)
		defer cancel()
		transport.ServeHTTP(w, r.WithContext(ctx))
	})
}

func addTool[In any](s *Server, name, description string, readOnly bool, handler func(context.Context, In) (any, error)) {
	closed := false
	mcp.AddTool(s.protocol, &mcp.Tool{Name: name, Description: description, Annotations: &mcp.ToolAnnotations{ReadOnlyHint: readOnly, OpenWorldHint: &closed}}, func(ctx context.Context, _ *mcp.CallToolRequest, in In) (*mcp.CallToolResult, any, error) {
		result, err := handler(ctx, in)
		if err != nil {
			return nil, nil, err
		}
		return nil, result, nil
	})
}

func publicError(err error) error {
	if err == nil {
		return nil
	}
	return errors.New("Could not read feedback board data")
}

func decodeEnvelope(body []byte) (string, string, error) {
	var envelope struct {
		Method string `json:"method"`
		Params struct {
			Name string `json:"name"`
		} `json:"params"`
	}
	err := json.Unmarshal(body, &envelope)
	return envelope.Method, envelope.Params.Name, err
}
