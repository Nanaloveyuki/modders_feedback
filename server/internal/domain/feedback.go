package domain

import (
	"errors"
	"net/url"
	"strings"
	"time"
)

var (
	ErrLastMod       = errors.New("cannot delete the last mod")
	ErrUsernameTaken = errors.New("username already exists")
	ErrEmailTaken    = errors.New("email already exists")
	ErrBadPassword   = errors.New("current password is incorrect")
)

const (
	CategoryBug      = "bug"
	CategoryFeature  = "feature"
	CategoryQuestion = "question"

	StatusOpen       = "open"
	StatusInProgress = "in_progress"
	StatusResolved   = "resolved"
	StatusClosed     = "closed"
	StatusWithdrawn  = "withdrawn"

	DefaultModSlug   = "rhah"
	DefaultSteamURL  = "https://steamcommunity.com/sharedfiles/filedetails/?id="
	DefaultGitHubURL = "https://github.com/Nanaloveyuki/modders_feedback"
)

type Feedback struct {
	ID             int64     `json:"id"`
	ModID          int64     `json:"modId"`
	Category       string    `json:"category"`
	CategoryNumber int64     `json:"categoryNumber"`
	Title          string    `json:"title"`
	Body           string    `json:"body"`
	Author         string    `json:"author"`
	GameVersion    string    `json:"gameVersion"`
	ModVersion     string    `json:"modVersion"`
	ModList        string    `json:"modList"`
	SaveLink       string    `json:"saveLink"`
	Status         string    `json:"status"`
	CreatedAt      time.Time `json:"createdAt"`
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
type FeedbackUpdate struct {
	Title       string `json:"title"`
	Body        string `json:"body"`
	GameVersion string `json:"gameVersion"`
	ModVersion  string `json:"modVersion"`
	ModList     string `json:"modList"`
	SaveLink    string `json:"saveLink"`
}

type SiteSettings struct {
	ModVersion  string `json:"modVersion"`
	GameVersion string `json:"gameVersion"`
	Icon        string `json:"icon"`
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

func ValidStatus(status string) bool {
	switch status {
	case StatusOpen, StatusInProgress, StatusResolved, StatusClosed, StatusWithdrawn:
		return true
	default:
		return false
	}
}
