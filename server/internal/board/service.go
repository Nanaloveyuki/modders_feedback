package board

import (
	"strings"
	"unicode/utf8"

	"modders-feedback/server/internal/store"
)

type Service struct {
	db *store.Store
}

func New(db *store.Store) *Service {
	return &Service{db: db}
}

type Error struct {
	Status    int
	MessageZh string
	MessageEn string
}

func (e *Error) Error() string {
	return e.MessageEn
}

func withinRunes(value string, min, max int) bool {
	count := utf8.RuneCountInString(value)
	return count >= min && count <= max
}

func patchedString(current string, patch *string) string {
	if patch == nil {
		return current
	}
	return strings.TrimSpace(*patch)
}
