package api

import (
	"bytes"
	"database/sql"
	"encoding/base64"
	"encoding/json"
	"errors"
	"image"
	"image/draw"
	"image/png"
	"io/fs"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"
	"unicode/utf8"

	_ "golang.org/x/image/webp"
	_ "image/gif"
	_ "image/jpeg"

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
	r.Post("/api/auth/register", h.register)
	r.Post("/api/auth/logout", h.logout)
	r.Get("/api/auth/me", h.me)
	r.Patch("/api/account", h.updateAccount)
	r.Patch("/api/account/password", h.updatePassword)
	r.Put("/api/account/avatar", h.updateAvatar)
	r.Get("/api/account/avatar", h.avatar)
	r.Get("/api/mods", h.listMods)
	r.Post("/api/mods", h.createMod)
	r.Patch("/api/mods/{id}", h.updateMod)
	r.Delete("/api/mods/{id}", h.deleteMod)
	r.Get("/api/feedback", h.listFeedback)
	r.Post("/api/feedback", h.createFeedback)
	r.Patch("/api/feedback/{id}/status", h.updateStatus)
	r.Patch("/api/feedback/{id}", h.updateFeedback)
	r.Delete("/api/feedback/{id}", h.deleteFeedback)
	r.Get("/api/mods/{slug}/feedback/{category}/{publicId}", h.getPublicFeedback)
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
	user, ok, err := h.store.Authenticate(input.Username, input.Password)
	if err != nil {
		httpx.Error(w, http.StatusInternalServerError, httpx.Text(r, "登录暂不可用", "Login is unavailable"))
		return
	}
	if !ok {
		httpx.Error(w, http.StatusUnauthorized, httpx.Text(r, "用户名或密码错误", "Incorrect username or password"))
		return
	}
	if err := h.sessions.Issue(w, user.Username); err != nil {
		httpx.Error(w, http.StatusInternalServerError, httpx.Text(r, "登录暂不可用", "Login is unavailable"))
		return
	}
	httpx.JSON(w, http.StatusOK, user)
}

func (h *Handler) register(w http.ResponseWriter, r *http.Request) {
	var input domain.Registration
	if !httpx.Decode(w, r, &input) {
		return
	}
	input.Username = strings.TrimSpace(input.Username)
	input.Email = strings.TrimSpace(input.Email)
	input.QQ = strings.TrimSpace(input.QQ)
	if !validUsername(input.Username) || !validPassword(input.Password) || !validOptionalEmail(input.Email) || !validOptionalQQ(input.QQ) {
		httpx.Error(w, http.StatusBadRequest, httpx.Text(r, "用户名为 3-32 位字母、数字、下划线或连字符；密码至少 12 位。邮箱和 QQ 号可选。", "Username must be 3-32 letters, digits, underscores, or hyphens; password must be at least 12 characters. Email and QQ are optional."))
		return
	}
	user, err := h.store.Register(input)
	if errors.Is(err, domain.ErrUsernameTaken) {
		httpx.Error(w, http.StatusConflict, httpx.Text(r, "用户名或邮箱已被使用", "Username or email is already in use"))
		return
	}
	if err != nil {
		httpx.Error(w, http.StatusInternalServerError, httpx.Text(r, "注册暂不可用", "Registration is unavailable"))
		return
	}
	if err := h.sessions.Issue(w, user.Username); err != nil {
		httpx.Error(w, http.StatusInternalServerError, httpx.Text(r, "注册暂不可用", "Registration is unavailable"))
		return
	}
	httpx.JSON(w, http.StatusCreated, user)
}

func (h *Handler) logout(w http.ResponseWriter, _ *http.Request) {
	h.sessions.Clear(w)
	w.WriteHeader(http.StatusNoContent)
}

func (h *Handler) me(w http.ResponseWriter, r *http.Request) {
	user, ok := h.accountUser(w, r)
	if !ok {
		return
	}
	httpx.JSON(w, http.StatusOK, user)
}

func (h *Handler) updateAccount(w http.ResponseWriter, r *http.Request) {
	current, ok := h.accountUser(w, r)
	if !ok {
		return
	}
	var input domain.ProfileUpdate
	if !httpx.Decode(w, r, &input) {
		return
	}
	input.Username = strings.TrimSpace(input.Username)
	input.Email = strings.TrimSpace(input.Email)
	input.QQ = strings.TrimSpace(input.QQ)
	if !validUsername(input.Username) || !validOptionalEmail(input.Email) || !validOptionalQQ(input.QQ) {
		httpx.Error(w, http.StatusBadRequest, httpx.Text(r, "用户名为 3-32 位字母、数字、下划线或连字符。邮箱和 QQ 号可选。", "Username must be 3-32 letters, digits, underscores, or hyphens. Email and QQ are optional."))
		return
	}
	saved, err := h.store.UpdateProfile(current.Username, input)
	if errors.Is(err, domain.ErrUsernameTaken) {
		httpx.Error(w, http.StatusConflict, httpx.Text(r, "用户名已被使用", "Username is already in use"))
		return
	}
	if errors.Is(err, domain.ErrEmailTaken) {
		httpx.Error(w, http.StatusConflict, httpx.Text(r, "邮箱已被使用", "Email is already in use"))
		return
	}
	if err != nil {
		httpx.Error(w, http.StatusInternalServerError, httpx.Text(r, "保存资料失败", "Could not save the profile"))
		return
	}
	if saved.Username != current.Username {
		if err := h.sessions.Issue(w, saved.Username); err != nil {
			httpx.Error(w, http.StatusInternalServerError, httpx.Text(r, "保存资料失败", "Could not save the profile"))
			return
		}
	}
	httpx.JSON(w, http.StatusOK, saved)
}

func (h *Handler) updatePassword(w http.ResponseWriter, r *http.Request) {
	current, ok := h.accountUser(w, r)
	if !ok {
		return
	}
	var input domain.PasswordUpdate
	if !httpx.Decode(w, r, &input) {
		return
	}
	if !validPassword(input.NewPassword) || input.CurrentPassword == "" || utf8.RuneCountInString(input.CurrentPassword) > 128 {
		httpx.Error(w, http.StatusBadRequest, httpx.Text(r, "新密码至少 12 位，且需要填写现密码。", "The new password must be at least 12 characters, and the current password is required."))
		return
	}
	err := h.store.UpdatePassword(current.Username, input.CurrentPassword, input.NewPassword)
	if errors.Is(err, domain.ErrBadPassword) {
		httpx.Error(w, http.StatusUnauthorized, httpx.Text(r, "现密码不正确", "Current password is incorrect"))
		return
	}
	if err != nil {
		httpx.Error(w, http.StatusInternalServerError, httpx.Text(r, "修改密码失败", "Could not change the password"))
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func (h *Handler) updateAvatar(w http.ResponseWriter, r *http.Request) {
	current, ok := h.accountUser(w, r)
	if !ok {
		return
	}
	var input struct {
		Image string `json:"image"`
		X     int    `json:"x"`
		Y     int    `json:"y"`
		Size  int    `json:"size"`
	}
	r.Body = http.MaxBytesReader(w, r.Body, 16<<20)
	decoder := json.NewDecoder(r.Body)
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(&input); err != nil || input.Image == "" {
		httpx.Error(w, http.StatusBadRequest, httpx.Text(r, "头像内容无效", "Invalid avatar"))
		return
	}
	raw, err := base64.StdEncoding.DecodeString(input.Image)
	if err != nil || len(raw) == 0 || len(raw) > 12<<20 {
		httpx.Error(w, http.StatusBadRequest, httpx.Text(r, "头像必须是不超过 12MiB 的 PNG、JPEG、GIF 或 WebP。", "The avatar must be a PNG, JPEG, GIF, or WebP no larger than 12MiB."))
		return
	}
	decoded, _, err := image.Decode(bytes.NewReader(raw))
	if err != nil {
		httpx.Error(w, http.StatusBadRequest, httpx.Text(r, "头像必须是 PNG、JPEG、GIF 或 WebP。", "The avatar must be a PNG, JPEG, GIF, or WebP."))
		return
	}
	bounds := decoded.Bounds()
	if bounds.Dx() < 1 || bounds.Dy() < 1 || bounds.Dx() > 4096 || bounds.Dy() > 4096 || input.Size < 1 || input.X < bounds.Min.X || input.Y < bounds.Min.Y || input.X+input.Size > bounds.Max.X || input.Y+input.Size > bounds.Max.Y {
		httpx.Error(w, http.StatusBadRequest, httpx.Text(r, "裁剪区域超出图片，或边长超过 4096。", "The crop is outside the image, or a side is longer than 4096 pixels."))
		return
	}
	cropped := scaleSquare(decoded, input.X, input.Y, input.Size)
	var encoded bytes.Buffer
	if err = png.Encode(&encoded, cropped); err != nil || encoded.Len() > 2<<20 {
		httpx.Error(w, http.StatusBadRequest, httpx.Text(r, "转换后的 PNG 超过 2MiB。", "The converted PNG is larger than 2MiB."))
		return
	}
	if err = h.store.SetAvatar(current.Username, encoded.Bytes()); err != nil {
		httpx.Error(w, http.StatusInternalServerError, httpx.Text(r, "保存头像失败", "Could not save the avatar"))
		return
	}
	current.AvatarURL = "/api/account/avatar?u=" + current.Username
	httpx.JSON(w, http.StatusOK, current)
}

func scaleSquare(source image.Image, x, y, size int) *image.NRGBA {
	out := image.NewNRGBA(image.Rect(0, 0, 256, 256))
	if size == 256 {
		draw.Draw(out, out.Bounds(), source, image.Point{x, y}, draw.Src)
		return out
	}
	for py := range 256 {
		sy := y + py*size/256
		for px := range 256 {
			out.Set(px, py, source.At(x+px*size/256, sy))
		}
	}
	return out
}

func (h *Handler) avatar(w http.ResponseWriter, r *http.Request) {
	username := strings.TrimSpace(r.URL.Query().Get("u"))
	if !validUsername(username) {
		http.NotFound(w, r)
		return
	}
	pngBytes, found, err := h.store.Avatar(username)
	if err != nil {
		httpx.Error(w, http.StatusInternalServerError, httpx.Text(r, "读取头像失败", "Could not load the avatar"))
		return
	}
	if !found {
		http.NotFound(w, r)
		return
	}
	w.Header().Set("Content-Type", "image/png")
	w.Header().Set("Cache-Control", "private, max-age=300")
	w.Header().Set("X-Content-Type-Options", "nosniff")
	_, _ = w.Write(pngBytes)
}

func (h *Handler) accountUser(w http.ResponseWriter, r *http.Request) (domain.User, bool) {
	username, ok := h.sessions.Username(r)
	if !ok {
		httpx.Error(w, http.StatusUnauthorized, httpx.Text(r, "请先登录", "Log in first"))
		return domain.User{}, false
	}
	user, found, err := h.store.User(username)
	if err != nil || !found {
		if err != nil {
			httpx.Error(w, http.StatusInternalServerError, httpx.Text(r, "读取账号失败", "Could not load the account"))
			return domain.User{}, false
		}
		h.sessions.Clear(w)
		httpx.Error(w, http.StatusUnauthorized, httpx.Text(r, "请先登录", "Log in first"))
		return domain.User{}, false
	}
	return user, true
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
	user, ok := h.currentUser(w, r)
	if !ok {
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
	id, ok := feedbackID(w, r)
	if !ok {
		return
	}
	item, err := h.store.GetFeedback(id)
	if errors.Is(err, sql.ErrNoRows) {
		httpx.Error(w, http.StatusNotFound, httpx.Text(r, "未找到该反馈", "Feedback was not found"))
		return
	}
	if err != nil {
		httpx.Error(w, http.StatusInternalServerError, httpx.Text(r, "更新反馈状态失败", "Could not update feedback status"))
		return
	}
	if user.Role != "admin" && (!authorOf(user, item) || !authorStatusChange(item.Status, input.Status)) {
		httpx.Error(w, http.StatusForbidden, httpx.Text(r, "只能在自己的反馈中切换待处理和已撤回", "You can only switch your own feedback between open and withdrawn"))
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
	user, ok := h.currentUser(w, r)
	if !ok {
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
	current, err := h.store.GetFeedback(id)
	if errors.Is(err, sql.ErrNoRows) {
		httpx.Error(w, http.StatusNotFound, httpx.Text(r, "未找到该反馈", "Feedback was not found"))
		return
	}
	if err != nil {
		httpx.Error(w, http.StatusInternalServerError, httpx.Text(r, "更新反馈失败", "Could not update feedback"))
		return
	}
	if user.Role != "admin" && !authorOf(user, current) {
		httpx.Error(w, http.StatusForbidden, httpx.Text(r, "只能编辑自己的反馈", "You can only edit your own feedback"))
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
	if _, ok := h.requireAdmin(w, r); !ok {
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
	if _, ok := h.requireAdmin(w, r); !ok {
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

func (h *Handler) getPublicFeedback(w http.ResponseWriter, r *http.Request) {
	slug := chi.URLParam(r, "slug")
	if !domain.ValidSlug(slug) {
		httpx.Error(w, http.StatusBadRequest, httpx.Text(r, "模组标识无效", "Invalid mod slug"))
		return
	}
	category, ok := domain.CategoryFromPublic(chi.URLParam(r, "category"))
	if !ok {
		httpx.Error(w, http.StatusBadRequest, httpx.Text(r, "反馈分类无效", "Invalid feedback category"))
		return
	}
	publicID := strings.ToLower(chi.URLParam(r, "publicId"))
	if !domain.ValidPublicID(publicID) {
		httpx.Error(w, http.StatusBadRequest, httpx.Text(r, "反馈编号无效", "Invalid feedback id"))
		return
	}
	mod, err := h.store.ModBySlug(slug)
	if errors.Is(err, sql.ErrNoRows) {
		httpx.Error(w, http.StatusNotFound, httpx.Text(r, "未找到该模组", "Mod was not found"))
		return
	}
	if err != nil {
		httpx.Error(w, http.StatusInternalServerError, httpx.Text(r, "读取反馈失败", "Could not load feedback"))
		return
	}
	item, err := h.store.GetFeedbackByPublicID(mod.ID, category, publicID)
	if errors.Is(err, sql.ErrNoRows) {
		httpx.Error(w, http.StatusNotFound, httpx.Text(r, "未找到该反馈", "Feedback was not found"))
		return
	}
	if err != nil {
		httpx.Error(w, http.StatusInternalServerError, httpx.Text(r, "读取反馈失败", "Could not load feedback"))
		return
	}
	httpx.JSON(w, http.StatusOK, item)
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
	if _, ok := h.requireAdmin(w, r); !ok {
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
	if _, ok := h.requireAdmin(w, r); !ok {
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
	if _, ok := h.requireAdmin(w, r); !ok {
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

func (h *Handler) currentUser(w http.ResponseWriter, r *http.Request) (domain.User, bool) {
	username, ok := h.sessions.Username(r)
	if !ok {
		httpx.Error(w, http.StatusUnauthorized, httpx.Text(r, "请先登录", "Log in first"))
		return domain.User{}, false
	}
	user, found, err := h.store.User(username)
	if err != nil {
		httpx.Error(w, http.StatusInternalServerError, httpx.Text(r, "读取账号失败", "Could not load the account"))
		return domain.User{}, false
	}
	if !found {
		h.sessions.Clear(w)
		httpx.Error(w, http.StatusUnauthorized, httpx.Text(r, "请先登录", "Log in first"))
		return domain.User{}, false
	}
	return user, true
}

func authorOf(user domain.User, item domain.Feedback) bool {
	return user.Username == item.Author
}

func authorStatusChange(current, next string) bool {
	return (current == domain.StatusOpen || current == domain.StatusWithdrawn) && (next == domain.StatusOpen || next == domain.StatusWithdrawn)
}

func (h *Handler) requireAdmin(w http.ResponseWriter, r *http.Request) (domain.User, bool) {
	username, ok := h.sessions.Username(r)
	if !ok {
		httpx.Error(w, http.StatusUnauthorized, httpx.Text(r, "请先登录", "Log in first"))
		return domain.User{}, false
	}
	user, found, err := h.store.User(username)
	if err != nil {
		httpx.Error(w, http.StatusInternalServerError, httpx.Text(r, "读取账号失败", "Could not load the account"))
		return domain.User{}, false
	}
	if !found || user.Role != "admin" {
		httpx.Error(w, http.StatusForbidden, httpx.Text(r, "需要管理员权限", "Administrator access is required"))
		return domain.User{}, false
	}
	return user, true
}

func validUsername(value string) bool {
	if !withinRunes(value, 3, 32) {
		return false
	}
	for _, char := range value {
		if (char < 'a' || char > 'z') && (char < 'A' || char > 'Z') && (char < '0' || char > '9') && char != '_' && char != '-' {
			return false
		}
	}
	return true
}

func validPassword(value string) bool {
	count := utf8.RuneCountInString(value)
	return count >= 12 && count <= 128
}

func validOptionalEmail(value string) bool {
	if value == "" {
		return true
	}
	if !withinRunes(value, 3, 254) || strings.ContainsAny(value, " \t\r\n") {
		return false
	}
	at := strings.IndexByte(value, '@')
	dot := strings.LastIndexByte(value, '.')
	return at > 0 && dot > at+1 && dot < len(value)-1 && strings.Count(value, "@") == 1
}

func validOptionalQQ(value string) bool {
	if value == "" {
		return true
	}
	if len(value) < 5 || len(value) > 11 || value[0] == '0' {
		return false
	}
	for _, char := range value {
		if char < '0' || char > '9' {
			return false
		}
	}
	return true
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
	input.SteamURL = strings.TrimSpace(input.SteamURL)
	input.GitHubURL = strings.TrimSpace(input.GitHubURL)
	if !domain.ValidSlug(input.Slug) || !withinRunes(input.Name, 1, 40) || !withinRunes(input.GameVersion, 1, 40) || !withinRunes(input.ModVersion, 1, 40) || !domain.ValidIcon(input.Icon) || !domain.ValidOptionalURL(input.SteamURL) || !domain.ValidOptionalURL(input.GitHubURL) {
		httpx.Error(w, http.StatusBadRequest, httpx.Text(r, "请检查模组标识、名称（1-40 字）、版本、图标和链接", "Check the slug, name (1-40 characters), versions, icon, and links"))
		return domain.ModInput{}, false
	}
	return input, true
}
