package models

import "time"

type User struct {
	ID             int       `db:"id" json:"ID"`
	Username       string    `db:"username" json:"username"`
	Email          string    `db:"email" json:"email"`
	Password       string    `db:"password" json:"password"`
	ProfilePicture string    `db:"profile_picture" json:"profilePicture"`
	CreatedAt      time.Time `db:"created_at" json:"createdAt"`
	GithubID       int64     `db:"github_id" json:"githubId"` // Add this field
	GoogleID       string    `db:"google_id" json:"google_id"`
}
