package handlers

import (
	"database/sql"
	"fmt"
	"forum/database"
	"forum/models"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"
)

// HandleCreatePost handles both GET and POST requests for post creation.
func HandleCreatePost(w http.ResponseWriter, r *http.Request) {
	user, _ := r.Context().Value("user").(*models.User)

	if r.Method == http.MethodGet {
		// Load categories for the form
		categories, err := LoadCategoriesFromDB()
		if err != nil {
			RenderErrorPage(w, r, http.StatusInternalServerError, "Failed to load categories")
			return
		}

		data := models.PageData{
			Username:   user.Username,
			IsLoggedIn: user != nil,
			Categories: categories,
		}
		RenderTemplate(w, "create-post", data)
		return
	}

	if r.Method == http.MethodPost {
		// 20 MB user limit
		const maxImage = 20 << 20 // 20 MiB

		mr, err := r.MultipartReader()
		if err != nil {
			RenderErrorPage(w, r, http.StatusBadRequest, "Malformed form data")
			return
		}

		var (
			title, content string
			imagePath      string
			categories     []string
		)

		for {
			part, err := mr.NextPart()
			if err == io.EOF {
				break
			}
			if err != nil {
				RenderErrorPage(w, r, http.StatusBadRequest, "Error reading form")
				return
			}

			switch part.FormName() {
			case "title":
				b, _ := io.ReadAll(io.LimitReader(part, 64*1024))
				title = strings.TrimSpace(string(b))
			case "content":
				b, _ := io.ReadAll(io.LimitReader(part, 2*1024*1024))
				content = strings.TrimSpace(string(b))
			case "categories":
				b, _ := io.ReadAll(io.LimitReader(part, 64*1024))
				category := strings.TrimSpace(string(b))
				if category != "" {
					categories = append(categories, category)
				}
			case "image":
				if part.FileName() == "" {
					break
				}
				if part.Header.Get("Content-Length") != "" {
					if sz, _ := strconv.ParseInt(part.Header.Get("Content-Length"), 10, 64); sz > maxImage {
						RenderErrorPage(w, r, http.StatusRequestEntityTooLarge,
							"Image too big (max 20 MB)")
						return
					}
				}
				// Even if Content-Length missing we'll truncate copy at 20 MB
				if err := os.MkdirAll("uploads", 0755); err != nil {
					RenderErrorPage(w, r, http.StatusInternalServerError, "Cannot create uploads dir")
					return
				}
				fname := fmt.Sprintf("uploads/%d_%s", time.Now().UnixNano(), filepath.Base(part.FileName()))
				dst, err := os.Create(fname)
				if err != nil {
					RenderErrorPage(w, r, http.StatusInternalServerError, "Failed saving image")
					return
				}
				n, _ := io.Copy(dst, io.LimitReader(part, maxImage+1))
				dst.Close()
				if n > maxImage {
					os.Remove(fname) // delete partial
					RenderErrorPage(w, r, http.StatusRequestEntityTooLarge,
						"Image too big (max 20 MB)")
					return
				}
				imagePath = fname
			}
		}

		// Validate form data
		errors := make(map[string]string)
		formData := make(map[string]string)

		if title == "" {
			errors["title"] = "Title is required"
		} else {
			formData["title"] = title
		}

		if content == "" {
			errors["content"] = "Content is required"
		} else {
			formData["content"] = content
		}

		if len(categories) == 0 {
			errors["categories"] = "At least one category is required"
		} else {
			// Store selected categories for form preservation
			formData["categories"] = strings.Join(categories, ",")
		}

		// If there are validation errors, reload the form with errors
		if len(errors) > 0 {
			categoriesList, err := LoadCategoriesFromDB()
			if err != nil {
				RenderErrorPage(w, r, http.StatusInternalServerError, "Failed to load categories")
				return
			}

			data := models.PageData{
				Username:   user.Username,
				IsLoggedIn: user != nil,
				Categories: categoriesList,
				Errors:     errors,
				FormData:   formData,
			}
			RenderTemplate(w, "create-post", data)
			return
		}

		var imageArg interface{}
		if imagePath != "" {
			imageArg = imagePath
		} else {
			imageArg = nil
		}

		// Start a transaction
		tx, err := database.DB.Begin()
		if err != nil {
			RenderErrorPage(w, r, http.StatusInternalServerError, "Failed to start transaction")
			return
		}
		defer tx.Rollback()

		// Insert the post (use first category as the main category for backward compatibility)
		result, err := tx.Exec(
			`INSERT INTO posts (user_id, title, content, category, image_url) VALUES (?, ?, ?, ?, ?)`,
			user.ID, title, content, categories[0], imageArg,
		)
		if err != nil {
			RenderErrorPage(w, r, http.StatusInternalServerError, "Failed to save post")
			return
		}

		postID, err := result.LastInsertId()
		if err != nil {
			RenderErrorPage(w, r, http.StatusInternalServerError, "Failed to get post ID")
			return
		}

		// Insert category relationships
		for _, categoryName := range categories {
			// Get category ID by name
			var categoryID int
			err := tx.QueryRow("SELECT id FROM categories WHERE name = ?", categoryName).Scan(&categoryID)
			if err != nil {
				RenderErrorPage(w, r, http.StatusInternalServerError, "Invalid category: "+categoryName)
				return
			}

			// Insert into post_categories table
			_, err = tx.Exec("INSERT INTO post_categories (post_id, category_id) VALUES (?, ?)", postID, categoryID)
			if err != nil {
				RenderErrorPage(w, r, http.StatusInternalServerError, "Failed to save category relationship")
				return
			}
		}

		// Commit the transaction
		if err := tx.Commit(); err != nil {
			RenderErrorPage(w, r, http.StatusInternalServerError, "Failed to commit transaction")
			return
		}

		http.Redirect(w, r, "/forum", http.StatusSeeOther)
		return
	}

	RenderErrorPage(w, r, http.StatusMethodNotAllowed, "Invalid method")
}

// LoadPostsFromDB loads recent posts with author usernames and their categories.
func LoadPostsFromDB(page int, user *models.User) ([]models.Post, int, error) {
	var totalPosts int
	err := database.DB.QueryRow("SELECT COUNT(*) FROM posts").Scan(&totalPosts)
	if err != nil {
		return nil, 0, err
	}

	// Calculate pagination
	postsPerPage := 5
	offset := (page - 1) * postsPerPage

	// Get paginated posts with profile pictures
	rows, err := database.DB.Query(`
		SELECT posts.id, posts.title, posts.content, posts.category, posts.image_url, posts.created_at, users.username, users.profile_picture,
			(SELECT COUNT(*) FROM likes WHERE post_id = posts.id AND is_liked = 1) as like_count,
			(SELECT COUNT(*) FROM likes WHERE post_id = posts.id AND is_liked = 0) as dislike_count,
			(SELECT COUNT(*) FROM comments WHERE post_id = posts.id) as comment_count
		FROM posts
		JOIN users ON posts.user_id = users.id
		ORDER BY posts.created_at DESC
		LIMIT ? OFFSET ?
	`, postsPerPage, offset)

	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	var posts []models.Post
	for rows.Next() {
		var post models.Post
		var username string
		var imageURL sql.NullString
		var profilePicture sql.NullString
		var likeCount int
		var dislikeCount int
		var commentCount int

		if err := rows.Scan(&post.ID, &post.Title, &post.Content, &post.Category, &imageURL, &post.CreatedAt, &username, &profilePicture, &likeCount, &dislikeCount, &commentCount); err != nil {
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

		if user != nil {
			err := database.DB.QueryRow(
				"SELECT EXISTS (SELECT 1 FROM likes WHERE user_id = ? AND post_id = ? AND is_liked = 1)",
				user.ID, post.ID,
			).Scan(&post.UserLiked)
			if err != nil {
				return nil, 0, err
			}

			err = database.DB.QueryRow(
				"SELECT EXISTS (SELECT 1 FROM likes WHERE user_id = ? AND post_id = ? AND is_liked = 0)",
				user.ID, post.ID,
			).Scan(&post.UserDisliked)
			if err != nil {
				return nil, 0, err
			}
		}

		post.Likes = likeCount
		post.Dislikes = dislikeCount
		post.Comments = commentCount

		// Correctly assign the post's main image URL
		post.ImageURL = imageURL

		// Correctly assign the author's avatar URL
		if profilePicture.Valid && profilePicture.String != "" {
			post.AvatarURL = profilePicture.String
		} else {
			// Use a default avatar if the user doesn't have one
			post.AvatarURL = "" // Let the template handle the default
		}

		post.Author = username
		post.TimeAgo = formatTimeAgo(post.CreatedAt)
		posts = append(posts, post)
	}

	return posts, totalPosts, nil
}

func LoadPostsFromDBByCategory(category string, page int, user *models.User) ([]models.Post, int, error) {
	var totalPosts int
	// Count posts that belong to the selected category (via post_categories)
	err := database.DB.QueryRow(`
		SELECT COUNT(DISTINCT posts.id)
		FROM posts
		JOIN post_categories ON posts.id = post_categories.post_id
		JOIN categories ON post_categories.category_id = categories.id
		WHERE categories.name = ?
	`, category).Scan(&totalPosts)
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
		JOIN post_categories ON posts.id = post_categories.post_id
		JOIN categories ON post_categories.category_id = categories.id
		WHERE categories.name = ?
		GROUP BY posts.id
		ORDER BY posts.created_at DESC
		LIMIT ? OFFSET ?
	`, category, postsPerPage, offset)

	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	var posts []models.Post
	for rows.Next() {
		var post models.Post
		var username string
		var imageURL sql.NullString
		var profilePicture sql.NullString
		var likeCount, dislikeCount, commentCount int

		if err := rows.Scan(&post.ID, &post.Title, &post.Content, &post.Category, &imageURL, &post.CreatedAt, &username, &profilePicture, &likeCount, &dislikeCount, &commentCount); err != nil {
			return nil, 0, err
		}

		// Load all categories for this post
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

		if user != nil {
			// Check if user liked this post
			err := database.DB.QueryRow(
				"SELECT EXISTS (SELECT 1 FROM likes WHERE user_id = ? AND post_id = ? AND is_liked = 1)",
				user.ID, post.ID,
			).Scan(&post.UserLiked)
			if err != nil {
				return nil, 0, err
			}

			// Check if user disliked this post
			err = database.DB.QueryRow(
				"SELECT EXISTS (SELECT 1 FROM likes WHERE user_id = ? AND post_id = ? AND is_liked = 0)",
				user.ID, post.ID,
			).Scan(&post.UserDisliked)
			if err != nil {
				return nil, 0, err
			}
		}

		post.Likes = likeCount
		post.Dislikes = dislikeCount
		post.Comments = commentCount
		post.ImageURL = imageURL

		if profilePicture.Valid && profilePicture.String != "" {
			post.AvatarURL = profilePicture.String
		} else {
			post.AvatarURL = ""
		}

		post.Author = username
		post.TimeAgo = formatTimeAgo(post.CreatedAt)
		posts = append(posts, post)
	}

	return posts, totalPosts, nil
}

// HandlePostDetail serves the post detail page for /posts/{id}.
func HandlePostDetail(w http.ResponseWriter, r *http.Request) {
	idStr := strings.TrimPrefix(r.URL.Path, "/posts/")
	postID, err := strconv.Atoi(idStr)
	if err != nil {
		RenderErrorPage(w, r, http.StatusNotFound, "Post Not Found")
		return
	}

	user, _ := r.Context().Value("user").(*models.User)
	post, err := getPostByID(postID, user)
	if err == sql.ErrNoRows {
		RenderErrorPage(w, r, http.StatusNotFound, "Post Not Found")
		return
	} else if err != nil {
		RenderErrorPage(w, r, http.StatusInternalServerError, "Server error")
		return
	}

	post.TimeAgo = formatTimeAgo(post.CreatedAt)

	comments, err := LoadCommentsForPost(postID, user)
	if err != nil {
		http.Error(w, "Failed to load comments", http.StatusInternalServerError)
		return
	}

	// Format comment dates
	for i := range comments {
		comments[i].TimeAgo = formatTimeAgo(comments[i].CreatedAt)
	}

	// Handle nil user case
	username := ""
	isLoggedIn := false
	if user != nil {
		username = user.Username
		isLoggedIn = true
	}

	data := models.PageData{
		IsLoggedIn: isLoggedIn,
		Username:   username,
		Post:       post,
		Comments:   comments,
	}

	RenderTemplate(w, "post", data)
}

// getPostByID fetches a post by ID including the author's username and categories.
func getPostByID(postID int, user *models.User) (models.Post, error) {
	var post models.Post
	var username string
	var imageURL sql.NullString
	var likeCount int
	var dislikeCount int
	var commentCount int
	var profilePicture sql.NullString

	err := database.DB.QueryRow(`
		SELECT posts.id, posts.title, posts.content, posts.category, posts.image_url, posts.created_at, users.username, users.profile_picture,
			(SELECT COUNT(*) FROM likes WHERE post_id = posts.id AND is_liked = 1) as like_count,
			(SELECT COUNT(*) FROM likes WHERE post_id = posts.id AND is_liked = 0) as dislike_count,
			(SELECT COUNT(*) FROM comments WHERE post_id = posts.id) as comment_count
		FROM posts
		JOIN users ON posts.user_id = users.id
		WHERE posts.id = ?
	`, postID).Scan(&post.ID, &post.Title, &post.Content, &post.Category, &imageURL, &post.CreatedAt, &username, &profilePicture, &likeCount, &dislikeCount, &commentCount)
	if err != nil {
		return models.Post{}, err
	}

	// Load categories for this post
	categoryRows, err := database.DB.Query(`
		SELECT categories.name 
		FROM categories 
		JOIN post_categories ON categories.id = post_categories.category_id 
		WHERE post_categories.post_id = ?
		ORDER BY categories.name
	`, postID)
	if err != nil {
		return models.Post{}, err
	}

	var categories []string
	for categoryRows.Next() {
		var categoryName string
		if err := categoryRows.Scan(&categoryName); err != nil {
			categoryRows.Close()
			return models.Post{}, err
		}
		categories = append(categories, categoryName)
	}
	categoryRows.Close()

	// If no categories found in post_categories table, use the main category
	if len(categories) == 0 {
		categories = []string{post.Category}
	}
	post.Categories = categories

	if user != nil {
		err := database.DB.QueryRow(
			"SELECT EXISTS (SELECT 1 FROM likes WHERE user_id = ? AND post_id = ? AND is_liked = 1)",
			user.ID, post.ID,
		).Scan(&post.UserLiked)
		if err != nil {
			return post, err
		}

		err = database.DB.QueryRow(
			"SELECT EXISTS (SELECT 1 FROM likes WHERE user_id = ? AND post_id = ? AND is_liked = 0)",
			user.ID, post.ID,
		).Scan(&post.UserDisliked)
		if err != nil {
			return post, err
		}
	}

	post.Likes = likeCount
	post.Dislikes = dislikeCount
	post.Comments = commentCount

	// Correctly assign the post's main image URL
	post.ImageURL = imageURL

	// Correctly assign the author's avatar URL
	if profilePicture.Valid && profilePicture.String != "" {
		post.AvatarURL = profilePicture.String
	}

	post.Author = username
	return post, nil
}

// formatTimeAgo formats a duration in a human-readable way
func formatTimeAgo(t time.Time) string {
	duration := time.Since(t)

	// If less than a minute
	if duration < time.Minute {
		return "just now"
	}

	// If less than an hour
	if duration < time.Hour {
		minutes := int(duration.Minutes())
		if minutes == 1 {
			return "1 minute ago"
		}
		return fmt.Sprintf("%d minutes ago", minutes)
	}

	// If less than a day
	if duration < 24*time.Hour {
		hours := int(duration.Hours())
		if hours == 1 {
			return "1 hour ago"
		}
		return fmt.Sprintf("%d hours ago", hours)
	}

	// If less than a week
	if duration < 7*24*time.Hour {
		days := int(duration.Hours() / 24)
		if days == 1 {
			return "1 day ago"
		}
		return fmt.Sprintf("%d days ago", days)
	}

	// If less than a month (30 days)
	if duration < 30*24*time.Hour {
		weeks := int(duration.Hours() / (24 * 7))
		if weeks == 1 {
			return "1 week ago"
		}
		return fmt.Sprintf("%d weeks ago", weeks)
	}

	// If less than a year
	if duration < 365*24*time.Hour {
		months := int(duration.Hours() / (24 * 30))
		if months == 1 {
			return "1 month ago"
		}
		return fmt.Sprintf("%d months ago", months)
	}

	// More than a year
	years := int(duration.Hours() / (24 * 365))
	if years == 1 {
		return "1 year ago"
	}
	return fmt.Sprintf("%d years ago", years)
}
