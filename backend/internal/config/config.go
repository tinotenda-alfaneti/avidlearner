package config

import (
	"os"
	"path/filepath"
	"time"
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
	ShutdownTimeout       time.Duration
}

func Load() Config {
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
		ShutdownTimeout:       defaultShutdownTimeout,
	}

	cfg.ProChallengesFile = resolveFileFallback(cfg.ProChallengesFile, filepath.Join("data", "pro_challenges.json"))
	cfg.LeaderboardFile = resolveDirFallback(cfg.LeaderboardFile, filepath.Join("data", "leaderboard.json"))

	return cfg
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
