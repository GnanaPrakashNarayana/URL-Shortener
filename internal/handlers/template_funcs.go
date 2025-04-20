package handlers

import (
	"html/template"
	"time"
)

// GetTemplateFuncs returns a FuncMap of custom template functions
func GetTemplateFuncs() template.FuncMap {
	return template.FuncMap{
		"add": func(a, b int) int {
			return a + b
		},
		"formatExpiryDate": func(t *time.Time) string {
			if t == nil {
				return "Never"
			}
			return t.Format("Jan 02, 2006 15:04")
		},
		"timeUntil": func(t *time.Time) string {
			if t == nil {
				return "Never"
			}

			duration := time.Until(*t)
			if duration < 0 {
				return "Expired"
			}

			days := int(duration.Hours() / 24)
			hours := int(duration.Hours()) % 24

			if days > 0 {
				return time.Until(*t).Round(time.Hour).String()
			} else if hours > 0 {
				return time.Until(*t).Round(time.Minute).String()
			} else {
				return time.Until(*t).Round(time.Second).String()
			}
		},
		"hasExpired": func(t *time.Time) bool {
			if t == nil {
				return false
			}
			return time.Now().After(*t)
		},
	}
}