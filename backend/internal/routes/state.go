package routes

import (
	"sync"
	"time"

	"avidlearner/internal/models"
	"avidlearner/internal/lessons"
)

const lessonRepeatWindow = 100

// ---------- Globals ----------
var (
	lessonsByCat      map[string][]models.Lesson
	categories        []string
	sessions          = map[string]*models.Profile{} // sid -> profile
	proChallenges     []models.ProChallenge
	proChallengesByID map[string]models.ProChallenge
	leaderboard       []models.LeaderboardEntry // in-memory leaderboard

	newsCache   = map[string]models.NewsCacheEntry{}
	newsCacheMu sync.RWMutex
	newsTTL     = 10 * time.Minute
	tldrNewsTTL = 30 * time.Minute
)

func newProfile() *models.Profile {
	return &models.Profile{
		HintIdx: map[string]int{},
	}
}

// ---------- Exports for main ----------
func RegisterAPIHandler() {
	registerAPIHandler()
}

func UpdateLessonMap(allLessons []lessons.Lesson) {
	updateLessonMap(allLessons)
}

func LoadLessons(path string) ([]models.Lesson, error) {
	return loadLessons(path)
}

func LoadProChallenges(path string) ([]models.ProChallenge, map[string]models.ProChallenge, error) {
	return loadProChallenges(path)
}

func LoadLeaderboard(path string) error {
	return loadLeaderboard(path)
}

func SaveLeaderboard(path string) error {
	return saveLeaderboard(path)
}

func SetProChallenges(list []models.ProChallenge, byID map[string]models.ProChallenge) {
	proChallenges = list
	proChallengesByID = byID
}

func SetLeaderboard(entries []models.LeaderboardEntry) {
	leaderboard = entries
}
