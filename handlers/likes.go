package handlers

import (
	"database/sql"
	"forum/database"
	"forum/models"
	"net/http"
	"strconv"
	"strings"
)

func HandleToggleLike(w http.ResponseWriter, r *http.Request) {
	user, ok := r.Context().Value("user").(*models.User)
	if !ok || user == nil {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	if err := r.ParseForm(); err != nil {
		http.Error(w, "Invalid Data", http.StatusBadRequest)
		return
	}

	postIDStr := r.FormValue("post_id")
	commentIDStr := r.FormValue("comment_id")

	var (
		postID, commentID sql.NullInt64
	)

	if postIDStr != "" {
		if id, err := strconv.Atoi(postIDStr); err == nil {
			postID = sql.NullInt64{Int64: int64(id), Valid: true}
		}
	}

	if commentIDStr != "" {
		if id, err := strconv.Atoi(commentIDStr); err == nil {
			commentID = sql.NullInt64{Int64: int64(id), Valid: true}
		}
	}

	var exists bool
	var query string
	var args []interface{}

	if postID.Valid {
		query = "SELECT EXISTS(SELECT 1 FROM likes WHERE user_id = ? AND post_id = ? AND is_liked = 1)"
		args = []interface{}{user.ID, postID.Int64}
	} else if commentID.Valid {
		query = "SELECT EXISTS(SELECT 1 FROM likes WHERE user_id = ? AND comment_id = ? AND is_liked = 1)"
		args = []interface{}{user.ID, commentID.Int64}
	} else {
		http.Error(w, "No post_id or comment_id provided", http.StatusBadRequest)
		return
	}

	err := database.DB.QueryRow(query, args...).Scan(&exists)
	if err != nil {
		http.Error(w, "DB error", http.StatusInternalServerError)
		return
	}

	if exists {
		if postID.Valid {
			_, err = database.DB.Exec("DELETE FROM likes WHERE user_id = ? AND post_id = ? AND is_liked = 1", user.ID, postID.Int64)
		} else {
			_, err = database.DB.Exec("DELETE FROM likes WHERE user_id = ? AND comment_id = ? AND is_liked = 1", user.ID, commentID.Int64)
		}
	} else {
		// Check if user already disliked this post/comment
		var dislikedExists bool
		if postID.Valid {
			err = database.DB.QueryRow("SELECT EXISTS(SELECT 1 FROM likes WHERE user_id = ? AND post_id = ? AND is_liked = 0)", user.ID, postID.Int64).Scan(&dislikedExists)
		} else {
			err = database.DB.QueryRow("SELECT EXISTS(SELECT 1 FROM likes WHERE user_id = ? AND comment_id = ? AND is_liked = 0)", user.ID, commentID.Int64).Scan(&dislikedExists)
		}

		if err != nil {
			http.Error(w, "DB error", http.StatusInternalServerError)
			return
		}

		if dislikedExists {
			// Remove the dislike first
			if postID.Valid {
				_, err = database.DB.Exec("DELETE FROM likes WHERE user_id = ? AND post_id = ? AND is_liked = 0", user.ID, postID.Int64)
			} else {
				_, err = database.DB.Exec("DELETE FROM likes WHERE user_id = ? AND comment_id = ? AND is_liked = 0", user.ID, commentID.Int64)
			}
			if err != nil {
				http.Error(w, "DB error", http.StatusInternalServerError)
				return
			}
		}

		_, err = database.DB.Exec(
			"INSERT INTO likes (user_id, post_id, comment_id, is_liked) VALUES (?, ?, ?, 1)",
			user.ID,
			nullOrValue(postID),
			nullOrValue(commentID),
		)
	}

	if err != nil {
		http.Error(w, "DB error", http.StatusInternalServerError)
		return
	}

	referer := r.Header.Get("Referer")
	if postIDStr != "" {
		referer += "#post-" + postIDStr
	}
	http.Redirect(w, r, r.Header.Get("Referer"), http.StatusSeeOther)

	w.WriteHeader(http.StatusOK)
}

func nullOrValue(v sql.NullInt64) interface{} {
	if v.Valid {
		return v.Int64
	}
	return nil
}

func CountLikesForPost(postID int) (int, error) {
	var count int
	err := database.DB.QueryRow(`SELECT COUNT(*) FROM likes WHERE post_id = ? AND is_liked = 1`, postID).Scan(&count)
	return count, err
}

// HandleUserLikes displays a page with all posts/comments liked/disliked by a specific user.
func HandleUserLikes(w http.ResponseWriter, r *http.Request) {
	username := strings.TrimPrefix(r.URL.Path, "/user-likes/")

	// Debug log
	println("Looking up user (likes):", username)

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
		println("DB error (likes):", err.Error())
		if err == sql.ErrNoRows {
			RenderErrorPage(w, r, http.StatusNotFound, "User not found")
		} else {
			RenderErrorPage(w, r, http.StatusInternalServerError, "Server Error")
		}
		return
	}

	// Load liked posts and comments
	likedPosts, err := LoadLikedPostsByUserID(user.ID)
	if err != nil {
		RenderErrorPage(w, r, http.StatusInternalServerError, "Could not load liked posts")
		return
	}
	likedComments, err := LoadLikedCommentsByUserID(user.ID)
	if err != nil {
		RenderErrorPage(w, r, http.StatusInternalServerError, "Could not load liked comments")
		return
	}

	data := GetPageData(r)
	data.User = &user
	data.LikedPosts = likedPosts
	data.LikedComments = likedComments

	RenderTemplate(w, "user-likes", data)
}

// LoadLikedPostsByUserID loads all posts liked by a specific user.
func LoadLikedPostsByUserID(userID int) ([]models.Post, error) {
	rows, err := database.DB.Query(`
		SELECT posts.id, posts.title, posts.content, posts.category, posts.image_url, posts.created_at, users.username, users.profile_picture
		FROM likes
		JOIN posts ON likes.post_id = posts.id
		JOIN users ON posts.user_id = users.id
		WHERE likes.user_id = ? AND likes.is_liked = 1 AND likes.post_id IS NOT NULL
		ORDER BY posts.created_at DESC
	`, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var posts []models.Post
	for rows.Next() {
		var post models.Post
		var imageURL sql.NullString
		var profilePicture sql.NullString

		if err := rows.Scan(&post.ID, &post.Title, &post.Content, &post.Category, &imageURL, &post.CreatedAt, &post.Author, &profilePicture); err != nil {
			return nil, err
		}

		post.ImageURL = imageURL
		if profilePicture.Valid && profilePicture.String != "" {
			post.AvatarURL = profilePicture.String
		} else {
			post.AvatarURL = ""
		}
		posts = append(posts, post)
	}

	return posts, nil
}

// LoadLikedCommentsByUserID loads all comments liked by a specific user.
func LoadLikedCommentsByUserID(userID int) ([]models.Comment, error) {
	rows, err := database.DB.Query(`
		SELECT comments.id, comments.post_id, comments.user_id, comments.content, comments.created_at, users.profile_picture
		FROM likes
		JOIN comments ON likes.comment_id = comments.id
		JOIN users ON comments.user_id = users.id
		WHERE likes.user_id = ? AND likes.is_liked = 1 AND likes.comment_id IS NOT NULL
		ORDER BY comments.created_at DESC
	`, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var comments []models.Comment
	for rows.Next() {
		var comment models.Comment
		var profilePicture sql.NullString

		if err := rows.Scan(&comment.ID, &comment.PostID, &comment.UserID, &comment.Content, &comment.CreatedAt, &profilePicture); err != nil {
			return nil, err
		}
		if profilePicture.Valid && profilePicture.String != "" {
			comment.AvatarURL = profilePicture.String
		} else {
			comment.AvatarURL = ""
		}
		comments = append(comments, comment)
	}

	return comments, nil
}
