package handlers

import (
	"io"
	"log"
	"net/http"
	"net/url"
	"os"
	"time"
)

const (
	nasaAPODURL = "https://api.nasa.gov/planetary/apod"
)

var (
	nasaAPIKey string
)

func init() {
	nasaAPIKey = os.Getenv("NASA_API_KEY")
	if nasaAPIKey == "" {
		log.Fatal("NASA_API_KEY environment variable not set")
	}
}

func ApodHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Methods", "GET")

	client := &http.Client{Timeout: 10 * time.Second}

	params := url.Values{}
	params.Add("api_key", nasaAPIKey)

	for key, values := range r.URL.Query() {
		if key != "api_key" {
			for _, value := range values {
				params.Add(key, value)
			}
		}
	}

	resp, err := client.Get(nasaAPODURL + "?" + params.Encode())
	if err != nil {
		http.Error(w, "Failed to fetch from NASA API", http.StatusInternalServerError)
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		http.Error(w, "NASA returned non-OK status", resp.StatusCode)
		return
	}

	w.Header().Set("Content-Type", resp.Header.Get("Content-Type"))
	w.WriteHeader(resp.StatusCode)

	bodyBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		http.Error(w, "Failed to read response", http.StatusInternalServerError)
		return
	}
	log.Println("Response body:", string(bodyBytes))
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(resp.StatusCode)
	w.Write(bodyBytes)

	if _, err := io.Copy(w, resp.Body); err != nil {
		log.Printf("Error streaming response: %v", err)
	}
}
