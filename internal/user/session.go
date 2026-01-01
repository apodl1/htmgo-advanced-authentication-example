package user

import (
	"crypto/rand"
	"encoding/hex"
	"github.com/maddalax/htmgo/framework/h"
	"github.com/maddalax/htmgo/framework/service"
	"net/http"
	"advancedauth/internal/db"
	"time"
)

type Session struct {
	Id         string
	Expiration time.Time
	UserId     int64
}

func CreateSession(ctx *h.RequestContext, userId int64) (Session, error) {
	sessionId, err := GenerateSessionID()

	if err != nil {
		return Session{}, err
	}

	// create a session in the database
	queries := service.Get[db.Queries](ctx.ServiceLocator())

	const sessionTime = time.Hour * 2
	created := Session{
		Id:         sessionId,
		Expiration: time.Now().UTC().Add(sessionTime),
		UserId:     userId,
	}

	err = queries.CreateSession(ctx.Request.Context(), db.CreateSessionParams{
		UserID:    created.UserId,
		SessionID: created.Id,
		ExpiresAt: created.Expiration.Format(time.RFC3339),
	})

	if err != nil {
		return Session{}, err
	}

	SetSessionCookie(ctx, created)

	return created, nil
}

func GetUserFromSession(ctx *h.RequestContext) (db.User, error) {
	cookie, err := ctx.Request.Cookie("session_id")
	if err != nil {
		return db.User{}, err
	}
	queries := service.Get[db.Queries](ctx.ServiceLocator())
	user, err := queries.GetUserBySessionID(ctx.Request.Context(), cookie.Value)
	if err != nil {
		return db.User{}, err
	}
	return user, nil
}

func SetSessionCookie(ctx *h.RequestContext, session Session) {
	cookie := http.Cookie{
		Name:     "session_id",
		Value:    session.Id,
		HttpOnly: true,
		Secure: true,
		SameSite: http.SameSiteStrictMode,
		Expires:  session.Expiration,
		Path:     "/",
	}
	ctx.SetCookie(&cookie)
}

func GenerateSessionID() (string, error) {
	// Create a byte slice for storing the random bytes
	bytes := make([]byte, 32) // 32 bytes = 256 bits, which is a secure length
	// Read random bytes from crypto/rand
	if _, err := rand.Read(bytes); err != nil {
		return "", err
	}
	// Encode to hexadecimal to get a string representation
	return hex.EncodeToString(bytes), nil
}
