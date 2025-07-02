package handlers

import (
	"forum/models"
	"net/http"
	"strconv"
	"strings"
)

func HandleLanding(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path != "/" {
		RenderErrorPage(w, r, http.StatusNotFound, "Page Not Found")
		return
	}

	data := GetPageData(r)

	RenderTemplate(w, "landing", data)
}

func HandleHome(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path != "/forum" {
		RenderErrorPage(w, r, http.StatusNotFound, "Page Not Found")
		return
	}

	// Get page number from query parameters
	page := 1
	if pageStr := r.URL.Query().Get("page"); pageStr != "" {
		if p, err := strconv.Atoi(pageStr); err == nil && p > 0 {
			page = p
		}
	}

	var posts []models.Post
	var totalPosts int
	var err error
	user, _ := r.Context().Value("user").(*models.User)
	category := strings.TrimSpace(r.URL.Query().Get("category"))
	if category == "" {
		posts, totalPosts, err = LoadPostsFromDB(page, user)
		if err != nil {
			RenderErrorPage(w, r, http.StatusInternalServerError, "Could not load posts")
			return
		}
	} else {
		posts, totalPosts, err = LoadPostsFromDBByCategory(category, page, user)
		if err != nil {
			RenderErrorPage(w, r, http.StatusInternalServerError, "Could not load posts for category")
			return
		}
	}

	categories, err := LoadCategoriesFromDB()
	if err != nil {
		RenderErrorPage(w, r, http.StatusInternalServerError, "Could not load categories")
		return
	}

	// Get selected category from query parameters
	selectedCategory := r.URL.Query().Get("category")

	// Calculate pagination info
	postsPerPage := 5
	totalPages := (totalPosts + postsPerPage - 1) / postsPerPage

	// Build Next page URL
	query := r.URL.Query()
	query.Set("page", strconv.Itoa(page+1))
	nextURL := "/forum?" + query.Encode()

	// Build Previous page URL
	query.Set("page", strconv.Itoa(page-1))
	prevURL := "/forum?" + query.Encode()

	data := GetPageData(r)
	data.Posts = posts
	data.Categories = categories
	data.SelectedCategory = selectedCategory
	data.CurrentPage = page
	data.TotalPages = totalPages
	data.NextURL = nextURL
	data.PrevURL = prevURL

	RenderTemplate(w, "home", data)
}

// NASA Astronomy Picture of the Day page
func HandleNasaApod(w http.ResponseWriter, r *http.Request) {
	user, _ := r.Context().Value("user").(*models.User)

	data := models.PageData{
		IsLoggedIn: true,
		Username:   user.Username,
	}

	RenderTemplate(w, "nasa-apod", data)
}
