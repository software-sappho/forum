package handlers

import (
	"database/sql"
	"fmt"
	"forum/database"
	"forum/models"
	"io"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"
)

var allowedTypes = map[string]bool{
	"image/jpeg": true,
	"image/png":  true,
	"image/gif":  true,
}

// HandleProfile handles both GET and POST requests for user profile management
func HandleProfile(w http.ResponseWriter, r *http.Request) {
	user, _ := r.Context().Value("user").(*models.User)

	if r.Method == http.MethodGet {
		// Load user's current profile data - try with profile_picture first
		var currentUser models.User
		err := database.DB.QueryRow(`
			SELECT id, username, email, profile_picture, created_at 
			FROM users WHERE id = ?
		`, user.ID).Scan(&currentUser.ID, &currentUser.Username, &currentUser.Email, &currentUser.ProfilePicture, &currentUser.CreatedAt)

		if err != nil {
			// If that fails, try without profile_picture column
			err = database.DB.QueryRow(`
				SELECT id, username, email, created_at 
				FROM users WHERE id = ?
			`, user.ID).Scan(&currentUser.ID, &currentUser.Username, &currentUser.Email, &currentUser.CreatedAt)

			if err != nil {
				log.Printf("Failed to load profile for user %d: %v", user.ID, err)
				RenderErrorPage(w, r, http.StatusInternalServerError, "Failed to load profile")
				return
			}
			// Set empty profile picture if column doesn't exist
			currentUser.ProfilePicture = ""
		}

		// Get user's posts count
		var postsCount int
		err = database.DB.QueryRow("SELECT COUNT(*) FROM posts WHERE user_id = ?", user.ID).Scan(&postsCount)
		if err != nil {
			postsCount = 0
		}

		// Get user's comments count
		var commentsCount int
		err = database.DB.QueryRow("SELECT COUNT(*) FROM comments WHERE user_id = ?", user.ID).Scan(&commentsCount)
		if err != nil {
			commentsCount = 0
		}

		// Get user's likes given count
		var likesGiven int
		err = database.DB.QueryRow("SELECT COUNT(*) FROM likes WHERE user_id = ? AND is_liked = 1", user.ID).Scan(&likesGiven)
		if err != nil {
			likesGiven = 0
		}

		// Calculate days active (days since account creation)
		daysActive := int(time.Since(currentUser.CreatedAt).Hours() / 24)
		if daysActive < 1 {
			daysActive = 1
		}

		data := models.PageData{
			IsLoggedIn:    true,
			Username:      user.Username,
			User:          &currentUser,
			PostsCount:    postsCount,
			CommentsCount: commentsCount,
			LikesGiven:    likesGiven,
			DaysActive:    daysActive,
		}
		RenderTemplate(w, "profile", data)
		return
	}

	if r.Method == http.MethodPost {
		const maxProfileImage = 500 * 1024 // 500 KB
		if err := r.ParseMultipartForm(maxProfileImage); err != nil {
			if err.Error() == "http: request body too large" {
				RenderErrorPage(w, r, http.StatusRequestEntityTooLarge, "Profile picture too big (max 500 KB)")
				return
			}
			RenderErrorPage(w, r, http.StatusBadRequest, "Invalid form data")
			return
		}

		file, header, err := r.FormFile("profile_picture")
		if err != nil {
			if err == http.ErrMissingFile {
				http.Redirect(w, r, "/profile", http.StatusSeeOther)
				return
			}
			RenderErrorPage(w, r, http.StatusBadRequest, "Error retrieving file")
			return
		}
		defer file.Close()

		// Check file type by reading the first 512 bytes
		buffer := make([]byte, 512)
		_, err = file.Read(buffer)
		if err != nil {
			RenderErrorPage(w, r, http.StatusInternalServerError, "Failed to read file for type checking")
			return
		}
		file.Seek(0, 0) // Reset read pointer

		contentType := http.DetectContentType(buffer)
		if !allowedTypes[contentType] {
			RenderErrorPage(w, r, http.StatusBadRequest, "Invalid file type. Only JPG, PNG, and GIF are allowed.")
			return
		}

		if header.Size > maxProfileImage {
			RenderErrorPage(w, r, http.StatusRequestEntityTooLarge, "Profile picture too big (max 500 KB)")
			return
		}

		if err := os.MkdirAll("uploads", 0755); err != nil {
			RenderErrorPage(w, r, http.StatusInternalServerError, "Cannot create uploads dir")
			return
		}

		fname := fmt.Sprintf("uploads/profile_%d_%s", time.Now().UnixNano(), filepath.Base(header.Filename))
		dst, err := os.Create(fname)
		if err != nil {
			RenderErrorPage(w, r, http.StatusInternalServerError, "Failed saving profile picture")
			return
		}
		defer dst.Close()

		if _, err := io.Copy(dst, file); err != nil {
			RenderErrorPage(w, r, http.StatusInternalServerError, "Failed to save file")
			return
		}

		// Delete old profile picture if it exists
		var oldPicture sql.NullString
		err = database.DB.QueryRow("SELECT profile_picture FROM users WHERE id = ?", user.ID).Scan(&oldPicture)
		if err == nil && oldPicture.Valid && oldPicture.String != "" {
			if strings.HasPrefix(oldPicture.String, "uploads/") {
				os.Remove(oldPicture.String)
			}
		}

		// Update with new profile picture
		_, err = database.DB.Exec("UPDATE users SET profile_picture = ? WHERE id = ?", fname, user.ID)
		if err != nil {
			RenderErrorPage(w, r, http.StatusInternalServerError, "Failed to update profile")
			return
		}

		http.Redirect(w, r, "/profile", http.StatusSeeOther)
		return
	}

	RenderErrorPage(w, r, http.StatusMethodNotAllowed, "Invalid method")
}
