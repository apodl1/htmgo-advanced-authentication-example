package pages

import (
	"github.com/maddalax/htmgo/framework/h"
	"net/http"

	"advancedauth/partials"
	"advancedauth/ui"
)

func Login(ctx *h.RequestContext) *h.Page {
 // Check for the security alert cookie
    alertCookie, _ := ctx.Request.Cookie("security_alert")
    hasAlert := alertCookie != nil && alertCookie.Value == "stolen_token"

    // If we found it, clear it so it doesn't show again on refresh
    if hasAlert {
        ctx.SetCookie(&http.Cookie{Name: "security_alert", MaxAge: -1, Path: "/"})
    }

	return h.NewPage(
		RootPage(
			ui.CenteredForm(ui.CenteredFormProps{
				Title:      "Sign In",
				SubmitText: "Sign In",
				PostUrl:    h.GetPartialPath(partials.LoginUser),
				Children: []h.Ren{
				 h.If(hasAlert, ui.SecurityAlert("A potential security issue was detected with your session. For your protection, you have been logged out of all devices. Please sign in again.")),

					ui.Input(ui.InputProps{
						Id:       "username",
						Name:     "email",
						Label:    "Email Address",
						Type:     "email",
						Required: true,
						Children: []h.Ren{
							h.Attribute("autocomplete", "off"),
							h.MaxLength(50),
						},
					}),

					ui.Input(ui.InputProps{
						Id:       "password",
						Name:     "password",
						Label:    "Password",
						Type:     "password",
						Required: true,
						Children: []h.Ren{
							h.MinLength(6),
						},
					}),

					ui.Input(ui.InputProps{
						Type: 	  "checkbox",
						Id:       "remember",
						Name:     "remember",
						Label:    "Remember me",
						Required: false,
					}),

					h.A(
						h.Href("/register"),
						h.Text("Don't have an account? Register here"),
						h.Class("text-blue-500"),
					),
				},
			}),
		),
	)
}
