package ui

import "github.com/maddalax/htmgo/framework/h"

func SecurityAlert(message string) *h.Element {
	return h.Div(
		h.Class("bg-red-50 border-l-4 border-red-400 p-4 mb-6"),
		h.Div(
			h.Class("flex items-center"),
			h.Div(
				h.Class("text-red-700 text-sm"),
				h.Text(message),
			),
		),
	)
}
