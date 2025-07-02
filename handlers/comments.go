package handlers

import (
	"database/sql"
	"forum/database"
	"forum/models"
	"net/http"
	"strconv"
)

func HandleAddComment(w http.ResponseWriter, r *http.Request) {
	user, ok := r.Context().Value("user").(*models.User)
	if !ok || user == nil {
		http.Redirect(w, r, "/login", http.StatusSeeOther)
		return
	}

	if r.Method != http.MethodPost {
		http.Error(w, "Invalid request method", http.StatusMethodNotAllowed)
		return
	}

	postIDStr := r.FormValue("post_id")
	content := r.FormValue("content")

	postID, err := strconv.Atoi(postIDStr)
	if err != nil || content == "" {
		http.Error(w, "Invalid comment data", http.StatusBadRequest)
		return
	}

	_, err = database.DB.Exec(`
		INSERT INTO comments (post_id, user_id, content, author)
		VALUES (?, ?, ?, ?)
	`, postID, user.ID, content, user.Username)
	if err != nil {
		http.Error(w, "Failed to save comment", http.StatusInternalServerError)
		return
	}

	http.Redirect(w, r, "/posts/"+postIDStr, http.StatusSeeOther)
}

func LoadCommentsForPost(postID int, user *models.User) ([]models.Comment, error) {
	rows, err := database.DB.Query(`
		SELECT comments.id, comments.post_id, comments.user_id, comments.content, comments.created_at, comments.author, users.profile_picture,
			(SELECT COUNT(*) FROM likes WHERE comment_id = comments.id AND is_liked = 1) as like_count,
			(SELECT COUNT(*) FROM likes WHERE comment_id = comments.id AND is_liked = 0) as dislike_count
		FROM comments
		JOIN users ON comments.user_id = users.id
		WHERE comments.post_id = ?
		ORDER BY comments.created_at ASC
	`, postID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var comments []models.Comment
	for rows.Next() {
		var c models.Comment
		var likeCount int
		var dislikeCount int
		var profilePicture sql.NullString
		if err := rows.Scan(&c.ID, &c.PostID, &c.UserID, &c.Content, &c.CreatedAt, &c.Author, &profilePicture, &likeCount, &dislikeCount); err != nil {
			return nil, err
		}

		// Check if current user liked/disliked this comment
		if user != nil {
			err := database.DB.QueryRow(
				"SELECT EXISTS (SELECT 1 FROM likes WHERE user_id = ? AND comment_id = ? AND is_liked = 1)",
				user.ID, c.ID,
			).Scan(&c.UserLiked)
			if err != nil {
				return nil, err
			}

			err = database.DB.QueryRow(
				"SELECT EXISTS (SELECT 1 FROM likes WHERE user_id = ? AND comment_id = ? AND is_liked = 0)",
				user.ID, c.ID,
			).Scan(&c.UserDisliked)
			if err != nil {
				return nil, err
			}
		}

		c.Likes = likeCount
		c.Dislikes = dislikeCount

		// Set the avatar URL for the comment
		if profilePicture.Valid && profilePicture.String != "" {
			c.AvatarURL = profilePicture.String
		} else {
			c.AvatarURL = ""
		}

		comments = append(comments, c)
	}
	return comments, nil
}
