package domain

import (
	"errors"
	"net/url"
	"path/filepath"
	"strings"
	"time"
)

var (
	ErrLastMod           = errors.New("cannot delete the last mod")
	ErrUsernameTaken     = errors.New("username already exists")
	ErrEmailTaken        = errors.New("email already exists")
	ErrBadPassword       = errors.New("current password is incorrect")
	ErrAttachmentMissing = errors.New("attachment was not found")
	ErrCommentTarget     = errors.New("comment reply target was not found")
)

const (
	CategoryBug      = "bug"
	CategoryFeature  = "feature"
	CategoryQuestion = "question"

	StatusOpen            = "open"
	StatusInProgress      = "in_progress"
	StatusTesting         = "testing"
	StatusFixedUnreleased = "fixed_unreleased"
	StatusResolved        = "resolved"
	StatusClosed          = "closed"
	StatusWithdrawn       = "withdrawn"

	DefaultModSlug   = "rhah"
	DefaultSteamURL  = "https://steamcommunity.com/sharedfiles/filedetails/?id="
	DefaultGitHubURL = "https://github.com/Nanaloveyuki/modders_feedback"
)

type Feedback struct {
	ID             int64        `json:"id"`
	PublicID       string       `json:"publicId"`
	ModID          int64        `json:"modId"`
	ModSlug        string       `json:"modSlug,omitempty"`
	ModName        string       `json:"modName,omitempty"`
	Category       string       `json:"category"`
	CategoryNumber int64        `json:"categoryNumber"`
	Title          string       `json:"title"`
	Body           string       `json:"body"`
	Author         string       `json:"author"`
	AuthorAvatar   string       `json:"authorAvatar"`
	GameVersion    string       `json:"gameVersion"`
	ModVersion     string       `json:"modVersion"`
	ModList        string       `json:"modList"`
	SaveLink       string       `json:"saveLink"`
	Status         string       `json:"status"`
	CreatedAt      time.Time    `json:"createdAt"`
	Attachments    []Attachment `json:"attachments"`
}

type Attachment struct {
	ID          string `json:"id"`
	Name        string `json:"name"`
	ContentType string `json:"contentType"`
	Size        int64  `json:"size"`
	URL         string `json:"url"`
}

type Credentials struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

type Registration struct {
	Username string `json:"username"`
	Password string `json:"password"`
	Email    string `json:"email"`
	QQ       string `json:"qq"`
}

type User struct {
	Username  string `json:"username"`
	Role      string `json:"role"`
	Email     string `json:"email"`
	QQ        string `json:"qq"`
	AvatarURL string `json:"avatarUrl"`
}

type ProfileUpdate struct {
	Username string `json:"username"`
	Email    string `json:"email"`
	QQ       string `json:"qq"`
}

type PasswordUpdate struct {
	CurrentPassword string `json:"currentPassword"`
	NewPassword     string `json:"newPassword"`
}

type StatusUpdate struct {
	Status string `json:"status"`
}

type Comment struct {
	ID           int64     `json:"id"`
	FeedbackID   int64     `json:"feedbackId"`
	Author       string    `json:"author"`
	AuthorAvatar string    `json:"authorAvatar"`
	Body         string    `json:"body"`
	ReplyTo      int64     `json:"replyTo,omitempty"`
	CreatedAt    time.Time `json:"createdAt"`
	UpdatedAt    time.Time `json:"updatedAt"`
}

type CommentInput struct {
	Body    string `json:"body"`
	ReplyTo int64  `json:"replyTo"`
}

const (
	EventComment       = "comment"
	EventStatus        = "status"
	EventEdited        = "edited"
	EventCommentEdited = "comment_edited"
	EventDeleted       = "deleted"
)

type TimelineEvent struct {
	ID          int64     `json:"id"`
	Kind        string    `json:"kind"`
	Actor       string    `json:"actor"`
	ActorAvatar string    `json:"actorAvatar"`
	Body        string    `json:"body,omitempty"`
	ReplyTo     int64     `json:"replyTo,omitempty"`
	Status      string    `json:"status,omitempty"`
	CommentID   int64     `json:"commentId,omitempty"`
	CreatedAt   time.Time `json:"createdAt"`
}

type Timeline struct {
	Events []TimelineEvent `json:"events"`
}
type FeedbackUpdate struct {
	Title       string `json:"title"`
	Body        string `json:"body"`
	GameVersion string `json:"gameVersion"`
	ModVersion  string `json:"modVersion"`
	ModList     string `json:"modList"`
	SaveLink    string `json:"saveLink"`
}

type SiteSettings struct {
	ModVersion    string        `json:"modVersion"`
	GameVersion   string        `json:"gameVersion"`
	Icon          string        `json:"icon"`
	AttachmentDir string        `json:"attachmentDir"`
	Statuses      []StatusLabel `json:"statuses"`
}

type StatusLabel struct {
	Key      string `json:"key"`
	LabelZh  string `json:"labelZh"`
	LabelEn  string `json:"labelEn"`
	Light    string `json:"light"`
	Dark     string `json:"dark"`
	Author   bool   `json:"author"`
	Archived bool   `json:"archived"`
}

type StatusCatalog struct {
	Statuses []StatusLabel `json:"statuses"`
}
type Mod struct {
	ID          int64  `json:"id"`
	Slug        string `json:"slug"`
	Name        string `json:"name"`
	GameVersion string `json:"gameVersion"`
	ModVersion  string `json:"modVersion"`
	Icon        string `json:"icon"`
	SteamURL    string `json:"steamUrl"`
	GitHubURL   string `json:"githubUrl"`
}

type ModInput struct {
	Slug        string `json:"slug"`
	Name        string `json:"name"`
	GameVersion string `json:"gameVersion"`
	ModVersion  string `json:"modVersion"`
	Icon        string `json:"icon"`
	SteamURL    string `json:"steamUrl"`
	GitHubURL   string `json:"githubUrl"`
}

func ValidCategory(category string) bool {
	switch category {
	case CategoryBug, CategoryFeature, CategoryQuestion:
		return true
	default:
		return false
	}
}

func ValidIcon(icon string) bool {
	switch icon {
	case "squirrel", "rat", "bug", "spark", "shield", "paw":
		return true
	default:
		return false
	}
}
func ValidSlug(slug string) bool {
	if len(slug) < 1 || len(slug) > 40 {
		return false
	}
	for i := range len(slug) {
		c := slug[i]
		if (c >= 'a' && c <= 'z') || (c >= '0' && c <= '9') || c == '-' {
			continue
		}
		return false
	}
	return slug[0] != '-' && slug[len(slug)-1] != '-'
}
func ValidOptionalURL(value string) bool {
	if value == "" {
		return true
	}
	if len(value) > 500 || strings.ContainsAny(value, " \t\r\n") {
		return false
	}
	parsed, err := url.ParseRequestURI(value)
	return err == nil && (parsed.Scheme == "http" || parsed.Scheme == "https") && parsed.Host != ""
}

func ValidStatusKey(key string) bool {
	if len(key) < 2 || len(key) > 32 {
		return false
	}
	for i, r := range key {
		ok := (r >= 'a' && r <= 'z') || (r >= '0' && r <= '9') || r == '_'
		if i == 0 && (r < 'a' || r > 'z') {
			return false
		}
		if !ok {
			return false
		}
	}
	return true
}

func ValidHexColor(value string) bool {
	if len(value) != 7 || value[0] != '#' {
		return false
	}
	for _, r := range value[1:] {
		ok := (r >= '0' && r <= '9') || (r >= 'a' && r <= 'f') || (r >= 'A' && r <= 'F')
		if !ok {
			return false
		}
	}
	return true
}

func ValidStatusLabel(item StatusLabel) bool {
	return ValidStatusKey(item.Key) && ValidHexColor(item.Light) && ValidHexColor(item.Dark) && withinLabel(item.LabelZh) && withinLabel(item.LabelEn)
}

func withinLabel(value string) bool {
	n := 0
	for range value {
		n++
		if n > 24 {
			return false
		}
	}
	return n >= 1
}

func ValidStatusCatalog(items []StatusLabel) bool {
	if len(items) == 0 || len(items) > 24 {
		return false
	}
	seen := map[string]bool{}
	hasOpen := false
	for _, item := range items {
		if !ValidStatusLabel(item) || seen[item.Key] {
			return false
		}
		seen[item.Key] = true
		if item.Key == StatusOpen {
			hasOpen = true
		}
	}
	return hasOpen
}

func DefaultStatuses() []StatusLabel {
	return []StatusLabel{
		{Key: StatusOpen, LabelZh: "待处理", LabelEn: "Open", Light: "#b8892e", Dark: "#e0b15a", Author: true},
		{Key: StatusInProgress, LabelZh: "处理中", LabelEn: "In progress", Light: "#3d7ea6", Dark: "#7eb6d4"},
		{Key: StatusTesting, LabelZh: "测试中", LabelEn: "Testing", Light: "#6a6db8", Dark: "#a8aae0"},
		{Key: StatusFixedUnreleased, LabelZh: "已修复未发布", LabelEn: "Fixed, unreleased", Light: "#2f8a78", Dark: "#5ec4ae", Archived: true},
		{Key: StatusResolved, LabelZh: "已解决", LabelEn: "Resolved", Light: "#3f8f6b", Dark: "#5dbe8a", Archived: true},
		{Key: StatusClosed, LabelZh: "已关闭", LabelEn: "Closed", Light: "#7d8794", Dark: "#a8b0bc"},
		{Key: StatusWithdrawn, LabelZh: "已撤回", LabelEn: "Withdrawn", Light: "#c45b70", Dark: "#e08a98", Author: true},
	}
}

func StatusByKey(items []StatusLabel, key string) (StatusLabel, bool) {
	for _, item := range items {
		if item.Key == key {
			return item, true
		}
	}
	return StatusLabel{}, false
}

func PublicCategory(category string) string {
	if category == CategoryBug {
		return "bugs"
	}
	return category
}

func CategoryFromPublic(segment string) (string, bool) {
	switch segment {
	case "bugs":
		return CategoryBug, true
	case CategoryFeature, CategoryQuestion:
		return segment, true
	default:
		return "", false
	}
}

func ValidPublicID(value string) bool {
	if len(value) != 36 {
		return false
	}
	for i := range len(value) {
		c := value[i]
		switch i {
		case 8, 13, 18, 23:
			if c != '-' {
				return false
			}
		default:
			if (c < '0' || c > '9') && (c < 'a' || c > 'f') {
				return false
			}
		}
	}
	return true
}

func FeedbackPath(slug, category, publicID string) string {
	return "/mod/" + slug + "/" + PublicCategory(category) + "/" + publicID
}

const (
	MaxAttachmentBytes = 12 << 20
	MaxAttachments     = 8
)

func ValidAttachmentDir(value string) bool {
	if value == "" || len(value) > 240 || strings.Contains(value, "\x00") {
		return false
	}
	return filepath.IsAbs(value) && !strings.Contains(value, "..")
}
