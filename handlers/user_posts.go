package handlers

import (
	"database/sql"
	"forum/database"
	"forum/models"
	"net/http"
	"strconv"
	"strings"
)

// HandleUserPosts displays a page with all posts from a specific user.
func HandleUserPosts(w http.ResponseWriter, r *http.Request) {
	username := strings.TrimPrefix(r.URL.Path, "/user-posts/")

	// Debug log
	println("Looking up user:", username)

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

	posts, totalPosts, err := LoadPostsByUserID(user.ID, page)
	if err != nil {
		RenderErrorPage(w, r, http.StatusInternalServerError, "Could not load posts")
		return
	}

	// Calculate pagination info
	postsPerPage := 5
	totalPages := (totalPosts + postsPerPage - 1) / postsPerPage

	data := GetPageData(r)
	data.Posts = posts
	data.User = &user
	data.CurrentPage = page
	data.TotalPages = totalPages

	RenderTemplate(w, "user-posts", data)
}

// LoadPostsByUserID loads paginated posts for a specific user with their categories.
func LoadPostsByUserID(userID, page int) ([]models.Post, int, error) {
	var totalPosts int
	err := database.DB.QueryRow("SELECT COUNT(*) FROM posts WHERE user_id = ?", userID).Scan(&totalPosts)
	if err != nil {
		return nil, 0, err
	}

	postsPerPage := 5
	offset := (page - 1) * postsPerPage

	rows, err := database.DB.Query(`
		SELECT posts.id, posts.title, posts.content, posts.category, posts.image_url, posts.created_at, users.username, users.profile_picture,
			(SELECT COUNT(*) FROM likes WHERE post_id = posts.id AND is_liked = 1) as like_count,
			(SELECT COUNT(*) FROM likes WHERE post_id = posts.id AND is_liked = 0) as dislike_count,
			(SELECT COUNT(*) FROM comments WHERE post_id = posts.id) as comment_count
		FROM posts
		JOIN users ON posts.user_id = users.id
		WHERE posts.user_id = ?
		ORDER BY posts.created_at DESC
		LIMIT ? OFFSET ?
	`, userID, postsPerPage, offset)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	var posts []models.Post
	for rows.Next() {
		var post models.Post
		var imageURL sql.NullString
		var profilePicture sql.NullString
		var likeCount int
		var dislikeCount int
		var commentCount int

		if err := rows.Scan(&post.ID, &post.Title, &post.Content, &post.Category, &imageURL, &post.CreatedAt, &post.Author, &profilePicture, &likeCount, &dislikeCount, &commentCount); err != nil {
			return nil, 0, err
		}

		// Load categories for this post
		categoryRows, err := database.DB.Query(`
			SELECT categories.name 
			FROM categories 
			JOIN post_categories ON categories.id = post_categories.category_id 
			WHERE post_categories.post_id = ?
			ORDER BY categories.name
		`, post.ID)
		if err != nil {
			return nil, 0, err
		}

		var categories []string
		for categoryRows.Next() {
			var categoryName string
			if err := categoryRows.Scan(&categoryName); err != nil {
				categoryRows.Close()
				return nil, 0, err
			}
			categories = append(categories, categoryName)
		}
		categoryRows.Close()

		// If no categories found in post_categories table, use the main category
		if len(categories) == 0 {
			categories = []string{post.Category}
		}
		post.Categories = categories

		post.Likes = likeCount
		post.Dislikes = dislikeCount
		post.Comments = commentCount
		post.TimeAgo = formatTimeAgo(post.CreatedAt)

		post.ImageURL = imageURL
		if profilePicture.Valid && profilePicture.String != "" {
			post.AvatarURL = profilePicture.String
		} else {
			post.AvatarURL = "" // Let the template handle the default
		}
		posts = append(posts, post)
	}

	return posts, totalPosts, nil
}
