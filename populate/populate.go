package populate

import (
	"encoding/json"
	"fmt"
	"forum/database"
	"forum/models"
	"os"
)

type MockData struct {
	Users      []models.User     `json:"users"`
	Categories []models.Category `json:"categories"`
	Posts      []models.Post     `json:"posts"`
}

func Populate() error {
	data, err := os.ReadFile("populate/seed.json")
	if err != nil {
		return fmt.Errorf("error reading mock data file: %v", err)
	}

	var mockData MockData
	if err := json.Unmarshal(data, &mockData); err != nil {
		return fmt.Errorf("error unmarshaling data: %v", err)
	}

	userStmt, err := database.DB.Prepare(`
		INSERT  OR IGNORE INTO users (id, username, email, password, created_at)
		VALUES (?, ?, ?, ?, ?)
	`)
	if err != nil {
		return fmt.Errorf("error preparing user statement: %v", err)
	}
	defer userStmt.Close()

	for _, user := range mockData.Users {
		_, err := userStmt.Exec(
			user.ID,
			user.Username,
			user.Email,
			user.Password,
			user.CreatedAt,
		)
		if err != nil {
			return fmt.Errorf("error inserting user %d: %v", user.ID, err)
		}
	}

	categoryStmt, err := database.DB.Prepare(`
		INSERT OR IGNORE INTO categories (id, name, description)
		VALUES (?, ?, ?)
	`)
	if err != nil {
		return fmt.Errorf("error preparing category statement: %v", err)
	}
	defer categoryStmt.Close()

	for _, category := range mockData.Categories {
		_, err := categoryStmt.Exec(
			category.ID,
			category.Name,
			category.Description,
		)
		if err != nil {
			return fmt.Errorf("error inserting category %d: %v", category.ID, err)
		}
	}

	postStmt, err := database.DB.Prepare(`
		INSERT OR IGNORE INTO posts (
			ID, user_id, title, content, category,
			created_at
		) VALUES (?, ?, ?, ?, ?, ?)
	`)
	if err != nil {
		return fmt.Errorf("error preparing post statement: %v", err)
	}
	defer postStmt.Close()

	for _, post := range mockData.Posts {
		_, err = postStmt.Exec(
			post.ID,
			post.UserID,
			post.Title,
			post.Content,
			post.Category,
			post.CreatedAt,
		)
		if err != nil {
			return fmt.Errorf("error inserting post %d: %v", post.ID, err)
		}
	}

	return nil
}

func getCategoryID(name string, categories []models.Category) (int, error) {
	for _, c := range categories {
		if c.Name == name {
			return c.ID, nil
		}
	}
	return 0, fmt.Errorf("category not found: %s", name)
}
