package main

import (
	"context"

	"log"
	"math/rand"
	"net/http"
	"os"
	"path/filepath"
	"sort"
	"sync"
	"time"

	. "avidlearner/internal/models"
	"avidlearner/lessons"
)


const lessonRepeatWindow = 100

func newProfile() *Profile {
	return &Profile{
		HintIdx: map[string]int{},
	}
}

// ---------- Globals ----------
var (
	lessonsByCat      map[string][]Lesson
	categories        []string
	sessions          = map[string]*Profile{} // sid -> profile
	proChallenges     []ProChallenge
	proChallengesByID map[string]ProChallenge
	leaderboard       []LeaderboardEntry // in-memory leaderboard
	lessonFetcher     *lessons.Fetcher

	newsCache   = map[string]NewsCacheEntry{}
	newsCacheMu sync.RWMutex
	newsTTL     = 10 * time.Minute
	tldrNewsTTL = 30 * time.Minute
)

// ---------- Main ----------

func main() {
	rand.Seed(time.Now().UnixNano())

	// Load lessons
	dataPath := os.Getenv("LESSONS_FILE")
	if dataPath == "" {
		dataPath = filepath.Join("..", "data", "lessons.json")
	}
	loaded, err := loadLessons(dataPath)
	if err != nil {
		log.Fatalf("failed to load lessons from %s: %v", dataPath, err)
	}
	coreCount := len(loaded)

	// Load secret knowledge lessons
	var secretLessons []Lesson
	secretKnowledgePath := filepath.Join("..", "data", "secret_knowledge_lessons.json")
	if _, err := os.Stat(secretKnowledgePath); err == nil {
		secretLessons, err = loadLessons(secretKnowledgePath)
		if err != nil {
			log.Printf("Warning: failed to load secret knowledge lessons: %v", err)
		} else {
			log.Printf("Loaded %d lessons from Book of Secret Knowledge", len(secretLessons))
		}
	}

	// Convert loaded lessons to lessons.Lesson type
	localLessons := make([]lessons.Lesson, 0, len(loaded)+len(secretLessons))

	// Add core lessons with "local" source
	for _, l := range loaded[:coreCount] {
		localLessons = append(localLessons, lessons.Lesson{
			Title:    l.Title,
			Category: l.Category,
			Text:     l.Text,
			Explain:  l.Explain,
			UseCases: l.UseCases,
			Tips:     l.Tips,
			Source:   "local",
		})
	}

	// Add secret knowledge lessons with "secret-knowledge" source
	for _, l := range secretLessons {
		localLessons = append(localLessons, lessons.Lesson{
			Title:    l.Title,
			Category: l.Category,
			Text:     l.Text,
			Explain:  l.Explain,
			UseCases: l.UseCases,
			Tips:     l.Tips,
			Source:   "secret-knowledge",
		})
	}

	// Initialize hybrid lesson fetcher (cache TTL: 6 hours)
	lessonFetcher = lessons.NewFetcher(localLessons, 6*time.Hour)

	// Start background refresh (every 6 hours)
	lessonFetcher.StartBackgroundRefresh(context.Background(), 6*time.Hour)

	// Get initial lessons (local + external)
	allLessons := lessonFetcher.GetLessons(context.Background())

	// Build category map from all lessons
	lessonsByCat = map[string][]Lesson{}
	for _, l := range allLessons {
		mainLesson := Lesson{
			Title:    l.Title,
			Category: l.Category,
			Text:     l.Text,
			Explain:  l.Explain,
			UseCases: l.UseCases,
			Tips:     l.Tips,
			Source:   l.Source,
		}
		lessonsByCat[l.Category] = append(lessonsByCat[l.Category], mainLesson)
	}
	for cat := range lessonsByCat {
		categories = append(categories, cat)
	}
	sort.Strings(categories)

	log.Printf("Loaded %d lessons total (%d local + external)", len(allLessons), len(loaded))

	// Periodically refresh lesson map from fetcher (every 10 minutes)
	go func() {
		// Wait a bit for first external fetch, then refresh immediately
		time.Sleep(15 * time.Second)
		allLessons := lessonFetcher.GetLessons(context.Background())
		updateLessonMap(allLessons)

		ticker := time.NewTicker(10 * time.Minute)
		defer ticker.Stop()
		for range ticker.C {
			allLessons := lessonFetcher.GetLessons(context.Background())
			updateLessonMap(allLessons)
		}
	}()

	// Load pro challenges
	proPath := os.Getenv("PRO_CHALLENGES_FILE")
	if proPath == "" {
		proPath = filepath.Join("..", "data", "pro_challenges.json")
		if _, err := os.Stat(proPath); os.IsNotExist(err) {
			proPath = filepath.Join("data", "pro_challenges.json")
		}
	}
	var pcErr error
	proChallenges, proChallengesByID, pcErr = loadProChallenges(proPath)
	if pcErr != nil {
		log.Fatalf("failed to load pro challenges from %s: %v", proPath, pcErr)
	}

	// Load leaderboard from disk
	leaderboardPath := os.Getenv("LEADERBOARD_FILE")
	if leaderboardPath == "" {
		leaderboardPath = filepath.Join("..", "data", "leaderboard.json")
		if _, err := os.Stat(filepath.Dir(leaderboardPath)); os.IsNotExist(err) {
			leaderboardPath = filepath.Join("data", "leaderboard.json")
		}
	}
	if err := loadLeaderboard(leaderboardPath); err != nil {
		log.Printf("Warning: failed to load leaderboard from %s: %v (starting fresh)", leaderboardPath, err)
		leaderboard = []LeaderboardEntry{}
	}

	// Save leaderboard periodically
	go func() {
		ticker := time.NewTicker(5 * time.Minute)
		defer ticker.Stop()
		for range ticker.C {
			if err := saveLeaderboard(leaderboardPath); err != nil {
				log.Printf("Error saving leaderboard: %v", err)
			}
		}
	}()

	// API
	registerAPIHandler()

	port := os.Getenv("PORT")
	if port == "" {
		port = "8081"
	}
	log.Printf("Server listening on :%s", port)
	log.Fatal(http.ListenAndServe(":"+port, nil))
}