package handlers

import (
	"forum/database"
	"forum/models"
	"log"
	"net/http"
	"strings"
)

func HandleCreateCategory(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		data := GetPageData(r)
		RenderTemplate(w, "create-category", data)

	case http.MethodPost:
		name := r.FormValue("name")
		description := r.FormValue("description")

		if strings.TrimSpace(name) == "" {
			RenderErrorPage(w, r, http.StatusBadRequest, "Category name is required")
			return
		}

		_, err := database.DB.Exec(`INSERT INTO categories (name, description) VALUES (?, ?)`, name, description)
		if err != nil {
			RenderErrorPage(w, r, http.StatusInternalServerError, "Failed to create category")
			log.Println("Create category error:", err)
			return
		}

		http.Redirect(w, r, "/forum", http.StatusSeeOther)

	default:
		RenderErrorPage(w, r, http.StatusMethodNotAllowed, "Method not allowed")
	}
}

func LoadCategoriesFromDB() ([]models.Category, error) {
	rows, err := database.DB.Query(`SELECT id, name, description FROM categories ORDER BY name ASC`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var categories []models.Category
	for rows.Next() {
		var c models.Category
		if err := rows.Scan(&c.ID, &c.Name, &c.Description); err != nil {
			return nil, err
		}
		categories = append(categories, c)
	}
	return categories, nil
}
