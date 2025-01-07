package main

//dependencies
import (
    "encoding/json"
	"fmt"
	"net/http"
	"github.com/gorilla/mux"
	"math/rand"
	"time"
)

// This will store our URL mappings
var urlStore = make(map[string]string)

// Function to generate a random shortened URL
func generateShortURL() string {
	const charset = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	rand.Seed(time.Now().UnixNano())
	shortURL := make([]byte, 6)
	for i := range shortURL {
		shortURL[i] = charset[rand.Intn(len(charset))]
	}
	return string(shortURL)
}

// Function to handle URL shortening
func shortenURL(w http.ResponseWriter, r *http.Request) {
	// Parse the incoming URL from the body
	var request struct {
		OriginalURL string `json:"original_url"`
	}
	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		http.Error(w, "Invalid input", http.StatusBadRequest)
		return
	}

	// Generate a short URL
	shortURL := generateShortURL()

	// Store the mapping
	urlStore[shortURL] = request.OriginalURL

	// Return the shortened URL
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	fmt.Fprintf(w, `{"short_url": "http://localhost:8080/%s"}`, shortURL)
}

// Function to handle URL redirection
func redirectURL(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	shortURL := vars["shortURL"]

	originalURL, exists := urlStore[shortURL]
	if !exists {
		http.Error(w, "URL not found", http.StatusNotFound)
		return
	}

	http.Redirect(w, r, originalURL, http.StatusFound)
}

func main() {
	// Initialize the router
	r := mux.NewRouter()

	// Define the routes
	r.HandleFunc("/shorten", shortenURL).Methods("POST")
	r.HandleFunc("/{shortURL}", redirectURL).Methods("GET")

	// Start the server
	http.Handle("/", r)
	fmt.Println("Server started at :8080")
	http.ListenAndServe(":8080", nil)
}
