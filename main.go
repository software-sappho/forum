package main

import (
	"forum/database"
	"forum/populate"
	"forum/routes"
	"log"
	"net/http"
)

func main() {
	// Initialize database
	if err := database.InitDB(); err != nil {
		log.Fatal(err)
	}

	if err := populate.Populate(); err != nil {
		log.Fatal("Error populating mock data:", err)
	}
	log.Println("Successfully populated mock data")

	// Setup routes
	routes.SetupRoutes()

	log.Println("Server starting on http://localhost:8080")
	log.Fatal(http.ListenAndServe(":8080", nil))
}
