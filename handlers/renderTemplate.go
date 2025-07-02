package handlers

import (
	"forum/models"
	"html/template"
	"net/http"
)

// Create template functions map
var funcMap = template.FuncMap{
	"add": func(a, b int) int {
		return a + b
	},
	"subtract": func(a, b int) int {
		return a - b
	},
	"truncate": func(s string, n int) string {
		if len(s) <= n {
			return s
		}
		return s[:n] + "..."
	},
}

func RenderTemplate(w http.ResponseWriter, tmpl string, data models.PageData) {
	var files []string
	switch tmpl {
	case "landing":
		files = []string{"templates/landing.html"}
	default:
		files = []string{"templates/base.html", "templates/" + tmpl + ".html"}
	}

	// Parse templates with functions
	var templates *template.Template
	var err error

	if tmpl == "landing" {
		templates, err = template.New("landing.html").Funcs(funcMap).ParseFiles(files...)
	} else {
		templates, err = template.New("base.html").Funcs(funcMap).ParseFiles(files...)
	}

	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// Execute the template
	if tmpl == "landing" {
		err = templates.ExecuteTemplate(w, "landing.html", data)
	} else {
		err = templates.Execute(w, data)
	}

	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}
