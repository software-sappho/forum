package middleware

import (
	"context"
	"fmt"
	"forum/database"
	"forum/models"
	"log"
	"net/http"
	"time"
)

// Auth middleware checks if user is authenticated
func Auth(next http.HandlerFunc) http.HandlerFunc {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		user, ok := r.Context().Value("user").(*models.User)
		if !ok || user == nil {
			http.Redirect(w, r, "/login", http.StatusSeeOther)
			fmt.Println("redirecting to login: user not in context")
			return
		}
		next.ServeHTTP(w, r)
	})
}

// RequireNoAuth middleware prevents logged-in users from accessing certain routes
func NoAuth(next http.HandlerFunc) http.HandlerFunc {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		loggedIn, _ := IsLoggedIn(r)
		if loggedIn {
			http.Redirect(w, r, "/forum", http.StatusSeeOther)
			fmt.Printf("User already logged in, redirecting to /forum\n")
			return
		}
		next.ServeHTTP(w, r)
	})
}

func LoggedIn(next http.HandlerFunc) http.HandlerFunc {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		user, err := GetUserFromSession(r)
		if err == nil && user != nil {
			ctx := context.WithValue(r.Context(), "user", user)
			r = r.WithContext(ctx)
		}
		next.ServeHTTP(w, r)
	})
}

func GetUserFromSession(r *http.Request) (*models.User, error) {
	cookie, err := r.Cookie("session")
	if err != nil {
		return nil, err // no session cookie present
	}

	var user models.User
	query := `
		SELECT users.id, users.username, users.email
		FROM sessions
		JOIN users ON sessions.user_id = users.id
		WHERE sessions.session_token = ? AND sessions.expires_at > ?
	`
	// Query user info only if session token is valid and not expired
	err = database.DB.QueryRow(query, cookie.Value, time.Now()).Scan(&user.ID, &user.Username, &user.Email)
	if err != nil {
		log.Printf("Error getting user from session: %v", err)
		return nil, err
	}

	return &user, nil
}

func IsLoggedIn(r *http.Request) (bool, *models.User) {
	user, err := GetUserFromSession(r)
	if err != nil {
		return false, nil
	}
	return true, user
}
