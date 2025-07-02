package models

type Likes struct {
	ID        int  `db:"id" json:"ID"`
	UserID    int  `db:"user_id" json:"userID"`
	PostID    *int `db:"post_id" json:"postID"`
	CommentID *int `db:"comment_id" json:"commentID"`
	IsLiked   bool `db:"is_liked" json:"isLiked"`
}
