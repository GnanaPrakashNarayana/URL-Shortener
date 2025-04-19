package handlers

import (
	"html/template"
)

// GetTemplateFuncs returns a FuncMap of custom template functions
func GetTemplateFuncs() template.FuncMap {
	return template.FuncMap{
		"add": func(a, b int) int {
			return a + b
		},
	}
}