package main

import (
	"context"
	"log"
	"math/rand"
	"net/http"
	"os"
	"path/filepath"
	"time"

	"avidlearner/internal/models"
	"avidlearner/internal/routes"
	"avidlearner/lessons"
)

const (
	lessonFetchTTL          = 6 * time.Hour
	lessonMapRefreshDelay   = 15 * time.Second
	lessonMapRefreshEvery   = 10 * time.Minute
	leaderboardSaveInterval = 5 * time.Minute
)

func envOrDefault(key, fallback string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return fallback
}

func loadSecretLessons(path string) []models.Lesson {
	if _, err := os.Stat(path); err != nil {
		return nil
	}

	secretLessons, err := routes.LoadLessons(path)
	if err != nil {
		log.Printf("Warning: failed to load secret knowledge lessons: %v", err)
		return nil
	}

	log.Printf("Loaded %d lessons from Book of Secret Knowledge", len(secretLessons))
	return secretLessons
}

func buildFetcherLessons(coreLessons []models.Lesson, secretLessons []models.Lesson) []lessons.Lesson {
	localLessons := make([]lessons.Lesson, 0, len(coreLessons)+len(secretLessons))
	localLessons = appendFetcherLessons(localLessons, coreLessons, "local")
	return appendFetcherLessons(localLessons, secretLessons, "secret-knowledge")
}

func appendFetcherLessons(dst []lessons.Lesson, src []models.Lesson, source string) []lessons.Lesson {
	for _, lesson := range src {
		dst = append(dst, lessons.Lesson{
			Title:    lesson.Title,
			Category: lesson.Category,
			Text:     lesson.Text,
			Explain:  lesson.Explain,
			UseCases: lesson.UseCases,
			Tips:     lesson.Tips,
			Source:   source,
		})
	}
	return dst
}

// ---------- Main ----------

func main() {
	rand.Seed(time.Now().UnixNano())
	ctx := context.Background()

	// Load lessons
	dataPath := envOrDefault("LESSONS_FILE", filepath.Join("..", "data", "lessons.json"))
	loaded, err := routes.LoadLessons(dataPath)
	if err != nil {
		log.Fatalf("failed to load lessons from %s: %v", dataPath, err)
	}

	// Load secret knowledge lessons
	secretKnowledgePath := filepath.Join("..", "data", "secret_knowledge_lessons.json")
	secretLessons := loadSecretLessons(secretKnowledgePath)

	// Convert loaded lessons to lessons.Lesson type
	localLessons := buildFetcherLessons(loaded, secretLessons)

	// Initialize hybrid lesson fetcher (cache TTL: 6 hours)
	lessonFetcher := lessons.NewFetcher(localLessons, lessonFetchTTL)

	// Start background refresh (every 6 hours)
	lessonFetcher.StartBackgroundRefresh(ctx, lessonFetchTTL)

	// Get initial lessons (local + external)
	allLessons := lessonFetcher.GetLessons(ctx)
	routes.UpdateLessonMap(allLessons)

	log.Printf("Loaded %d lessons total (%d local + external)", len(allLessons), len(loaded))

	// Periodically refresh lesson map from fetcher (every 10 minutes)
	go func() {
		// Wait a bit for first external fetch, then refresh immediately
		time.Sleep(lessonMapRefreshDelay)
		allLessons := lessonFetcher.GetLessons(ctx)
		routes.UpdateLessonMap(allLessons)

		ticker := time.NewTicker(lessonMapRefreshEvery)
		defer ticker.Stop()
		for range ticker.C {
			allLessons := lessonFetcher.GetLessons(ctx)
			routes.UpdateLessonMap(allLessons)
		}
	}()

	// Load pro challenges
	proPath := envOrDefault("PRO_CHALLENGES_FILE", filepath.Join("..", "data", "pro_challenges.json"))
	if _, err := os.Stat(proPath); os.IsNotExist(err) {
		proPath = filepath.Join("data", "pro_challenges.json")
	}
	challenges, byID, pcErr := routes.LoadProChallenges(proPath)
	if pcErr != nil {
		log.Fatalf("failed to load pro challenges from %s: %v", proPath, pcErr)
	}
	routes.SetProChallenges(challenges, byID)

	// Load leaderboard from disk
	leaderboardPath := envOrDefault("LEADERBOARD_FILE", filepath.Join("..", "data", "leaderboard.json"))
	if _, err := os.Stat(filepath.Dir(leaderboardPath)); os.IsNotExist(err) {
		leaderboardPath = filepath.Join("data", "leaderboard.json")
	}
	if err := routes.LoadLeaderboard(leaderboardPath); err != nil {
		log.Printf("Warning: failed to load leaderboard from %s: %v (starting fresh)", leaderboardPath, err)
		routes.SetLeaderboard([]models.LeaderboardEntry{})
	}

	// Save leaderboard periodically
	go func() {
		ticker := time.NewTicker(leaderboardSaveInterval)
		defer ticker.Stop()
		for range ticker.C {
			if err := routes.SaveLeaderboard(leaderboardPath); err != nil {
				log.Printf("Error saving leaderboard: %v", err)
			}
		}
	}()

	// API
	routes.RegisterAPIHandler()

	port := os.Getenv("PORT")
	if port == "" {
		port = "8081"
	}
	log.Printf("Server listening on :%s", port)
	log.Fatal(http.ListenAndServe(":"+port, nil))
}
