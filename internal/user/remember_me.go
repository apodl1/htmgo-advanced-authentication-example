package user

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"crypto/subtle"
	"encoding/hex"
	"github.com/maddalax/htmgo/framework/h"
	"github.com/maddalax/htmgo/framework/service"
	"net/http"
	"advancedauth/internal/db"
	"time"
	"strings"
	"errors"
)


type RememberToken struct {
	Selector   string
	Validator  string
	ValidatorHash string
	Expiration time.Time
	UserId     int64
}

const rememberTime = time.Hour * 24 * 30


func CreateRememberToken(ctx *h.RequestContext, userId int64) (RememberToken, error) {
	selector, err := GenerateSelector()
	if err != nil {
		return RememberToken{}, err
	}

	validator, err := GenerateValidator()
	if err != nil {
		return RememberToken{}, err
	}

	hashedValidator := HashValidator(validator)

	// Create a rememberToken in the database
	queries := service.Get[db.Queries](ctx.ServiceLocator())

	created := RememberToken{
		Selector:   selector,
		Validator:  validator,
		ValidatorHash: hashedValidator,
		Expiration: time.Now().UTC().Add(rememberTime),
		UserId:     userId,
	}

	err = queries.CreateRememberToken(ctx.Request.Context(), db.CreateRememberTokenParams{
		UserID: created.UserId,
		Selector:    created.Selector,
		ValidatorHash: created.ValidatorHash,
		ExpiresAt: created.Expiration.Format(time.RFC3339),
	})

	if err != nil {
		return RememberToken{}, err
	}

	SetRememberMeCookie(ctx, created)

	return created, nil
}


func GetUserFromRememberToken(ctx *h.RequestContext) (db.User, error) {
	cookie, err := ctx.Request.Cookie("remember_me")
	if err != nil {
		return db.User{}, err
	}
	selectorFromCookie, validatorFromCookie, err := ParseSelectorAndValidator(cookie.Value)
	if err != nil {
		return db.User{}, err
	}

	queries := service.Get[db.Queries](ctx.ServiceLocator())
	row, err := queries.GetUserAndValidatorBySelector(ctx.Request.Context(), selectorFromCookie)
	if err != nil {
		return db.User{}, err
	}

 user := db.User{
        ID:        row.ID,
        Email:     row.Email,
        Password:  row.Password,
        Metadata:  row.Metadata,
        CreatedAt: row.CreatedAt,
        UpdatedAt: row.UpdatedAt,
    }

  validatorHashFromDB := row.ValidatorHash

  if !ValidatorMatches(validatorFromCookie, validatorHashFromDB) {
    // Not matching validator -> likely cookie has been stolen -> log the user out
 	  queries.DeleteAllUserSessions(context.Background(), user.ID)
    queries.DeleteAllUserRememberTokens(context.Background(), user.ID)

    // Set a "Flash Cookie" for the notification
    ctx.SetCookie(&http.Cookie{
        Name:     "security_alert",
        Value:    "stolen_token",
        Path:     "/",
        HttpOnly: true,
        Secure:   true,
        MaxAge:   60, // Expires in 1 minute
    })

    return db.User{}, errors.New("validator does not match!")
  }

  // Matching newValidator -> all good -> rotate newValidator and give user the new one
 	newValidator, err := GenerateValidator()
	if err != nil {
		return db.User{}, err
	}

	hashedNewValidator := HashValidator(newValidator)

	updated := RememberToken{
		Selector:   selectorFromCookie,
		Validator:  newValidator,
		ValidatorHash: hashedNewValidator,
		Expiration: time.Now().UTC().Add(rememberTime),
		UserId:     user.ID,
	}

	err = queries.RotateRememberToken(ctx.Request.Context(), db.RotateRememberTokenParams{
		Selector:    updated.Selector,
		ValidatorHash: updated.ValidatorHash,
		ExpiresAt: updated.Expiration.Format(time.RFC3339),
	})

	if err != nil {
		return db.User{}, err
	}

	SetRememberMeCookie(ctx, updated)

	return user, nil
}


func SetRememberMeCookie(ctx *h.RequestContext, rememberToken RememberToken) {
	value := rememberToken.Selector + ":" + rememberToken.Validator

	cookie := http.Cookie{
		Name:     "remember_me",
		Value:    value,
		HttpOnly: true,
		Secure:   true,
		SameSite: http.SameSiteStrictMode,
		Expires:  rememberToken.Expiration,
		Path:     "/",
	}
	ctx.SetCookie(&cookie)
}


func GenerateSelector() (string, error) {
	selectorBytes := make([]byte, 12)  // 12 bytes = 24 hex chars, enough for a unique identifier

	if _, err := rand.Read(selectorBytes); err != nil {
		return "", err
	}

	selector := hex.EncodeToString(selectorBytes)

	return selector, nil
}


func GenerateValidator() (string, error) {
	validatorBytes := make([]byte, 32) // 32 bytes = 64 hex chars, effectively impossible to brute force

	if _, err := rand.Read(validatorBytes); err != nil {
		return "", err
	}
	validator := hex.EncodeToString(validatorBytes)

	return validator, nil
}


func HashValidator(validator string) string {
		hash := sha256.Sum256([]byte(validator))
		return hex.EncodeToString(hash[:])
}


func ParseSelectorAndValidator(value string) (string, string, error) {
	parts := strings.Split(value, ":")
	if len(parts) != 2 {
		return "", "", errors.New("invalid cookie value")
	}
	return parts[0], parts[1], nil
}


func ValidatorMatches(validator string, hashedValidator string) bool {
	hash := sha256.Sum256([]byte(validator))
	comparisonHash, _ := hex.DecodeString(hashedValidator)
	return subtle.ConstantTimeCompare(hash[:], comparisonHash) == 1
}
