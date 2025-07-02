package handlers

import (
	"io"
	"net/http"
	"os"
)

func HandleNasaAPI(w http.ResponseWriter, r *http.Request) {
	apiKey := os.Getenv("NASA_API_KEY")
	resp, err := http.Get("https://api.nasa.gov/planetary/apod?api_key=" + apiKey)
	if err != nil {
		http.Error(w, "Failed to fetch APOD", http.StatusInternalServerError)
		return
	}
	defer resp.Body.Close()
	w.Header().Set("Content-Type", "application/json")
	io.Copy(w, resp.Body)
}
