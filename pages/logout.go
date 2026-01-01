package pages

import (
	"context"
	"net/http"

	"github.com/maddalax/htmgo/framework/h"
	"github.com/maddalax/htmgo/framework/service"

	"advancedauth/internal/db"
	"advancedauth/internal/user"
)

func LogoutPage(ctx *h.RequestContext) *h.Page {

	// Delete the session from the database
	cookie, err := ctx.Request.Cookie("session_id")
	if err == nil {
		queries := service.Get[db.Queries](ctx.ServiceLocator())
		queries.DeleteSessionByID(context.Background(), cookie.Value)
	}

	// clear the session cookie
	ctx.SetCookie(&http.Cookie{
        Name:     "session_id",
        Value:    "",
        Path:     "/",
        MaxAge:   -1,
        HttpOnly: true,
    })

	// Delete the series from the database
	cookie, err = ctx.Request.Cookie("remember_me")
	if err == nil {
		selectorFromCookie, _, err := user.ParseSelectorAndValidator(cookie.Value)
		if err == nil {
			queries := service.Get[db.Queries](ctx.ServiceLocator())
			queries.DeleteRememberTokenBySelector(context.Background(), selectorFromCookie)
		}
	}

	// clear the remember cookie
	ctx.SetCookie(&http.Cookie{
        Name:     "remember_me",
        Value:    "",
        Path:     "/",
        MaxAge:   -1,
        HttpOnly: true,
    })

	ctx.Response.Header().Set(
		"Location",
		"/login",
	)

	ctx.Response.WriteHeader(
		302,
	)

	return nil
}
