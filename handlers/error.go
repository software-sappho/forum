package handlers

import (
	"net/http"

	"forum/models"
)

// RenderErrorPage centralizes rendering of a user-friendly error template.
// It also preserves session info so the navigation bar can reflect login state.
func RenderErrorPage(w http.ResponseWriter, r *http.Request, code int, message string) {
	// Attempt to extract user from context (set by middleware.LoggedIn).
	var (
		isLogged bool
		username string
	)
	if user, ok := r.Context().Value("user").(*models.User); ok && user != nil {
		isLogged = true
		username = user.Username
	}

	data := models.PageData{
		IsLoggedIn:   isLogged,
		Username:     username,
		ErrorCode:    code,
		ErrorMessage: message,
	}

	// Write the status code BEFORE executing the template so Go's http package
	// sets the correct response.
	w.WriteHeader(code)
	RenderTemplate(w, "error", data)
}
