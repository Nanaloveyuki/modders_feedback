package api

import (
	"database/sql"
	"errors"
	"net/http"
	"strconv"
	"strings"

	"github.com/go-chi/chi/v5"
	"modders-feedback/server/internal/domain"
	"modders-feedback/server/internal/httpx"
)

func (h *Handler) listTimeline(w http.ResponseWriter, r *http.Request) {
	id, ok := feedbackID(w, r)
	if !ok {
		return
	}
	if _, err := h.store.GetFeedback(id); errors.Is(err, sql.ErrNoRows) {
		httpx.Error(w, http.StatusNotFound, httpx.Text(r, "未找到该反馈", "Feedback was not found"))
		return
	} else if err != nil {
		httpx.Error(w, http.StatusInternalServerError, httpx.Text(r, "读取评论失败", "Could not load comments"))
		return
	}
	timeline, err := h.store.ListTimeline(id)
	if err != nil {
		httpx.Error(w, http.StatusInternalServerError, httpx.Text(r, "读取评论失败", "Could not load comments"))
		return
	}
	httpx.JSON(w, http.StatusOK, timeline)
}

func (h *Handler) createComment(w http.ResponseWriter, r *http.Request) {
	user, ok := h.currentUser(w, r)
	if !ok {
		return
	}
	id, ok := feedbackID(w, r)
	if !ok {
		return
	}
	var input domain.CommentInput
	if !httpx.Decode(w, r, &input) {
		return
	}
	input.Body = strings.TrimSpace(input.Body)
	if !withinRunes(input.Body, 1, 8000) || input.ReplyTo < 0 {
		httpx.Error(w, http.StatusBadRequest, httpx.Text(r, "评论需要 1-8000 字", "A comment needs 1-8000 characters"))
		return
	}
	if _, err := h.store.GetFeedback(id); errors.Is(err, sql.ErrNoRows) {
		httpx.Error(w, http.StatusNotFound, httpx.Text(r, "未找到该反馈", "Feedback was not found"))
		return
	} else if err != nil {
		httpx.Error(w, http.StatusInternalServerError, httpx.Text(r, "发表评论失败", "Could not publish the comment"))
		return
	}
	created, err := h.store.CreateComment(id, user.Username, input)
	if errors.Is(err, domain.ErrCommentTarget) || errors.Is(err, domain.ErrAttachmentMissing) {
		httpx.Error(w, http.StatusBadRequest, httpx.Text(r, "回复目标或附件无效", "The reply target or attachment is invalid"))
		return
	}
	if err != nil {
		httpx.Error(w, http.StatusInternalServerError, httpx.Text(r, "发表评论失败", "Could not publish the comment"))
		return
	}
	httpx.JSON(w, http.StatusCreated, created)
}

func (h *Handler) updateComment(w http.ResponseWriter, r *http.Request) {
	user, ok := h.currentUser(w, r)
	if !ok {
		return
	}
	feedback, comment, ok := h.commentTarget(w, r)
	if !ok {
		return
	}
	var input domain.CommentInput
	if !httpx.Decode(w, r, &input) {
		return
	}
	input.Body = strings.TrimSpace(input.Body)
	if !withinRunes(input.Body, 1, 8000) || input.ReplyTo != 0 {
		httpx.Error(w, http.StatusBadRequest, httpx.Text(r, "评论需要 1-8000 字", "A comment needs 1-8000 characters"))
		return
	}
	if user.Role != "admin" && user.Username != comment.Author {
		httpx.Error(w, http.StatusForbidden, httpx.Text(r, "只能编辑自己的评论", "You can only edit your own comment"))
		return
	}
	if comment.FeedbackID != feedback.ID {
		httpx.Error(w, http.StatusNotFound, httpx.Text(r, "未找到该评论", "Comment was not found"))
		return
	}
	saved, updated, err := h.store.UpdateComment(comment.ID, input.Body)
	if errors.Is(err, domain.ErrAttachmentMissing) {
		httpx.Error(w, http.StatusBadRequest, httpx.Text(r, "回复目标或附件无效", "The reply target or attachment is invalid"))
		return
	}
	if err != nil {
		httpx.Error(w, http.StatusInternalServerError, httpx.Text(r, "更新评论失败", "Could not update the comment"))
		return
	}
	if !updated {
		httpx.Error(w, http.StatusNotFound, httpx.Text(r, "未找到该评论", "Comment was not found"))
		return
	}
	httpx.JSON(w, http.StatusOK, saved)
}

func (h *Handler) deleteComment(w http.ResponseWriter, r *http.Request) {
	user, ok := h.currentUser(w, r)
	if !ok {
		return
	}
	feedback, comment, ok := h.commentTarget(w, r)
	if !ok {
		return
	}
	if user.Role != "admin" && user.Username != comment.Author {
		httpx.Error(w, http.StatusForbidden, httpx.Text(r, "只能删除自己的评论", "You can only delete your own comment"))
		return
	}
	if comment.FeedbackID != feedback.ID {
		httpx.Error(w, http.StatusNotFound, httpx.Text(r, "未找到该评论", "Comment was not found"))
		return
	}
	deleted, err := h.store.DeleteComment(comment.ID, user.Username)
	if err != nil {
		httpx.Error(w, http.StatusInternalServerError, httpx.Text(r, "删除评论失败", "Could not delete the comment"))
		return
	}
	if !deleted {
		httpx.Error(w, http.StatusNotFound, httpx.Text(r, "未找到该评论", "Comment was not found"))
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func (h *Handler) commentTarget(w http.ResponseWriter, r *http.Request) (domain.Feedback, domain.Comment, bool) {
	id, ok := feedbackID(w, r)
	if !ok {
		return domain.Feedback{}, domain.Comment{}, false
	}
	commentID, err := strconv.ParseInt(chi.URLParam(r, "commentId"), 10, 64)
	if err != nil || commentID < 1 {
		httpx.Error(w, http.StatusBadRequest, httpx.Text(r, "评论编号无效", "Invalid comment id"))
		return domain.Feedback{}, domain.Comment{}, false
	}
	feedback, err := h.store.GetFeedback(id)
	if errors.Is(err, sql.ErrNoRows) {
		httpx.Error(w, http.StatusNotFound, httpx.Text(r, "未找到该反馈", "Feedback was not found"))
		return domain.Feedback{}, domain.Comment{}, false
	}
	if err != nil {
		httpx.Error(w, http.StatusInternalServerError, httpx.Text(r, "读取评论失败", "Could not load comments"))
		return domain.Feedback{}, domain.Comment{}, false
	}
	comment, err := h.store.GetComment(commentID)
	if errors.Is(err, sql.ErrNoRows) || comment.FeedbackID != feedback.ID {
		httpx.Error(w, http.StatusNotFound, httpx.Text(r, "未找到该评论", "Comment was not found"))
		return domain.Feedback{}, domain.Comment{}, false
	}
	if err != nil {
		httpx.Error(w, http.StatusInternalServerError, httpx.Text(r, "读取评论失败", "Could not load comments"))
		return domain.Feedback{}, domain.Comment{}, false
	}
	return feedback, comment, true
}
