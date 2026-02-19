package app

import (
	"context"
	"errors"
	"fmt"
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
	defaultLessonFetchTTL          = 6 * time.Hour
	defaultLessonMapRefreshDelay   = 15 * time.Second
	defaultLessonMapRefreshEvery   = 10 * time.Minute
	defaultLeaderboardSaveInterval = 5 * time.Minute
	defaultShutdownTimeout         = 10 * time.Second
)

type Config struct {
	LessonsFile           string
	SecretLessonsFile     string
	ProChallengesFile     string
	LeaderboardFile       string
	Port                  string
	LessonFetchTTL        time.Duration
	LessonMapRefreshDelay time.Duration
	LessonMapRefreshEvery time.Duration
	LeaderboardSaveEvery  time.Duration
}

func Run(ctx context.Context) error {
	rand.Seed(time.Now().UnixNano())
	cfg := loadConfig()
	return run(ctx, cfg)
}

func loadConfig() Config {
	cfg := Config{
		LessonsFile:           envOrDefault("LESSONS_FILE", filepath.Join("..", "data", "lessons.json")),
		SecretLessonsFile:     filepath.Join("..", "data", "secret_knowledge_lessons.json"),
		ProChallengesFile:     envOrDefault("PRO_CHALLENGES_FILE", filepath.Join("..", "data", "pro_challenges.json")),
		LeaderboardFile:       envOrDefault("LEADERBOARD_FILE", filepath.Join("..", "data", "leaderboard.json")),
		Port:                  envOrDefault("PORT", "8081"),
		LessonFetchTTL:        defaultLessonFetchTTL,
		LessonMapRefreshDelay: defaultLessonMapRefreshDelay,
		LessonMapRefreshEvery: defaultLessonMapRefreshEvery,
		LeaderboardSaveEvery:  defaultLeaderboardSaveInterval,
	}

	cfg.ProChallengesFile = resolveFileFallback(cfg.ProChallengesFile, filepath.Join("data", "pro_challenges.json"))
	cfg.LeaderboardFile = resolveDirFallback(cfg.LeaderboardFile, filepath.Join("data", "leaderboard.json"))

	return cfg
}

func run(ctx context.Context, cfg Config) error {
	loaded, err := routes.LoadLessons(cfg.LessonsFile)
	if err != nil {
		return fmt.Errorf("load lessons from %s: %w", cfg.LessonsFile, err)
	}

	secretLessons, err := loadSecretLessons(cfg.SecretLessonsFile)
	if err != nil {
		log.Printf("Warning: failed to load secret knowledge lessons: %v", err)
	}

	localLessons := buildFetcherLessons(loaded, secretLessons)
	lessonFetcher := lessons.NewFetcher(localLessons, cfg.LessonFetchTTL)
	lessonFetcher.StartBackgroundRefresh(ctx, cfg.LessonFetchTTL)

	allLessons := lessonFetcher.GetLessons(ctx)
	routes.UpdateLessonMap(allLessons)
	log.Printf("Loaded %d lessons total (%d local + external)", len(allLessons), len(loaded))

	startLessonMapRefresh(ctx, lessonFetcher, cfg.LessonMapRefreshDelay, cfg.LessonMapRefreshEvery)

	challenges, byID, err := routes.LoadProChallenges(cfg.ProChallengesFile)
	if err != nil {
		return fmt.Errorf("load pro challenges from %s: %w", cfg.ProChallengesFile, err)
	}
	routes.SetProChallenges(challenges, byID)

	if err := routes.LoadLeaderboard(cfg.LeaderboardFile); err != nil {
		log.Printf("Warning: failed to load leaderboard from %s: %v (starting fresh)", cfg.LeaderboardFile, err)
		routes.SetLeaderboard([]models.LeaderboardEntry{})
	}

	startLeaderboardSaver(ctx, cfg.LeaderboardFile, cfg.LeaderboardSaveEvery)

	routes.RegisterAPIHandler()

	return runServer(ctx, cfg.Port, defaultShutdownTimeout)
}

func runServer(ctx context.Context, port string, shutdownTimeout time.Duration) error {
	server := &http.Server{
		Addr: ":" + port,
	}

	errCh := make(chan error, 1)
	go func() {
		log.Printf("Server listening on :%s", port)
		errCh <- server.ListenAndServe()
	}()

	select {
	case err := <-errCh:
		if err != nil && !errors.Is(err, http.ErrServerClosed) {
			return err
		}
		return nil
	case <-ctx.Done():
		shutdownCtx, cancel := context.WithTimeout(context.Background(), shutdownTimeout)
		defer cancel()

		if err := server.Shutdown(shutdownCtx); err != nil {
			return fmt.Errorf("server shutdown: %w", err)
		}

		err := <-errCh
		if err != nil && !errors.Is(err, http.ErrServerClosed) {
			return err
		}
		return nil
	}
}

func startLessonMapRefresh(ctx context.Context, fetcher *lessons.Fetcher, delay, every time.Duration) {
	go func() {
		timer := time.NewTimer(delay)
		defer timer.Stop()

		select {
		case <-timer.C:
			allLessons := fetcher.GetLessons(ctx)
			routes.UpdateLessonMap(allLessons)
		case <-ctx.Done():
			return
		}

		ticker := time.NewTicker(every)
		defer ticker.Stop()
		for {
			select {
			case <-ticker.C:
				allLessons := fetcher.GetLessons(ctx)
				routes.UpdateLessonMap(allLessons)
			case <-ctx.Done():
				return
			}
		}
	}()
}

func startLeaderboardSaver(ctx context.Context, path string, every time.Duration) {
	go func() {
		ticker := time.NewTicker(every)
		defer ticker.Stop()
		for {
			select {
			case <-ticker.C:
				if err := routes.SaveLeaderboard(path); err != nil {
					log.Printf("Error saving leaderboard: %v", err)
				}
			case <-ctx.Done():
				return
			}
		}
	}()
}

func envOrDefault(key, fallback string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return fallback
}

func resolveFileFallback(path string, fallback string) string {
	if _, err := os.Stat(path); os.IsNotExist(err) {
		return fallback
	}
	return path
}

func resolveDirFallback(path string, fallback string) string {
	if _, err := os.Stat(filepath.Dir(path)); os.IsNotExist(err) {
		return fallback
	}
	return path
}

func loadSecretLessons(path string) ([]models.Lesson, error) {
	if _, err := os.Stat(path); err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}
		return nil, err
	}

	secretLessons, err := routes.LoadLessons(path)
	if err != nil {
		return nil, err
	}

	log.Printf("Loaded %d lessons from Book of Secret Knowledge", len(secretLessons))
	return secretLessons, nil
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
