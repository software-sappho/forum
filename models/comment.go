package models

import "time"

type Comment struct {
	ID           int       `db:"id" json:"ID"`
	PostID       int       `db:"post_id" json:"postID"`
	PostTitle    string    `db:"post_title" json:"postTitle"`
	UserID       int       `db:"user_id" json:"userID"`
	Content      string    `db:"content" json:"content"`
	Author       string    `db:"author" json:"author"`
	AvatarURL    string    `db:"avatar_url" json:"avatarURL"`
	Likes        int       `db:"likes" json:"likes"`
	Dislikes     int       `db:"dislikes" json:"dislikes"`
	UserLiked    bool      `db:"userLiked" json:"-"`
	UserDisliked bool      `db:"userDisliked" json:"-"`
	CreatedAt    time.Time `db:"created_at" json:"createdAt"`
	TimeAgo      string    `db:"-" json:"timeAgo"`
}
