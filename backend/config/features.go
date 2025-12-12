package config

import (
	"os"
	"strconv"
	"sync"
)

// FeatureFlags holds all feature flag states
type FeatureFlags struct {
	mu                 sync.RWMutex
	aiLessonsEnabled   bool
	aiProvider         string // "openai", "anthropic", "local"
	aiModel            string
	maxAILessonsPerDay int
}

var (
	instance *FeatureFlags
	once     sync.Once
)

// GetFeatureFlags returns the singleton feature flags instance
func GetFeatureFlags() *FeatureFlags {
	once.Do(func() {
		instance = &FeatureFlags{
			aiLessonsEnabled:   getEnvBool("ENABLE_AI_LESSONS", false),
			aiProvider:         getEnvString("AI_PROVIDER", "openai"),
			aiModel:            getEnvString("AI_MODEL", "gpt-4"),
			maxAILessonsPerDay: getEnvInt("MAX_AI_LESSONS_PER_DAY", 10),
		}
	})
	return instance
}

// IsAILessonsEnabled returns whether AI lesson generation is enabled
func (f *FeatureFlags) IsAILessonsEnabled() bool {
	f.mu.RLock()
	defer f.mu.RUnlock()
	return f.aiLessonsEnabled
}

// SetAILessonsEnabled sets the AI lessons feature flag (for testing/runtime toggle)
func (f *FeatureFlags) SetAILessonsEnabled(enabled bool) {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.aiLessonsEnabled = enabled
}

// GetAIProvider returns the configured AI provider
func (f *FeatureFlags) GetAIProvider() string {
	f.mu.RLock()
	defer f.mu.RUnlock()
	return f.aiProvider
}

// GetAIModel returns the configured AI model
func (f *FeatureFlags) GetAIModel() string {
	f.mu.RLock()
	defer f.mu.RUnlock()
	return f.aiModel
}

// GetMaxAILessonsPerDay returns the max AI lessons per day limit
func (f *FeatureFlags) GetMaxAILessonsPerDay() int {
	f.mu.RLock()
	defer f.mu.RUnlock()
	return f.maxAILessonsPerDay
}

// Helper functions
func getEnvBool(key string, defaultVal bool) bool {
	val := os.Getenv(key)
	if val == "" {
		return defaultVal
	}
	b, err := strconv.ParseBool(val)
	if err != nil {
		return defaultVal
	}
	return b
}

func getEnvString(key, defaultVal string) string {
	val := os.Getenv(key)
	if val == "" {
		return defaultVal
	}
	return val
}

func getEnvInt(key string, defaultVal int) int {
	val := os.Getenv(key)
	if val == "" {
		return defaultVal
	}
	i, err := strconv.Atoi(val)
	if err != nil {
		return defaultVal
	}
	return i
}
