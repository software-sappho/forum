package handlers

import (
	"forum/models"
	"net/http"
)

// GetPageData extracts IsLoggedIn and Username from the request context.
// Handlers should build their page-specific data on top of the struct it returns.
func GetPageData(r *http.Request) models.PageData {
	user, ok := r.Context().Value("user").(*models.User)
	if ok && user != nil {
		return models.PageData{
			IsLoggedIn: true,
			Username:   user.Username,
		}
	}
	return models.PageData{}
}
