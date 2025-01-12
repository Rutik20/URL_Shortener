package main

//dependencies
import (
    "encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"sort"
	"strings"
	"github.com/gorilla/mux"
	"math/rand"
	"time"
)

// This will store our URL mappings
var urlStore = make(map[string]string)
var originalToShort = make(map[string]string) // Reverse mapping from original URL to short URL
var domainCounts = make(map[string]int)       // Count of shortened domains

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

// Function to extract the domain from a URL
func extractDomain(originalURL string) (string, error) {
	parsedURL, err := url.Parse(originalURL)
	if err != nil {
		return "", err
	}
	domain := strings.TrimPrefix(parsedURL.Host, "www.")
	return domain, nil
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

	// Check if the original URL already exists in the reverse mapping
	if shortURL, exists := originalToShort[request.OriginalURL]; exists {
		// Return the existing shortened URL
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		fmt.Fprintf(w, `{"short_url": "http://localhost:8080/%s"}`, shortURL)
		return
	}

	// Generate a new short URL
	shortURL := generateShortURL()

	// Extract the domain and update the count
	domain, err := extractDomain(request.OriginalURL)
	if err != nil {
		http.Error(w, "Invalid URL", http.StatusBadRequest)
		return
	}
	domainCounts[domain]++

	// Store the mapping
	urlStore[shortURL] = request.OriginalURL
	originalToShort[request.OriginalURL] = shortURL

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

// Function to return top 3 most shortened domains
func metrics(w http.ResponseWriter, r *http.Request) {
    // Add logging
    fmt.Println("Metrics endpoint called")
    fmt.Printf("Current domain counts: %+v\n", domainCounts)

    // Convert the domainCounts map to a sortable slice
    type domainStat struct {
        Domain string
        Count  int
    }
    var stats []domainStat
    for domain, count := range domainCounts {
        stats = append(stats, domainStat{Domain: domain, Count: count})
        fmt.Printf("Adding domain %s with count %d\n", domain, count)
    }

    // Sort the slice by count in descending order
    sort.Slice(stats, func(i, j int) bool {
        return stats[i].Count > stats[j].Count
    })

    // Take the top 3 domains
    if len(stats) > 3 {
        stats = stats[:3]
    }

    // Format the result
    result := make(map[string]int)
    for _, stat := range stats {
        result[stat.Domain] = stat.Count
    }

    // Log the final result
    fmt.Printf("Returning metrics result: %+v\n", result)

    // Return the result as JSON
    w.Header().Set("Content-Type", "application/json")
    w.WriteHeader(http.StatusOK)
    if err := json.NewEncoder(w).Encode(result); err != nil {
        fmt.Printf("Error encoding JSON response: %v\n", err)
        http.Error(w, "Error encoding response", http.StatusInternalServerError)
        return
    }
}

func main() {
	// Initialize the router
	r := mux.NewRouter()

	// Define the routes
	r.HandleFunc("/shorten", shortenURL).Methods("POST")
	r.HandleFunc("/metrics", metrics).Methods("GET")
	r.HandleFunc("/{shortURL}", redirectURL).Methods("GET")

	// Start the server
	http.Handle("/", r)
	fmt.Println("Server started at :8080")
	http.ListenAndServe(":8080", nil)
}
