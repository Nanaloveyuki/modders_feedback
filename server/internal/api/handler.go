package api

import (
	"errors"
	"io/fs"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"
	"unicode/utf8"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"modders-feedback/server/internal/auth"
	"modders-feedback/server/internal/domain"
	"modders-feedback/server/internal/httpx"
	"modders-feedback/server/internal/store"
)

type Handler struct {
	store    *store.Store
	sessions auth.Sessions
	webDir   string
}

func New(store *store.Store, sessions auth.Sessions, webDir string) *Handler {
	if _, err := os.Stat(webDir); errors.Is(err, os.ErrNotExist) {
		webDir = "../web/dist"
	}
	return &Handler{store: store, sessions: sessions, webDir: webDir}
}

func (h *Handler) Routes() *chi.Mux {
	r := chi.NewRouter()
	r.Use(middleware.RequestID, middleware.RealIP, middleware.Recoverer, middleware.Timeout(15*time.Second))
	r.Get("/api/health", h.health)
	r.Post("/api/auth/login", h.login)
	r.Post("/api/auth/logout", h.logout)
	r.Get("/api/auth/me", h.me)
	r.Get("/api/feedback", h.listFeedback)
	r.Post("/api/feedback", h.createFeedback)
	r.Patch("/api/feedback/{id}/status", h.updateStatus)
	r.Handle("/*", h.static())
	return r
}

func (h *Handler) health(w http.ResponseWriter, _ *http.Request) {
	httpx.JSON(w, http.StatusOK, map[string]string{"status": "ok"})
}

func (h *Handler) login(w http.ResponseWriter, r *http.Request) {
	var input domain.Credentials
	if !httpx.Decode(w, r, &input) {
		return
	}
	ok, err := h.store.Authenticate(input.Username, input.Password)
	if err != nil {
		httpx.Error(w, http.StatusInternalServerError, "登录暂不可用")
		return
	}
	if !ok {
		httpx.Error(w, http.StatusUnauthorized, "用户名或密码错误")
		return
	}
	if err := h.sessions.Issue(w, input.Username); err != nil {
		httpx.Error(w, http.StatusInternalServerError, "登录暂不可用")
		return
	}
	httpx.JSON(w, http.StatusOK, map[string]string{"username": input.Username})
}

func (h *Handler) logout(w http.ResponseWriter, _ *http.Request) {
	h.sessions.Clear(w)
	w.WriteHeader(http.StatusNoContent)
}

func (h *Handler) me(w http.ResponseWriter, r *http.Request) {
	username, ok := h.sessions.Username(r)
	if !ok {
		httpx.Error(w, http.StatusUnauthorized, "请先登录")
		return
	}
	httpx.JSON(w, http.StatusOK, map[string]string{"username": username})
}

func (h *Handler) listFeedback(w http.ResponseWriter, r *http.Request) {
	category := r.URL.Query().Get("category")
	if category != "" && !domain.ValidCategory(category) {
		httpx.Error(w, http.StatusBadRequest, "反馈分类无效")
		return
	}
	items, err := h.store.ListFeedback(category)
	if err != nil {
		httpx.Error(w, http.StatusInternalServerError, "读取反馈失败")
		return
	}
	httpx.JSON(w, http.StatusOK, items)
}

func (h *Handler) createFeedback(w http.ResponseWriter, r *http.Request) {
	username, ok := h.sessions.Username(r)
	if !ok {
		httpx.Error(w, http.StatusUnauthorized, "登录后才能提交反馈")
		return
	}
	var item domain.Feedback
	if !httpx.Decode(w, r, &item) {
		return
	}
	item.Category = strings.TrimSpace(item.Category)
	item.Title = strings.TrimSpace(item.Title)
	item.Body = strings.TrimSpace(item.Body)
	if !domain.ValidCategory(item.Category) || !withinRunes(item.Title, 5, 120) || !withinRunes(item.Body, 10, 12000) {
		httpx.Error(w, http.StatusBadRequest, "请检查分类、标题（5-120 字）和描述（10-12000 字）")
		return
	}
	if len(item.GameVersion) > 40 || len(item.ModVersion) > 80 || len(item.ModList) > 6000 || len(item.SaveLink) > 500 {
		httpx.Error(w, http.StatusBadRequest, "版本、模组列表或存档链接超出长度限制")
		return
	}
	item.Author = username
	created, err := h.store.CreateFeedback(item)
	if err != nil {
		httpx.Error(w, http.StatusInternalServerError, "提交反馈失败")
		return
	}
	httpx.JSON(w, http.StatusCreated, created)
}

func (h *Handler) updateStatus(w http.ResponseWriter, r *http.Request) {
	if _, ok := h.sessions.Username(r); !ok {
		httpx.Error(w, http.StatusUnauthorized, "请先登录")
		return
	}
	var input domain.StatusUpdate
	if !httpx.Decode(w, r, &input) {
		return
	}
	if !domain.ValidStatus(input.Status) {
		httpx.Error(w, http.StatusBadRequest, "反馈状态无效")
		return
	}
	updated, err := h.store.UpdateStatus(chi.URLParam(r, "id"), input.Status)
	if err != nil {
		httpx.Error(w, http.StatusBadRequest, "反馈编号无效")
		return
	}
	if !updated {
		httpx.Error(w, http.StatusNotFound, "未找到该反馈")
		return
	}
	httpx.JSON(w, http.StatusOK, map[string]string{"status": input.Status})
}

func (h *Handler) static() http.Handler {
	files := http.FileServer(http.Dir(h.webDir))
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if strings.HasPrefix(r.URL.Path, "/api/") || r.URL.Path == "/api" {
			httpx.JSON(w, http.StatusNotFound, map[string]string{"error": "未找到接口"})
			return
		}
		if strings.Contains(r.URL.Path, "..") {
			httpx.JSON(w, http.StatusBadRequest, map[string]string{"error": "请求路径无效"})
			return
		}
		if _, err := os.Stat(filepath.Join(h.webDir, filepath.Clean(r.URL.Path))); errors.Is(err, fs.ErrNotExist) {
			r.URL.Path = "/"
		}
		files.ServeHTTP(w, r)
	})
}

func withinRunes(value string, min, max int) bool {
	count := utf8.RuneCountInString(value)
	return count >= min && count <= max
}
