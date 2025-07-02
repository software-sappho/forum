package filters

import (
	"forum/models"
	"net/http"
	"strings"
)

func Filters(posts []models.Post, totalPosts int, r *http.Request) ([]models.Post, int) {
	category := strings.TrimSpace(r.URL.Query().Get("category"))
	if category == "" {
		return posts, totalPosts
	}

	countPosts := 0
	var filtered []models.Post
	for _, post := range posts {
		// Check if the post has multiple categories
		if len(post.Categories) > 0 {
			// Check if any of the post's categories match the selected category
			for _, postCategory := range post.Categories {
				if strings.EqualFold(postCategory, category) {
					filtered = append(filtered, post)
					countPosts++
					break // Found a match, no need to check other categories
				}
			}
		} else {
			// Fallback to the old single category field
			if strings.EqualFold(post.Category, category) {
				filtered = append(filtered, post)
				countPosts++
			}
		}
	}
	return filtered, countPosts
}
