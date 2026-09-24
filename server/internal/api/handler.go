package api

import (
	"errors"
	"io/fs"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
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
	r.Get("/api/mods", h.listMods)
	r.Post("/api/mods", h.createMod)
	r.Patch("/api/mods/{id}", h.updateMod)
	r.Delete("/api/mods/{id}", h.deleteMod)
	r.Get("/api/feedback", h.listFeedback)
	r.Post("/api/feedback", h.createFeedback)
	r.Patch("/api/feedback/{id}/status", h.updateStatus)
	r.Patch("/api/feedback/{id}", h.updateFeedback)
	r.Delete("/api/feedback/{id}", h.deleteFeedback)
	r.Get("/api/settings", h.getSettings)
	r.Put("/api/settings", h.updateSettings)
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
		httpx.Error(w, http.StatusInternalServerError, httpx.Text(r, "登录暂不可用", "Login is unavailable"))
		return
	}
	if !ok {
		httpx.Error(w, http.StatusUnauthorized, httpx.Text(r, "用户名或密码错误", "Incorrect username or password"))
		return
	}
	if err := h.sessions.Issue(w, input.Username); err != nil {
		httpx.Error(w, http.StatusInternalServerError, httpx.Text(r, "登录暂不可用", "Login is unavailable"))
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
		httpx.Error(w, http.StatusUnauthorized, httpx.Text(r, "请先登录", "Log in first"))
		return
	}
	httpx.JSON(w, http.StatusOK, map[string]string{"username": username})
}

func (h *Handler) listFeedback(w http.ResponseWriter, r *http.Request) {
	mod, ok := h.requestedMod(w, r)
	if !ok {
		return
	}
	category := r.URL.Query().Get("category")
	if category != "" && !domain.ValidCategory(category) {
		httpx.Error(w, http.StatusBadRequest, httpx.Text(r, "反馈分类无效", "Invalid feedback category"))
		return
	}
	limit, err := boundedQuery(r, "limit", 50, 100)
	if err != nil {
		httpx.Error(w, http.StatusBadRequest, httpx.Text(r, "分页参数无效", "Invalid pagination"))
		return
	}
	offset, err := boundedQuery(r, "offset", 0, 10000)
	if err != nil {
		httpx.Error(w, http.StatusBadRequest, httpx.Text(r, "分页参数无效", "Invalid pagination"))
		return
	}
	items, err := h.store.ListFeedback(mod.ID, category, limit, offset)
	if err != nil {
		httpx.Error(w, http.StatusInternalServerError, httpx.Text(r, "读取反馈失败", "Could not load feedback"))
		return
	}
	httpx.JSON(w, http.StatusOK, items)
}

func (h *Handler) createFeedback(w http.ResponseWriter, r *http.Request) {
	username, ok := h.sessions.Username(r)
	if !ok {
		httpx.Error(w, http.StatusUnauthorized, httpx.Text(r, "登录后才能提交反馈", "Log in before submitting feedback"))
		return
	}
	mod, ok := h.requestedMod(w, r)
	if !ok {
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
		httpx.Error(w, http.StatusBadRequest, httpx.Text(r, "请检查分类、标题（5-120 字）和描述（10-12000 字）", "Check the category, title (5-120 characters), and description (10-12000 characters)"))
		return
	}
	if len(item.GameVersion) > 40 || len(item.ModVersion) > 80 || len(item.ModList) > 6000 || len(item.SaveLink) > 500 {
		httpx.Error(w, http.StatusBadRequest, httpx.Text(r, "版本、模组列表或存档链接超出长度限制", "Version, mod list, or save link is too long"))
		return
	}
	item.Author = username
	item.ModID = mod.ID
	created, err := h.store.CreateFeedback(item)
	if err != nil {
		httpx.Error(w, http.StatusInternalServerError, httpx.Text(r, "提交反馈失败", "Could not submit feedback"))
		return
	}
	httpx.JSON(w, http.StatusCreated, created)
}

func (h *Handler) updateStatus(w http.ResponseWriter, r *http.Request) {
	if _, ok := h.sessions.Username(r); !ok {
		httpx.Error(w, http.StatusUnauthorized, httpx.Text(r, "请先登录", "Log in first"))
		return
	}
	var input domain.StatusUpdate
	if !httpx.Decode(w, r, &input) {
		return
	}
	if !domain.ValidStatus(input.Status) {
		httpx.Error(w, http.StatusBadRequest, httpx.Text(r, "反馈状态无效", "Invalid feedback status"))
		return
	}
	id, err := strconv.ParseInt(chi.URLParam(r, "id"), 10, 64)
	if err != nil || id < 1 {
		httpx.Error(w, http.StatusBadRequest, httpx.Text(r, "反馈编号无效", "Invalid feedback id"))
		return
	}
	updated, err := h.store.UpdateStatus(id, input.Status)
	if err != nil {
		httpx.Error(w, http.StatusInternalServerError, httpx.Text(r, "更新反馈状态失败", "Could not update feedback status"))
		return
	}
	if !updated {
		httpx.Error(w, http.StatusNotFound, httpx.Text(r, "未找到该反馈", "Feedback was not found"))
		return
	}
	httpx.JSON(w, http.StatusOK, map[string]string{"status": input.Status})
}

func (h *Handler) updateFeedback(w http.ResponseWriter, r *http.Request) {
	if _, ok := h.sessions.Username(r); !ok {
		httpx.Error(w, http.StatusUnauthorized, httpx.Text(r, "请先登录", "Log in first"))
		return
	}
	id, ok := feedbackID(w, r)
	if !ok {
		return
	}
	var input domain.FeedbackUpdate
	if !httpx.Decode(w, r, &input) {
		return
	}
	input.Title = strings.TrimSpace(input.Title)
	input.Body = strings.TrimSpace(input.Body)
	input.GameVersion = strings.TrimSpace(input.GameVersion)
	input.ModVersion = strings.TrimSpace(input.ModVersion)
	input.ModList = strings.TrimSpace(input.ModList)
	input.SaveLink = strings.TrimSpace(input.SaveLink)
	if !withinRunes(input.Title, 5, 120) || !withinRunes(input.Body, 10, 12000) {
		httpx.Error(w, http.StatusBadRequest, httpx.Text(r, "请检查标题（5-120 字）和描述（10-12000 字）", "Check the title (5-120 characters) and description (10-12000 characters)"))
		return
	}
	if len(input.GameVersion) > 40 || len(input.ModVersion) > 80 || len(input.ModList) > 6000 || len(input.SaveLink) > 500 {
		httpx.Error(w, http.StatusBadRequest, httpx.Text(r, "版本、模组列表或存档链接超出长度限制", "Version, mod list, or save link is too long"))
		return
	}
	item, updated, err := h.store.UpdateFeedback(id, input)
	if err != nil {
		httpx.Error(w, http.StatusInternalServerError, httpx.Text(r, "更新反馈失败", "Could not update feedback"))
		return
	}
	if !updated {
		httpx.Error(w, http.StatusNotFound, httpx.Text(r, "未找到该反馈", "Feedback was not found"))
		return
	}
	httpx.JSON(w, http.StatusOK, item)
}

func (h *Handler) deleteFeedback(w http.ResponseWriter, r *http.Request) {
	if _, ok := h.sessions.Username(r); !ok {
		httpx.Error(w, http.StatusUnauthorized, httpx.Text(r, "请先登录", "Log in first"))
		return
	}
	id, ok := feedbackID(w, r)
	if !ok {
		return
	}
	deleted, err := h.store.DeleteFeedback(id)
	if err != nil {
		httpx.Error(w, http.StatusInternalServerError, httpx.Text(r, "删除反馈失败", "Could not delete feedback"))
		return
	}
	if !deleted {
		httpx.Error(w, http.StatusNotFound, httpx.Text(r, "未找到该反馈", "Feedback was not found"))
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func (h *Handler) getSettings(w http.ResponseWriter, r *http.Request) {
	settings, err := h.store.SiteSettings()
	if err != nil {
		httpx.Error(w, http.StatusInternalServerError, httpx.Text(r, "读取站点设置失败", "Could not load site settings"))
		return
	}
	httpx.JSON(w, http.StatusOK, settings)
}

func (h *Handler) updateSettings(w http.ResponseWriter, r *http.Request) {
	if _, ok := h.sessions.Username(r); !ok {
		httpx.Error(w, http.StatusUnauthorized, httpx.Text(r, "请先登录", "Log in first"))
		return
	}
	var input domain.SiteSettings
	if !httpx.Decode(w, r, &input) {
		return
	}
	input.ModVersion = strings.TrimSpace(input.ModVersion)
	input.GameVersion = strings.TrimSpace(input.GameVersion)
	input.Icon = strings.TrimSpace(input.Icon)
	if !withinRunes(input.ModVersion, 1, 40) || !withinRunes(input.GameVersion, 1, 40) || !domain.ValidIcon(input.Icon) {
		httpx.Error(w, http.StatusBadRequest, httpx.Text(r, "请检查版本号（1-40 字）和图标", "Check the versions (1-40 characters) and icon"))
		return
	}
	if err := h.store.UpdateSiteSettings(input); err != nil {
		httpx.Error(w, http.StatusInternalServerError, httpx.Text(r, "保存站点设置失败", "Could not save site settings"))
		return
	}
	httpx.JSON(w, http.StatusOK, input)
}

func feedbackID(w http.ResponseWriter, r *http.Request) (int64, bool) {
	id, err := strconv.ParseInt(chi.URLParam(r, "id"), 10, 64)
	if err != nil || id < 1 {
		httpx.Error(w, http.StatusBadRequest, httpx.Text(r, "反馈编号无效", "Invalid feedback id"))
		return 0, false
	}
	return id, true
}

func (h *Handler) static() http.Handler {
	files := http.FileServer(http.Dir(h.webDir))
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if strings.HasPrefix(r.URL.Path, "/api/") || r.URL.Path == "/api" {
			httpx.JSON(w, http.StatusNotFound, map[string]string{"error": httpx.Text(r, "未找到接口", "Endpoint not found")})
			return
		}
		if strings.Contains(r.URL.Path, "..") {
			httpx.JSON(w, http.StatusBadRequest, map[string]string{"error": httpx.Text(r, "请求路径无效", "Invalid request path")})
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

func boundedQuery(r *http.Request, key string, fallback, max int) (int, error) {
	value := r.URL.Query().Get(key)
	if value == "" {
		return fallback, nil
	}
	parsed, err := strconv.Atoi(value)
	if err != nil || parsed < 0 || parsed > max {
		return 0, errOrRange(err)
	}
	return parsed, nil
}

func errOrRange(err error) error {
	if err != nil {
		return err
	}
	return strconv.ErrRange
}

func (h *Handler) listMods(w http.ResponseWriter, r *http.Request) {
	mods, err := h.store.ListMods()
	if err != nil {
		httpx.Error(w, http.StatusInternalServerError, httpx.Text(r, "读取模组失败", "Could not load mods"))
		return
	}
	httpx.JSON(w, http.StatusOK, mods)
}

func (h *Handler) createMod(w http.ResponseWriter, r *http.Request) {
	if _, ok := h.sessions.Username(r); !ok {
		httpx.Error(w, http.StatusUnauthorized, httpx.Text(r, "请先登录", "Log in first"))
		return
	}
	input, ok := decodeMod(w, r)
	if !ok {
		return
	}
	created, err := h.store.CreateMod(input)
	if err != nil {
		httpx.Error(w, http.StatusConflict, httpx.Text(r, "无法添加模组，短标识可能已存在", "Could not add the mod; the slug may already exist"))
		return
	}
	httpx.JSON(w, http.StatusCreated, created)
}

func (h *Handler) updateMod(w http.ResponseWriter, r *http.Request) {
	if _, ok := h.sessions.Username(r); !ok {
		httpx.Error(w, http.StatusUnauthorized, httpx.Text(r, "请先登录", "Log in first"))
		return
	}
	id, ok := feedbackID(w, r)
	if !ok {
		return
	}
	input, ok := decodeMod(w, r)
	if !ok {
		return
	}
	updated, found, err := h.store.UpdateMod(id, input)
	if err != nil {
		httpx.Error(w, http.StatusConflict, httpx.Text(r, "无法保存模组，短标识可能已存在", "Could not save the mod; the slug may already exist"))
		return
	}
	if !found {
		httpx.Error(w, http.StatusNotFound, httpx.Text(r, "未找到该模组", "Mod was not found"))
		return
	}
	httpx.JSON(w, http.StatusOK, updated)
}

func (h *Handler) deleteMod(w http.ResponseWriter, r *http.Request) {
	if _, ok := h.sessions.Username(r); !ok {
		httpx.Error(w, http.StatusUnauthorized, httpx.Text(r, "请先登录", "Log in first"))
		return
	}
	id, ok := feedbackID(w, r)
	if !ok {
		return
	}
	deleted, err := h.store.DeleteMod(id)
	if errors.Is(err, domain.ErrLastMod) {
		httpx.Error(w, http.StatusConflict, httpx.Text(r, "至少保留一个模组", "Keep at least one mod"))
		return
	}
	if err != nil {
		httpx.Error(w, http.StatusInternalServerError, httpx.Text(r, "删除模组失败", "Could not delete the mod"))
		return
	}
	if !deleted {
		httpx.Error(w, http.StatusNotFound, httpx.Text(r, "未找到该模组", "Mod was not found"))
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func (h *Handler) requestedMod(w http.ResponseWriter, r *http.Request) (domain.Mod, bool) {
	slug := strings.TrimSpace(r.URL.Query().Get("mod"))
	if slug == "" {
		slug = domain.DefaultModSlug
	}
	if !domain.ValidSlug(slug) {
		httpx.Error(w, http.StatusBadRequest, httpx.Text(r, "模组标识无效", "Invalid mod slug"))
		return domain.Mod{}, false
	}
	mod, err := h.store.ModBySlug(slug)
	if err != nil {
		httpx.Error(w, http.StatusNotFound, httpx.Text(r, "未找到该模组", "Mod was not found"))
		return domain.Mod{}, false
	}
	return mod, true
}

func decodeMod(w http.ResponseWriter, r *http.Request) (domain.ModInput, bool) {
	var input domain.ModInput
	if !httpx.Decode(w, r, &input) {
		return domain.ModInput{}, false
	}
	input.Slug = strings.TrimSpace(input.Slug)
	input.Name = strings.TrimSpace(input.Name)
	input.GameVersion = strings.TrimSpace(input.GameVersion)
	input.ModVersion = strings.TrimSpace(input.ModVersion)
	input.Icon = strings.TrimSpace(input.Icon)
	if !domain.ValidSlug(input.Slug) || !withinRunes(input.Name, 1, 40) || !withinRunes(input.GameVersion, 1, 40) || !withinRunes(input.ModVersion, 1, 40) || !domain.ValidIcon(input.Icon) {
		httpx.Error(w, http.StatusBadRequest, httpx.Text(r, "请检查模组标识、名称（1-40 字）、版本和图标", "Check the slug, name (1-40 characters), versions, and icon"))
		return domain.ModInput{}, false
	}
	return input, true
}
