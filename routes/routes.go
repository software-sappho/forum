package routes

import (
	"forum/handlers"
	"forum/middleware"
	"net/http"
)

func SetupRoutes() {
	// Serve static files with no caching
	fs := http.FileServer(http.Dir("static"))
	http.Handle("/static/", http.StripPrefix("/static/", fs))
	http.Handle("/uploads/", http.StripPrefix("/uploads/", http.FileServer(http.Dir("uploads"))))

	var middlewares = []func(http.HandlerFunc) http.HandlerFunc{
		middleware.Logging,
		middleware.Recovery,
		middleware.LoggedIn,
	}

	type routeConfig struct {
		handler     http.HandlerFunc
		middlewares []func(http.HandlerFunc) http.HandlerFunc
	}

	routes := map[string]routeConfig{
		"/forum":                {handlers.HandleHome, nil},
		"/register":             {handlers.HandleRegister, []func(http.HandlerFunc) http.HandlerFunc{middleware.NoAuth}},
		"/login":                {handlers.HandleLogin, []func(http.HandlerFunc) http.HandlerFunc{middleware.NoAuth}},
		"/logout":               {handlers.HandleLogout, []func(http.HandlerFunc) http.HandlerFunc{middleware.Auth}},
		"/profile":              {handlers.HandleProfile, []func(http.HandlerFunc) http.HandlerFunc{middleware.Auth}},
		"/create-post":          {handlers.HandleCreatePost, []func(http.HandlerFunc) http.HandlerFunc{middleware.Auth}},
		"/create-category":      {handlers.HandleCreateCategory, []func(http.HandlerFunc) http.HandlerFunc{middleware.Auth}},
		"/comments":             {handlers.HandleAddComment, []func(http.HandlerFunc) http.HandlerFunc{middleware.Auth}},
		"/toggle-like":          {handlers.HandleToggleLike, []func(http.HandlerFunc) http.HandlerFunc{middleware.Auth}},
		"/toggle-dislike":       {handlers.HandleToggleDislike, []func(http.HandlerFunc) http.HandlerFunc{middleware.Auth}},
		"/posts/":               {handlers.HandlePostDetail, nil},
		"/contact":              {handlers.HandleContact, nil},
		"/privacy":              {handlers.HandlePrivacy, nil},
		"/faq":                  {handlers.HandleFaq, nil},
		"/terms":                {handlers.HandleTerms, nil},
		"/nasa-apod":            {handlers.HandleNasaApod, nil},
		"/about":                {handlers.HandleAbout, nil},
		"/api/nasa-apod":        {handlers.ApodHandler, nil},
		"/auth/github":          {handlers.HandleGitHubLogin, nil},
		"/auth/github/callback": {handlers.HandleGitHubCallback, nil},
		"/auth/google":          {handlers.HandleGoogleLogin, nil},
		"/auth/google/callback": {handlers.HandleGoogleCallback, nil},
	}

	// Apply middleware chain to each route
	for path, route := range routes {
		h := route.handler

		// Apply route-specific middleware
		if route.middlewares != nil {
			for _, m := range route.middlewares {
				h = m(h)
			}
		}

		// Apply global middleware
		for _, m := range middlewares {
			h = m(h)
		}

		http.HandleFunc(path, h)
	}

	// Register prefix handlers for user posts, likes, and comments (for variable username)
	http.HandleFunc("/user-posts/", middleware.LoggedIn(middleware.Auth(handlers.HandleUserPosts)))
	http.HandleFunc("/user-likes/", middleware.LoggedIn(middleware.Auth(handlers.HandleUserLikes)))
	http.HandleFunc("/user-comments/", middleware.LoggedIn(middleware.Auth(handlers.HandleUserComments)))

	// Catch-all handler: landing page for "/", branded 404 otherwise
	catchAll := func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/" {
			handlers.HandleLanding(w, r)
			return
		}
		handlers.RenderErrorPage(w, r, http.StatusNotFound, "Page Not Found")
	}

	// Apply global middlewares to catchAll
	h := catchAll
	for _, m := range middlewares {
		h = m(h)
	}

	http.HandleFunc("/", h)
}
