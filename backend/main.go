package main

import (
	"encoding/json"
	"log"
	"math/rand"
	"net/http"
	"os"
	"sort"
	"time"
)

// Lesson model
type Lesson struct {
	Title    string   `json:"title"`
	Category string   `json:"category"`
	Text     string   `json:"text"`
	Explain  string   `json:"explain"`
	UseCases []string `json:"useCases"`
	Tips     []string `json:"tips"`
}

type LessonsResponse struct {
	Categories []string            `json:"categories"`
	Lessons    map[string][]Lesson `json:"lessons"`
}

// seedLessons: moved to data/lessons.json

var lessonsByCat map[string][]Lesson
var categories []string

func main() {
	rand.Seed(time.Now().UnixNano())

	// Load lessons from JSON file
	lessonsFile := os.Getenv("LESSONS_FILE")
	if lessonsFile == "" {
		lessonsFile = "../data/lessons.json"
	}
	loaded, err := loadLessons(lessonsFile)
	if err != nil {
		log.Fatalf("failed to load lessons from %s: %v", lessonsFile, err)
	}
	lessonsByCat = map[string][]Lesson{}
	for _, l := range loaded {
		lessonsByCat[l.Category] = append(lessonsByCat[l.Category], l)
	}
	for cat := range lessonsByCat {
		categories = append(categories, cat)
	}
	sort.Strings(categories)

	fs := http.FileServer(http.Dir("../frontend"))
	http.Handle("/", fs)

	// Wrap API handlers with CORS middleware. Allowed origin can be set
	// with the ALLOWED_ORIGIN environment variable (defaults to http://localhost:8081).
	http.HandleFunc("/api/lessons", corsMiddleware(handleLessons))
	http.HandleFunc("/api/random", corsMiddleware(handleRandom))

	port := os.Getenv("PORT")
	if port == "" {
		port = "80"
	}
	log.Printf("Backend running on http://localhost:%s …\n", port)
	if err := http.ListenAndServe(":"+port, nil); err != nil {
		log.Fatal(err)
	}
}

func loadLessons(path string) ([]Lesson, error) {
	f, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	var lessons []Lesson
	if err := json.Unmarshal(f, &lessons); err != nil {
		return nil, err
	}
	return lessons, nil
}

func handleLessons(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	resp := LessonsResponse{Categories: categories, Lessons: lessonsByCat}
	_ = json.NewEncoder(w).Encode(resp)
}

func handleRandom(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	cat := r.URL.Query().Get("category")
	var pool []Lesson
	if cat == "" || cat == "any" {
		for _, ls := range lessonsByCat {
			pool = append(pool, ls...)
		}
	} else {
		pool = lessonsByCat[cat]
	}
	if len(pool) == 0 {
		w.WriteHeader(http.StatusNotFound)
		_ = json.NewEncoder(w).Encode(map[string]string{"error": "no lessons for category"})
		return
	}
	l := pool[rand.Intn(len(pool))]
	_ = json.NewEncoder(w).Encode(l)
}

// corsMiddleware wraps an http.HandlerFunc and sets CORS headers.
// Allowed origin can be specified with ALLOWED_ORIGIN env var. If empty,
// the middleware will allow the request origin or default to http://localhost:8081.
func corsMiddleware(next http.HandlerFunc) http.HandlerFunc {
	allowed := os.Getenv("ALLOWED_ORIGIN")
	if allowed == "" {
		allowed = "http://localhost:80"
	}
	return func(w http.ResponseWriter, r *http.Request) {
		origin := r.Header.Get("Origin")
		if origin != "" {
			// If ALLOWED_ORIGIN is a wildcard, allow all. Otherwise allow only matching origin.
			if allowed == "*" || origin == allowed {
				w.Header().Set("Access-Control-Allow-Origin", origin)
				w.Header().Set("Vary", "Origin")
			}
		} else {
			// If request has no Origin header (e.g., same-origin), still set default
			if allowed != "" {
				w.Header().Set("Access-Control-Allow-Origin", allowed)
			}
		}
		w.Header().Set("Access-Control-Allow-Methods", "GET, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type")

		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusNoContent)
			return
		}
		next(w, r)
	}
}
