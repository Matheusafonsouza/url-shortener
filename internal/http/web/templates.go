package web

import "html/template"

func ParseTemplates() (*template.Template, error) {
	return template.ParseGlob("web/templates/*.html")
}
