package board

import (
	"database/sql"
	"errors"
	"net/http"

	"modders-feedback/server/internal/domain"
)

type FeedbackPatch struct {
	Title       *string `json:"title,omitempty"`
	Body        *string `json:"body,omitempty"`
	GameVersion *string `json:"gameVersion,omitempty"`
	ModVersion  *string `json:"modVersion,omitempty"`
	ModList     *string `json:"modList,omitempty"`
	SaveLink    *string `json:"saveLink,omitempty"`
}

func (s *Service) UpdateFeedback(actor domain.User, id int64, patch FeedbackPatch) (domain.Feedback, error) {
	if patch.Title == nil && patch.Body == nil && patch.GameVersion == nil && patch.ModVersion == nil && patch.ModList == nil && patch.SaveLink == nil {
		return domain.Feedback{}, &Error{http.StatusBadRequest, "请至少提供一个可修改字段", "Provide at least one editable field"}
	}
	input := domain.FeedbackUpdate{
		Title:       patchedString("", patch.Title),
		Body:        patchedString("", patch.Body),
		GameVersion: patchedString("", patch.GameVersion),
		ModVersion:  patchedString("", patch.ModVersion),
		ModList:     patchedString("", patch.ModList),
		SaveLink:    patchedString("", patch.SaveLink),
	}
	if (patch.Title != nil && !withinRunes(input.Title, 5, 120)) || (patch.Body != nil && !withinRunes(input.Body, 10, 12000)) {
		return domain.Feedback{}, &Error{http.StatusBadRequest, "请检查标题（5-120 字）和描述（10-12000 字）", "Check the title (5-120 characters) and description (10-12000 characters)"}
	}
	if len(input.GameVersion) > 40 || len(input.ModVersion) > 80 || len(input.ModList) > 6000 || len(input.SaveLink) > 500 {
		return domain.Feedback{}, &Error{http.StatusBadRequest, "版本、模组列表或存档链接超出长度限制", "Version, mod list, or save link is too long"}
	}
	current, err := s.db.GetFeedback(id)
	if errors.Is(err, sql.ErrNoRows) {
		return domain.Feedback{}, &Error{http.StatusNotFound, "未找到该反馈", "Feedback was not found"}
	}
	if err != nil {
		return domain.Feedback{}, &Error{http.StatusInternalServerError, "更新反馈失败", "Could not update feedback"}
	}
	if actor.Role != "admin" && actor.Username != current.Author {
		return domain.Feedback{}, &Error{http.StatusForbidden, "只能编辑自己的反馈", "You can only edit your own feedback"}
	}
	if patch.Title == nil {
		input.Title = current.Title
	}
	if patch.Body == nil {
		input.Body = current.Body
	}
	if patch.GameVersion == nil {
		input.GameVersion = current.GameVersion
	}
	if patch.ModVersion == nil {
		input.ModVersion = current.ModVersion
	}
	if patch.ModList == nil {
		input.ModList = current.ModList
	}
	if patch.SaveLink == nil {
		input.SaveLink = current.SaveLink
	}
	item, updated, err := s.db.UpdateFeedback(id, actor.Username, input)
	if err != nil {
		return domain.Feedback{}, &Error{http.StatusInternalServerError, "更新反馈失败", "Could not update feedback"}
	}
	if !updated {
		return domain.Feedback{}, &Error{http.StatusNotFound, "未找到该反馈", "Feedback was not found"}
	}
	return item, nil
}

func (s *Service) SetStatus(actor domain.User, id int64, status string) error {
	settings, err := s.db.SiteSettings()
	if err != nil {
		return &Error{http.StatusInternalServerError, "更新反馈状态失败", "Could not update feedback status"}
	}
	next, ok := domain.StatusByKey(settings.Statuses, status)
	if !ok {
		return &Error{http.StatusBadRequest, "反馈状态无效", "Invalid feedback status"}
	}
	item, err := s.db.GetFeedback(id)
	if errors.Is(err, sql.ErrNoRows) {
		return &Error{http.StatusNotFound, "未找到该反馈", "Feedback was not found"}
	}
	if err != nil {
		return &Error{http.StatusInternalServerError, "更新反馈状态失败", "Could not update feedback status"}
	}
	current, known := domain.StatusByKey(settings.Statuses, item.Status)
	if actor.Role != "admin" && (actor.Username != item.Author || !known || !current.Author || !next.Author) {
		return &Error{http.StatusForbidden, "只能在自己的反馈中切换允许的状态", "You can only switch your own feedback between allowed statuses"}
	}
	updated, err := s.db.UpdateStatus(id, actor.Username, status)
	if err != nil {
		return &Error{http.StatusInternalServerError, "更新反馈状态失败", "Could not update feedback status"}
	}
	if !updated {
		return &Error{http.StatusNotFound, "未找到该反馈", "Feedback was not found"}
	}
	return nil
}
