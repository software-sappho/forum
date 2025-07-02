package handlers

import (
	"database/sql"
	"forum/database"
	"forum/models"
	"net/http"
	"strconv"
	"strings"
)

// HandleUserComments displays a page with all comments from a specific user.
func HandleUserComments(w http.ResponseWriter, r *http.Request) {
	username := strings.TrimPrefix(r.URL.Path, "/user-comments/")

	// Debug log
	println("Looking up user for comments:", username)

	// Get user by username to get their ID
	var user models.User
	var profilePicture sql.NullString
	err := database.DB.QueryRow("SELECT id, username, email, profile_picture, created_at FROM users WHERE username = ?", username).Scan(&user.ID, &user.Username, &user.Email, &profilePicture, &user.CreatedAt)
	if profilePicture.Valid {
		user.ProfilePicture = profilePicture.String
	} else {
		user.ProfilePicture = ""
	}
	if err != nil {
		println("DB error:", err.Error())
		if err == sql.ErrNoRows {
			RenderErrorPage(w, r, http.StatusNotFound, "User not found")
		} else {
			RenderErrorPage(w, r, http.StatusInternalServerError, "Server Error")
		}
		return
	}

	page := 1
	if pageStr := r.URL.Query().Get("page"); pageStr != "" {
		if p, err := strconv.Atoi(pageStr); err == nil && p > 0 {
			page = p
		}
	}

	comments, totalComments, err := LoadCommentsByUserID(user.ID, page)
	if err != nil {
		RenderErrorPage(w, r, http.StatusInternalServerError, "Could not load comments")
		return
	}

	// Calculate pagination info
	commentsPerPage := 10
	totalPages := (totalComments + commentsPerPage - 1) / commentsPerPage

	data := GetPageData(r)
	data.Comments = comments
	data.User = &user
	data.CurrentPage = page
	data.TotalPages = totalPages

	RenderTemplate(w, "user-comments", data)
}

// LoadCommentsByUserID loads paginated comments for a specific user.
func LoadCommentsByUserID(userID, page int) ([]models.Comment, int, error) {
	var totalComments int
	err := database.DB.QueryRow("SELECT COUNT(*) FROM comments WHERE user_id = ?", userID).Scan(&totalComments)
	if err != nil {
		return nil, 0, err
	}

	commentsPerPage := 10
	offset := (page - 1) * commentsPerPage

	rows, err := database.DB.Query(`
		SELECT comments.id, comments.content, comments.created_at, comments.post_id,
		       posts.title as post_title,
		       users.username, users.profile_picture,
		       (SELECT COUNT(*) FROM likes WHERE comment_id = comments.id AND is_liked = 1) as like_count,
		       (SELECT COUNT(*) FROM likes WHERE comment_id = comments.id AND is_liked = 0) as dislike_count
		FROM comments
		JOIN posts ON comments.post_id = posts.id
		JOIN users ON comments.user_id = users.id
		WHERE comments.user_id = ?
		ORDER BY comments.created_at DESC
		LIMIT ? OFFSET ?
	`, userID, commentsPerPage, offset)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	var comments []models.Comment
	for rows.Next() {
		var comment models.Comment
		var profilePicture sql.NullString
		var likeCount int
		var dislikeCount int

		if err := rows.Scan(&comment.ID, &comment.Content, &comment.CreatedAt, &comment.PostID,
			&comment.PostTitle, &comment.Author, &profilePicture, &likeCount, &dislikeCount); err != nil {
			return nil, 0, err
		}

		comment.Likes = likeCount
		comment.Dislikes = dislikeCount
		comment.TimeAgo = formatTimeAgo(comment.CreatedAt)

		if profilePicture.Valid && profilePicture.String != "" {
			comment.AvatarURL = profilePicture.String
		} else {
			comment.AvatarURL = ""
		}

		comments = append(comments, comment)
	}

	return comments, totalComments, nil
}
