package ui

import (
	"github.com/maddalax/htmgo/framework/h"
	"github.com/maddalax/htmgo/framework/hx"
)

type InputProps struct {
	Id             string
	Label          string
	Name           string
	Type           string
	DefaultValue   string
	Placeholder    string
	Required       bool
	ValidationPath string
	Error          string
	Children       []h.Ren
}

func Input(props InputProps) *h.Element {
	validation := h.If(
		props.ValidationPath != "",
		h.Children(
			h.Post(props.ValidationPath, hx.BlurEvent),
			h.Attribute("hx-swap", "innerHTML transition:true"),
			h.Attribute("hx-target", "next div"),
		),
	)

	if props.Type == "" {
		props.Type = "text"
	}

	input := h.Input(
		props.Type,
		h.If(props.Id != "", h.Id(props.Id)),
		h.ClassX("rounded focus:outline-none focus:ring focus:ring-slate-800", map[string]bool{
			// Standard input styles
			"border p-2": props.Type != "checkbox",
			// Checkbox specific styles
			"w-4 h-4 cursor-pointer accent-slate-800": props.Type == "checkbox",
		}),
		h.If(
			props.Name != "",
			h.Name(props.Name),
		),

		h.If(
			props.Children != nil,
			h.Children(props.Children...),
		),
		h.If(
			props.Required,
			h.Required(),
		),
		h.If(
			props.Placeholder != "",
			h.Placeholder(props.Placeholder),
		),
		h.If(
			props.DefaultValue != "",
			h.Attribute("value", props.DefaultValue),
		),
		validation,
	)

	label := h.If(
		props.Label != "",
		h.Label(
			h.If(props.Id != "", h.Attribute("for", props.Id)),
			h.Text(props.Label),
		),
	)

	wrapped := h.Div(
		h.ClassX("", map[string]bool{
			"flex flex-col gap-1":           		props.Type != "checkbox",
			"flex flex-row items-center gap-2": props.Type == "checkbox",
		}),
		h.If(props.Type != "checkbox", label),
		input,
		h.If(props.Type == "checkbox", label),
		h.Div(
			h.If(props.Id != "", h.Id(props.Id+"-error")),
			h.Class("text-red-500"),
		),
	)

	return wrapped
}
