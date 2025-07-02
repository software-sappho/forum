package handlers

import (
	"database/sql"
	"forum/database"
	"forum/models"
	"net/http"
	"strconv"
)

func HandleToggleDislike(w http.ResponseWriter, r *http.Request) {
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
		query = "SELECT EXISTS(SELECT 1 FROM likes WHERE user_id = ? AND post_id = ? AND is_liked = 0)"
		args = []interface{}{user.ID, postID.Int64}
	} else if commentID.Valid {
		query = "SELECT EXISTS(SELECT 1 FROM likes WHERE user_id = ? AND comment_id = ? AND is_liked = 0)"
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
		// Remove dislike
		if postID.Valid {
			_, err = database.DB.Exec("DELETE FROM likes WHERE user_id = ? AND post_id = ? AND is_liked = 0", user.ID, postID.Int64)
		} else {
			_, err = database.DB.Exec("DELETE FROM likes WHERE user_id = ? AND comment_id = ? AND is_liked = 0", user.ID, commentID.Int64)
		}
	} else {
		// Check if user already liked this post/comment
		var likedExists bool
		if postID.Valid {
			err = database.DB.QueryRow("SELECT EXISTS(SELECT 1 FROM likes WHERE user_id = ? AND post_id = ? AND is_liked = 1)", user.ID, postID.Int64).Scan(&likedExists)
		} else {
			err = database.DB.QueryRow("SELECT EXISTS(SELECT 1 FROM likes WHERE user_id = ? AND comment_id = ? AND is_liked = 1)", user.ID, commentID.Int64).Scan(&likedExists)
		}

		if err != nil {
			http.Error(w, "DB error", http.StatusInternalServerError)
			return
		}

		if likedExists {
			// Remove the like first
			if postID.Valid {
				_, err = database.DB.Exec("DELETE FROM likes WHERE user_id = ? AND post_id = ? AND is_liked = 1", user.ID, postID.Int64)
			} else {
				_, err = database.DB.Exec("DELETE FROM likes WHERE user_id = ? AND comment_id = ? AND is_liked = 1", user.ID, commentID.Int64)
			}
			if err != nil {
				http.Error(w, "DB error", http.StatusInternalServerError)
				return
			}
		}

		// Add dislike
		_, err = database.DB.Exec(
			"INSERT INTO likes (user_id, post_id, comment_id, is_liked) VALUES (?, ?, ?, 0)",
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

func CountDislikesForPost(postID int) (int, error) {
	var count int
	err := database.DB.QueryRow(`SELECT COUNT(*) FROM likes WHERE post_id = ? AND is_liked = 0`, postID).Scan(&count)
	return count, err
}

func CountDislikesForComment(commentID int) (int, error) {
	var count int
	err := database.DB.QueryRow(`SELECT COUNT(*) FROM likes WHERE comment_id = ? AND is_liked = 0`, commentID).Scan(&count)
	return count, err
}
