package auth

import (
	"errors"
	"net/http"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

const cookieName = "session"

const sessionTTL = 7 * 24 * time.Hour

type Sessions struct {
	secret []byte
	secure bool
}

func New(secret string, secure bool) Sessions {
	return Sessions{secret: []byte(secret), secure: secure}
}

func (s Sessions) Issue(w http.ResponseWriter, username string) error {
	now := time.Now()
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"sub": username,
		"exp": now.Add(sessionTTL).Unix(),
		"iat": now.Unix(),
	})
	value, err := token.SignedString(s.secret)
	if err != nil {
		return err
	}
	http.SetCookie(w, s.cookie(value, int(sessionTTL.Seconds())))
	return nil
}

func (s Sessions) Clear(w http.ResponseWriter) {
	http.SetCookie(w, s.cookie("", -1))
}

func (s Sessions) Username(r *http.Request) (string, bool) {
	cookie, err := r.Cookie(cookieName)
	if err != nil {
		return "", false
	}
	claims := jwt.MapClaims{}
	token, err := jwt.ParseWithClaims(cookie.Value, claims, func(token *jwt.Token) (any, error) {
		if token.Method != jwt.SigningMethodHS256 {
			return nil, errors.New("unexpected signing method")
		}
		return s.secret, nil
	})
	if err != nil || !token.Valid {
		return "", false
	}
	subject, ok := claims["sub"].(string)
	return subject, ok && subject != ""
}

func (s Sessions) cookie(value string, maxAge int) *http.Cookie {
	return &http.Cookie{
		Name:     cookieName,
		Value:    value,
		Path:     "/",
		HttpOnly: true,
		Secure:   s.secure,
		SameSite: http.SameSiteLaxMode,
		MaxAge:   maxAge,
	}
}
