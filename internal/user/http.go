package user

import (
	"github.com/maddalax/htmgo/framework/h"
	"advancedauth/internal/db"
)

func GetUserOrRedirect(ctx *h.RequestContext) (db.User, bool) {
	user, err := GetUserFromSession(ctx)

	if err != nil {
		// No session -> try rememberMe
		user, err = GetUserFromRememberToken(ctx)

		if err != nil {
			// No rememberMe either -> redirect to login
			// Check if this is an HTMX request
			if ctx.Request.Header.Get("HX-Request") != "" {
				// Tell HTMX to perform a full client-side redirect
				ctx.Response.Header().Set("HX-Redirect", "/login")
			} else {
				// Standard HTTP redirect for full page loads
				ctx.Redirect("/login", 302)
			}
			return db.User{}, false
		}

		// Is rememberMe but no session -> create new session
		_, err = CreateSession(ctx, user.ID)
		if err != nil {
			return db.User{}, false
		}
	}

	// Return user regardless of was it found in session or rememberMe
	return user, true
}
