package routes

import (
	"net/http"
	"sync"
	"time"

	. "avidlearner/internal/models"
	"avidlearner/lessons"
)

const lessonRepeatWindow = 100

const LessonRepeatWindow = lessonRepeatWindow

// ---------- Globals ----------
var (
	lessonsByCat      map[string][]Lesson
	categories        []string
	sessions          = map[string]*Profile{} // sid -> profile
	proChallenges     []ProChallenge
	proChallengesByID map[string]ProChallenge
	leaderboard       []LeaderboardEntry // in-memory leaderboard

	newsCache   = map[string]NewsCacheEntry{}
	newsCacheMu sync.RWMutex
	newsTTL     = 10 * time.Minute
	tldrNewsTTL = 30 * time.Minute
)

func newProfile() *Profile {
	return &Profile{
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

func LoadLessons(path string) ([]Lesson, error) {
	return loadLessons(path)
}

func LoadProChallenges(path string) ([]ProChallenge, map[string]ProChallenge, error) {
	return loadProChallenges(path)
}

func LoadLeaderboard(path string) error {
	return loadLeaderboard(path)
}

func SaveLeaderboard(path string) error {
	return saveLeaderboard(path)
}

func SetProChallenges(list []ProChallenge, byID map[string]ProChallenge) {
	proChallenges = list
	proChallengesByID = byID
}

func SetLeaderboard(entries []LeaderboardEntry) {
	leaderboard = entries
}

// ---------- Exports for tests ----------
func NewProfile() *Profile {
	return newProfile()
}

func WithSession(next http.Handler) http.Handler {
	return withSession(next)
}

func GetProfile(r *http.Request) *Profile {
	return getProfile(r)
}

func HandleLessons(w http.ResponseWriter, r *http.Request) {
	handleLessons(w, r)
}

func HandleRandom(w http.ResponseWriter, r *http.Request) {
	handleRandom(w, r)
}

func PickRandomLesson(cat string) *Lesson {
	return pickRandomLesson(cat)
}

func AllLessons() []Lesson {
	return allLessons()
}

func FindLessonByTitle(title string) *Lesson {
	return findLessonByTitle(title)
}

func UniqueStrings(ss []string) []string {
	return uniqueStrings(ss)
}

func PickLessonForProfile(p *Profile, cat string, source string) *Lesson {
	return pickLessonForProfile(p, cat, source)
}

func SetLessonsByCategory(data map[string][]Lesson) {
	lessonsByCat = data
}

func SetCategories(data []string) {
	categories = data
}

func LessonsByCategory() map[string][]Lesson {
	return lessonsByCat
}

func Categories() []string {
	return categories
}

func ResetSessions() {
	sessions = map[string]*Profile{}
}

func SetSession(id string, profile *Profile) {
	sessions[id] = profile
}

func Sessions() map[string]*Profile {
	return sessions
}
