package domain

import "time"

const (
	CategoryBug      = "bug"
	CategoryFeature  = "feature"
	CategoryQuestion = "question"

	StatusOpen       = "open"
	StatusInProgress = "in_progress"
	StatusResolved   = "resolved"
	StatusClosed     = "closed"
)

type Feedback struct {
	ID             int64     `json:"id"`
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

type StatusUpdate struct {
	Status string `json:"status"`
}

func ValidCategory(category string) bool {
	switch category {
	case CategoryBug, CategoryFeature, CategoryQuestion:
		return true
	default:
		return false
	}
}

func ValidStatus(status string) bool {
	switch status {
	case StatusOpen, StatusInProgress, StatusResolved, StatusClosed:
		return true
	default:
		return false
	}
}
