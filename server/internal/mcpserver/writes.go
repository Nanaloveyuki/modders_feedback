package mcpserver

import (
	"context"
	"errors"

	"modders-feedback/server/internal/board"
	"modders-feedback/server/internal/domain"
)

type feedbackEditInput struct {
	feedbackRef
	board.FeedbackPatch
}

type statusInput struct {
	feedbackRef
	Status string `json:"status" jsonschema:"Configured status key from get_catalog"`
}

type modEditInput struct {
	ID int64 `json:"id" jsonschema:"Stable numeric mod ID from list_mods"`
	board.ModPatch
}

type modCreateInput struct {
	Slug        string `json:"slug"`
	Name        string `json:"name"`
	GameVersion string `json:"gameVersion"`
	ModVersion  string `json:"modVersion"`
	Icon        string `json:"icon"`
	SteamURL    string `json:"steamUrl,omitempty"`
	GitHubURL   string `json:"githubUrl,omitempty"`
}

func addAuthenticatedTool[In any](s *Server, name, description string, readOnly bool, handler func(context.Context, domain.User, In) (any, error)) {
	addTool(s, name, description, readOnly, func(ctx context.Context, in In) (any, error) {
		actor, err := actorFrom(ctx)
		if err != nil {
			return nil, err
		}
		return handler(ctx, actor, in)
	})
}

func (s *Server) registerWrites() {
	addAuthenticatedTool(s, "whoami", "Read the dedicated MCP token's current username and role. Requires Bearer authentication; website cookies are not accepted.", true, func(_ context.Context, actor domain.User, _ struct{}) (any, error) {
		return struct {
			Username string `json:"username"`
			Role     string `json:"role"`
		}{actor.Username, actor.Role}, nil
	})
	addAuthenticatedTool(s, "update_feedback", "Edit feedback by public UUID. Omit fields to preserve them; empty optional strings clear them. Members may edit only their own feedback; administrators may edit any. Requires Bearer authentication.", false, s.updateFeedback)
	addAuthenticatedTool(s, "set_feedback_status", "Set a configured status by public UUID. Members may switch only their own feedback when both current and target statuses allow author changes. Requires Bearer authentication.", false, s.setStatus)
	addAuthenticatedTool(s, "create_mod", "Create a mod with slug, name, gameVersion, modVersion, icon, and optional Steam/GitHub URLs. Administrator Bearer authentication required.", false, func(_ context.Context, actor domain.User, in modCreateInput) (any, error) {
		return s.board.CreateMod(actor, domain.ModInput{Slug: in.Slug, Name: in.Name, GameVersion: in.GameVersion, ModVersion: in.ModVersion, Icon: in.Icon, SteamURL: in.SteamURL, GitHubURL: in.GitHubURL})
	})
	addAuthenticatedTool(s, "update_mod", "Partially edit a mod by its stable numeric ID. Omitted fields stay unchanged; empty link strings clear URLs. Administrator Bearer authentication required.", false, func(_ context.Context, actor domain.User, in modEditInput) (any, error) {
		if in.ID < 1 {
			return nil, errors.New("id must be a positive mod ID")
		}
		return s.board.UpdateMod(actor, in.ID, in.ModPatch)
	})
}

func (s *Server) updateFeedback(_ context.Context, actor domain.User, in feedbackEditInput) (any, error) {
	item, err := s.findFeedback(in.PublicID)
	if err != nil {
		return nil, err
	}
	updated, err := s.board.UpdateFeedback(actor, item.ID, in.FeedbackPatch)
	if err != nil {
		return nil, err
	}
	return feedbackDetail{Feedback: updated, URL: feedbackURL(item)}, nil
}

func (s *Server) setStatus(_ context.Context, actor domain.User, in statusInput) (any, error) {
	item, err := s.findFeedback(in.PublicID)
	if err != nil {
		return nil, err
	}
	if err := s.board.SetStatus(actor, item.ID, in.Status); err != nil {
		return nil, err
	}
	item.Status = in.Status
	return feedbackDetail{Feedback: item, URL: feedbackURL(item)}, nil
}
