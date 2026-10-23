package board

import (
	"database/sql"
	"errors"
	"net/http"
	"strings"

	"modders-feedback/server/internal/domain"
)

type ModPatch struct {
	Slug        *string `json:"slug,omitempty"`
	Name        *string `json:"name,omitempty"`
	GameVersion *string `json:"gameVersion,omitempty"`
	ModVersion  *string `json:"modVersion,omitempty"`
	Icon        *string `json:"icon,omitempty"`
	SteamURL    *string `json:"steamUrl,omitempty"`
	GitHubURL   *string `json:"githubUrl,omitempty"`
}

func (s *Service) CreateMod(actor domain.User, input domain.ModInput) (domain.Mod, error) {
	if actor.Role != "admin" {
		return domain.Mod{}, &Error{http.StatusForbidden, "需要管理员权限", "Administrator access is required"}
	}
	input.Slug = strings.TrimSpace(input.Slug)
	input.Name = strings.TrimSpace(input.Name)
	input.GameVersion = strings.TrimSpace(input.GameVersion)
	input.ModVersion = strings.TrimSpace(input.ModVersion)
	input.Icon = strings.TrimSpace(input.Icon)
	input.SteamURL = strings.TrimSpace(input.SteamURL)
	input.GitHubURL = strings.TrimSpace(input.GitHubURL)
	if !domain.ValidSlug(input.Slug) || !withinRunes(input.Name, 1, 40) || !withinRunes(input.GameVersion, 1, 40) || !withinRunes(input.ModVersion, 1, 40) || !domain.ValidIcon(input.Icon) || !domain.ValidOptionalURL(input.SteamURL) || !domain.ValidOptionalURL(input.GitHubURL) {
		return domain.Mod{}, invalidMod()
	}
	created, err := s.db.CreateMod(input)
	if err != nil {
		return domain.Mod{}, &Error{http.StatusConflict, "无法添加模组，短标识可能已存在", "Could not add the mod; the slug may already exist"}
	}
	return created, nil
}

func (s *Service) UpdateMod(actor domain.User, id int64, patch ModPatch) (domain.Mod, error) {
	if actor.Role != "admin" {
		return domain.Mod{}, &Error{http.StatusForbidden, "需要管理员权限", "Administrator access is required"}
	}
	if patch.Slug == nil && patch.Name == nil && patch.GameVersion == nil && patch.ModVersion == nil && patch.Icon == nil && patch.SteamURL == nil && patch.GitHubURL == nil {
		return domain.Mod{}, &Error{http.StatusBadRequest, "请至少提供一个可修改字段", "Provide at least one editable field"}
	}
	input := domain.ModInput{
		Slug:        patchedString("", patch.Slug),
		Name:        patchedString("", patch.Name),
		GameVersion: patchedString("", patch.GameVersion),
		ModVersion:  patchedString("", patch.ModVersion),
		Icon:        patchedString("", patch.Icon),
		SteamURL:    patchedString("", patch.SteamURL),
		GitHubURL:   patchedString("", patch.GitHubURL),
	}
	if (patch.Slug != nil && !domain.ValidSlug(input.Slug)) ||
		(patch.Name != nil && !withinRunes(input.Name, 1, 40)) ||
		(patch.GameVersion != nil && !withinRunes(input.GameVersion, 1, 40)) ||
		(patch.ModVersion != nil && !withinRunes(input.ModVersion, 1, 40)) ||
		(patch.Icon != nil && !domain.ValidIcon(input.Icon)) ||
		!domain.ValidOptionalURL(input.SteamURL) || !domain.ValidOptionalURL(input.GitHubURL) {
		return domain.Mod{}, invalidMod()
	}
	current, err := s.db.ModByID(id)
	if errors.Is(err, sql.ErrNoRows) {
		return domain.Mod{}, &Error{http.StatusNotFound, "未找到该模组", "Mod was not found"}
	}
	if err != nil {
		return domain.Mod{}, &Error{http.StatusConflict, "无法保存模组，短标识可能已存在", "Could not save the mod; the slug may already exist"}
	}
	if patch.Slug == nil {
		input.Slug = current.Slug
	}
	if patch.Name == nil {
		input.Name = current.Name
	}
	if patch.GameVersion == nil {
		input.GameVersion = current.GameVersion
	}
	if patch.ModVersion == nil {
		input.ModVersion = current.ModVersion
	}
	if patch.Icon == nil {
		input.Icon = current.Icon
	}
	if patch.SteamURL == nil {
		input.SteamURL = current.SteamURL
	}
	if patch.GitHubURL == nil {
		input.GitHubURL = current.GitHubURL
	}
	updated, found, err := s.db.UpdateMod(id, input)
	if err != nil {
		return domain.Mod{}, &Error{http.StatusConflict, "无法保存模组，短标识可能已存在", "Could not save the mod; the slug may already exist"}
	}
	if !found {
		return domain.Mod{}, &Error{http.StatusNotFound, "未找到该模组", "Mod was not found"}
	}
	return updated, nil
}

func invalidMod() *Error {
	return &Error{http.StatusBadRequest, "请检查模组标识、名称（1-40 字）、版本、图标和链接", "Check the slug, name (1-40 characters), versions, icon, and links"}
}
