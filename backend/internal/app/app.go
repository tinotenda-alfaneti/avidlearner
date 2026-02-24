package app

import (
	"context"
	"errors"
	"fmt"
	"log"
	"math/rand"
	"net/http"
	"os"
	"time"

	"avidlearner/internal/config"
	"avidlearner/internal/models"
	"avidlearner/internal/routes"
	"avidlearner/internal/lessons"
)

func Run(ctx context.Context) error {
	rand.Seed(time.Now().UnixNano())
	cfg := config.Load()
	return run(ctx, cfg)
}

func run(ctx context.Context, cfg config.Config) error {
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

	if err := routes.LoadUsers(cfg.UsersFile); err != nil {
		log.Printf("Warning: failed to load users from %s: %v (starting fresh)", cfg.UsersFile, err)
	}

	if err := routes.SetAuthConfig(cfg.AuthSecret, cfg.AuthTokenTTL); err != nil {
		return fmt.Errorf("auth config: %w", err)
	}

	startUsersSaver(ctx, cfg.UsersFile, cfg.UsersSaveEvery)

	routes.RegisterAPIHandler()

	return runServer(ctx, cfg.Port, cfg.ShutdownTimeout)
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

func startUsersSaver(ctx context.Context, path string, every time.Duration) {
	go func() {
		ticker := time.NewTicker(every)
		defer ticker.Stop()
		for {
			select {
			case <-ticker.C:
				if err := routes.SaveUsers(path); err != nil {
					log.Printf("Error saving users: %v", err)
				}
			case <-ctx.Done():
				return
			}
		}
	}()
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
