package models

import (
	"database/sql"
	"strings"
	"time"
)

type Post struct {
	ID           int            `db:"id" json:"ID"`
	UserID       int            `db:"user_id" json:"userID"`
	Author       string         `db:"author" json:"author"`
	AvatarURL    string         `db:"avatar_url" json:"avatarURL"`
	Title        string         `db:"title" json:"title"`
	Content      string         `db:"content" json:"content"`
	ImageURL     sql.NullString `db:"image_url" json:"imageURL"`
	Category     string         `db:"category" json:"category"`
	Categories   []string       `db:"-" json:"categories"`
	CreatedAt    time.Time      `db:"created_at" json:"createdAt"`
	Comments     int            `db:"comments" json:"comments"`
	Likes        int            `db:"likes" json:"likes"`
	Dislikes     int            `db:"dislikes" json:"dislikes"`
	UserLiked    bool           `db:"userLiked" json:"-"`
	UserDisliked bool           `db:"userDisliked" json:"-"`
	TimeAgo      string         `db:"time_ago" json:"timeAgo"`
}

// GetCategoryNames returns a comma-separated string of category names
func (p *Post) GetCategoryNames() string {
	if len(p.Categories) > 0 {
		return strings.Join(p.Categories, ", ")
	}
	return p.Category // Fallback to old single category
}
