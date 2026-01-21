package main

import (
	"bytes"
	"context"
	crand "crypto/rand"
	"encoding/hex"
	"encoding/json"
	"encoding/xml"
	"errors"
	"fmt"
	"log"
	mrand "math/rand"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"sort"
	"strings"
	"time"
	"sync"

	"avidlearner/ai"
	"avidlearner/config"
	"avidlearner/lessons"
)

// ---------- Models ----------

type Lesson struct {
	Title    string   `json:"title"`
	Category string   `json:"category"`
	Text     string   `json:"text"`
	Explain  string   `json:"explain"`
	UseCases []string `json:"useCases"`
	Tips     []string `json:"tips"`
	Source   string   `json:"source,omitempty"` // "local", "github", "devto"
}

type LessonsResponse struct {
	Categories []string            `json:"categories"`
	Lessons    map[string][]Lesson `json:"lessons"`
}

const lessonRepeatWindow = 100

type SessionState struct {
	Stage string `json:"stage"`
	// lesson/reading
	Lesson *Lesson `json:"lesson,omitempty"`
	// quiz
	Question string   `json:"question,omitempty"`
	Options  []string `json:"options,omitempty"`
	Index    int      `json:"index,omitempty"`
	Total    int      `json:"total,omitempty"`
	// result/answer
	Correct     bool   `json:"correct,omitempty"`
	CoinsEarned int    `json:"coinsEarned,omitempty"`
	CoinsTotal  int    `json:"coinsTotal,omitempty"`
	XPTotal     int    `json:"xpTotal,omitempty"`
	More        bool   `json:"more,omitempty"`
	Message     string `json:"message,omitempty"`
}

// One generated MCQ
type QuizQuestion struct {
	LessonTitle  string
	Question     string
	Options      []string
	CorrectIndex int
}

type ChallengeStarter struct {
	Filename string `json:"filename"`
	Code     string `json:"code"`
}

type ChallengeReward struct {
	XP    int `json:"xp"`
	Coins int `json:"coins"`
}

type ProChallenge struct {
	ID          string           `json:"id"`
	Title       string           `json:"title"`
	Difficulty  string           `json:"difficulty"`
	Topics      []string         `json:"topics"`
	Description string           `json:"description"`
	Starter     ChallengeStarter `json:"starter"`
	Hints       []string         `json:"hints"`
	Reward      ChallengeReward  `json:"reward"`
}

type testFailure struct {
	Name   string `json:"name"`
	Output string `json:"output"`
}

type challengeTestResult struct {
	Passed   bool
	Total    int
	Failures []testFailure
	Stdout   string
	Stderr   string
}

// Per-session state
type profile struct {
	Coins       int
	Streak      int
	XP          int
	LessonsSeen []string

	CurrentQuiz   []QuizQuestion
	QuizIndex     int
	LastLesson    *Lesson
	RecentLessons []string
	HintIdx       map[string]int // challengeID -> next hint index
	PlayerName    string         // for leaderboard

	QuizScore       int       // Current quiz session score
	TypingScore     int       // Best typing score this session
	CodingScore     int       // Coding challenges score
	LastScoreSubmit time.Time // Prevent spam submissions
}

// Leaderboard entry
type LeaderboardEntry struct {
	Name     string    `json:"name"`
	Score    int       `json:"score"`
	Mode     string    `json:"mode"` // "quiz", "typing", "coding"
	Date     time.Time `json:"date"`
	Category string    `json:"category,omitempty"`
}

func newProfile() *profile {
	return &profile{
		HintIdx: map[string]int{},
	}
}

// ---------- Globals ----------
var (
	lessonsByCat      map[string][]Lesson
	categories        []string
	sessions          = map[string]*profile{} // sid -> profile
	proChallenges     []ProChallenge
	proChallengesByID map[string]ProChallenge
	leaderboard       []LeaderboardEntry // in-memory leaderboard
	lessonFetcher     *lessons.Fetcher
)

// Simple in-memory cache for news feeds
type newsCacheEntry struct {
	ts   time.Time
	data []byte
}

var (
	newsCache   = map[string]newsCacheEntry{}
	newsCacheMu sync.RWMutex
	newsTTL     = 10 * time.Minute
)

// ---------- Main ----------

func main() {
	mrand.Seed(time.Now().UnixNano())

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

	// Static frontend (vite build output)
	frontendDist := filepath.Join("frontend", "dist")
	if _, err := os.Stat(frontendDist); os.IsNotExist(err) {
		frontendDist = "/app/frontend/dist"
	}
	http.Handle("/", withSession(http.FileServer(http.Dir(frontendDist))))

	// Health
	http.HandleFunc("/healthz", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		fmt.Fprint(w, "ok")
	})

	// API
	http.HandleFunc("/api/lessons", cors(handleLessons))
	http.HandleFunc("/api/random", cors(handleRandom))
	http.HandleFunc("/api/session", cors(handleSession)) // multi-stage + POSTs
	http.HandleFunc("/api/ai/generate", cors(handleAIGenerate))
	http.HandleFunc("/api/ai/config", cors(handleAIConfig))
	http.HandleFunc("/api/prochallenge", cors(handleProChallenge))
	http.HandleFunc("/api/prochallenge/submit", cors(handleProChallengeSubmit))
	http.HandleFunc("/api/prochallenge/hint", cors(handleProChallengeHint))
	http.HandleFunc("/api/leaderboard", cors(handleLeaderboard))
	http.HandleFunc("/api/leaderboard/submit", cors(handleLeaderboardSubmit))
	http.HandleFunc("/api/typing/score", cors(handleTypingScore))

	port := os.Getenv("PORT")
	if port == "" {
		port = "8081"
	}
	log.Printf("Server listening on :%s", port)
	log.Fatal(http.ListenAndServe(":"+port, nil))
}

// ---------- Helpers ----------

func updateLessonMap(allLessons []lessons.Lesson) {
	newLessonsByCat := map[string][]Lesson{}
	newCategories := []string{}

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
		newLessonsByCat[l.Category] = append(newLessonsByCat[l.Category], mainLesson)
	}
	for cat := range newLessonsByCat {
		newCategories = append(newCategories, cat)
	}
	sort.Strings(newCategories)

	// Atomic update of globals
	lessonsByCat = newLessonsByCat
	categories = newCategories
	log.Printf("Refreshed lesson map: %d lessons total", len(allLessons))
}

func loadLessons(path string) ([]Lesson, error) {
	b, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	var L []Lesson
	if err := json.Unmarshal(b, &L); err != nil {
		return nil, err
	}
	return L, nil
}

func loadProChallenges(path string) ([]ProChallenge, map[string]ProChallenge, error) {
	b, err := os.ReadFile(path)
	if err != nil {
		return nil, nil, err
	}
	var list []ProChallenge
	if err := json.Unmarshal(b, &list); err != nil {
		return nil, nil, err
	}
	byID := make(map[string]ProChallenge, len(list))
	for _, ch := range list {
		if ch.ID == "" {
			continue
		}
		byID[ch.ID] = ch
	}
	return list, byID, nil
}

func loadLeaderboard(path string) error {
	b, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return nil // Fresh start is OK
		}
		return err
	}
	var entries []LeaderboardEntry
	if err := json.Unmarshal(b, &entries); err != nil {
		return err
	}
	leaderboard = entries
	return nil
}

func saveLeaderboard(path string) error {
	// Ensure directory exists
	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return err
	}

	b, err := json.MarshalIndent(leaderboard, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(path, b, 0644)
}

func pickRandomLesson(cat string) *Lesson {
	var pool []Lesson
	if cat == "" || strings.EqualFold(cat, "any") {
		for _, ls := range lessonsByCat {
			pool = append(pool, ls...)
		}
	} else {
		pool = lessonsByCat[cat]
	}
	if len(pool) == 0 {
		return nil
	}
	l := pool[mrand.Intn(len(pool))]
	return &l
}

func allLessons() []Lesson {
	var pool []Lesson
	for _, ls := range lessonsByCat {
		pool = append(pool, ls...)
	}
	return pool
}

func findLessonByTitle(title string) *Lesson {
	for _, ls := range lessonsByCat {
		for _, l := range ls {
			if l.Title == title {
				ll := l
				return &ll
			}
		}
	}
	return nil
}

func uniqueStrings(ss []string) []string {
	seen := map[string]struct{}{}
	var out []string
	for _, s := range ss {
		if _, ok := seen[s]; ok {
			continue
		}
		seen[s] = struct{}{}
		out = append(out, s)
	}
	return out
}

func pickLessonForProfile(p *profile, cat string, source string) *Lesson {
	if p == nil {
		return pickRandomLesson(cat)
	}

	var pool []Lesson
	if cat == "" || strings.EqualFold(cat, "any") {
		for _, ls := range lessonsByCat {
			pool = append(pool, ls...)
		}
	} else {
		pool = append(pool, lessonsByCat[cat]...)
	}
	log.Printf("pickLessonForProfile: after category filter, pool=%d lessons", len(pool))

	// Filter by source if specified
	if source != "" && source != "all" {
		var filtered []Lesson
		for _, l := range pool {
			if l.Source == source {
				filtered = append(filtered, l)
			}
		}
		log.Printf("pickLessonForProfile: filtering by source=%s, filtered=%d lessons", source, len(filtered))
		pool = filtered
	}

	if len(pool) == 0 {
		log.Printf("pickLessonForProfile: pool is empty after filtering!")
		return nil
	}

	maxAvoid := lessonRepeatWindow
	if len(pool) <= 1 {
		maxAvoid = 0
	} else if maxAvoid > len(pool)-1 {
		maxAvoid = len(pool) - 1
	}

	avoid := map[string]struct{}{}
	if maxAvoid > 0 {
		for i := len(p.RecentLessons) - 1; i >= 0 && len(avoid) < maxAvoid; i-- {
			title := p.RecentLessons[i]
			if title == "" {
				continue
			}
			if _, exists := avoid[title]; exists {
				continue
			}
			avoid[title] = struct{}{}
		}
	}

	selection := pool
	if len(avoid) > 0 {
		var candidates []Lesson
		for _, l := range pool {
			if _, ok := avoid[l.Title]; ok {
				continue
			}
			candidates = append(candidates, l)
		}
		if len(candidates) > 0 {
			selection = candidates
		}
	}

	chosen := selection[mrand.Intn(len(selection))]
	p.RecentLessons = append(p.RecentLessons, chosen.Title)
	if lessonRepeatWindow > 0 && len(p.RecentLessons) > lessonRepeatWindow*2 {
		p.RecentLessons = p.RecentLessons[len(p.RecentLessons)-lessonRepeatWindow:]
	}
	return &chosen
}

// ---------- Middleware ----------

func withSession(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		c, err := r.Cookie("sid")
		if err != nil || c.Value == "" {
			buf := make([]byte, 16)
			_, _ = crand.Read(buf)
			sid := hex.EncodeToString(buf)
			http.SetCookie(w, &http.Cookie{
				Name:     "sid",
				Value:    sid,
				Path:     "/",
				HttpOnly: true,
				MaxAge:   60 * 60 * 24 * 30,
				SameSite: http.SameSiteLaxMode,
			})
			sessions[sid] = newProfile()
			r.AddCookie(&http.Cookie{Name: "sid", Value: sid})
		}
		next.ServeHTTP(w, r)
	})
}

func getProfile(r *http.Request) *profile {
	c, err := r.Cookie("sid")
	if err != nil {
		return newProfile()
	}
	p, ok := sessions[c.Value]
	if !ok {
		p = newProfile()
		sessions[c.Value] = p
	}
	return p
}

// ---------- Handlers ----------

func handleLessons(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	_ = json.NewEncoder(w).Encode(LessonsResponse{
		Categories: categories,
		Lessons:    lessonsByCat,
	})
}

func handleRandom(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	p := getProfile(r)
	cat := r.URL.Query().Get("category")
	lesson := pickLessonForProfile(p, cat, "")
	if lesson == nil {
		w.WriteHeader(http.StatusNotFound)
		_ = json.NewEncoder(w).Encode(map[string]string{"error": "no lessons for category"})
		return
	}
	_ = json.NewEncoder(w).Encode(lesson)
}

// /api/session supports:
// GET  stage=lesson           -> returns a random lesson for reading
// POST stage=add              -> body: {"title":"..."} adds lesson to LessonsSeen
// POST stage=startQuiz        -> builds quiz from LessonsSeen (or all if empty) and returns first question
// GET  stage=quiz             -> returns current question (index/total)
// POST stage=answer           -> body: {"answerIndex":0..3} evals; returns result + maybe next question (More=true)
func handleSession(w http.ResponseWriter, r *http.Request) {
	p := getProfile(r)
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	stage := r.URL.Query().Get("stage")
	if stage == "" {
		stage = "lesson"
	}

	switch r.Method {
	case http.MethodGet:
		switch stage {
		case "lesson":
			cat := r.URL.Query().Get("category")
			source := r.URL.Query().Get("source")
			log.Printf("handleSession: category=%s, source=%s, total lessons=%d", cat, source, len(lessonsByCat))
			lesson := pickLessonForProfile(p, cat, source)
			if lesson == nil {
				log.Printf("pickLessonForProfile returned nil for cat=%s, source=%s", cat, source)
				http.Error(w, "no lessons", http.StatusInternalServerError)
				return
			}
			p.LastLesson = lesson
			_ = json.NewEncoder(w).Encode(SessionState{
				Stage:      "lesson",
				Lesson:     lesson,
				CoinsTotal: p.Coins,
				XPTotal:    p.XP,
			})
			return

		case "quiz":
			if len(p.CurrentQuiz) == 0 {
				http.Error(w, "no active quiz", http.StatusBadRequest)
				return
			}
			q := p.CurrentQuiz[p.QuizIndex]
			_ = json.NewEncoder(w).Encode(SessionState{
				Stage:      "quiz",
				Question:   q.Question,
				Options:    q.Options,
				Index:      p.QuizIndex + 1,
				Total:      len(p.CurrentQuiz),
				CoinsTotal: p.Coins,
				XPTotal:    p.XP,
			})
			return
		}

	case http.MethodPost:
		switch stage {
		case "add":
			var body struct {
				Title string `json:"title"`
			}
			if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
				http.Error(w, "bad body", http.StatusBadRequest)
				return
			}
			if body.Title != "" {
				p.LessonsSeen = uniqueStrings(append(p.LessonsSeen, body.Title))
			}
			_ = json.NewEncoder(w).Encode(map[string]any{
				"stage":       "added",
				"lessonsSeen": p.LessonsSeen,
				"count":       len(p.LessonsSeen),
				"message":     "lesson added to study list",
			})
			return

		case "startQuiz":
			var pool []Lesson
			if len(p.LessonsSeen) == 0 {
				pool = allLessons()
			} else {
				for _, t := range p.LessonsSeen {
					if l := findLessonByTitle(t); l != nil {
						pool = append(pool, *l)
					}
				}
			}
			if len(pool) == 0 {
				http.Error(w, "no lessons to quiz", http.StatusBadRequest)
				return
			}
			// build quiz
			p.CurrentQuiz = nil
			for _, l := range pool {
				qq := buildQuizForLesson(l)
				p.CurrentQuiz = append(p.CurrentQuiz, qq)
			}
			// shuffle questions
			mrand.Shuffle(len(p.CurrentQuiz), func(i, j int) { p.CurrentQuiz[i], p.CurrentQuiz[j] = p.CurrentQuiz[j], p.CurrentQuiz[i] })
			p.QuizIndex = 0
			p.QuizScore = 0 // Reset score for new quiz
			first := p.CurrentQuiz[0]
			_ = json.NewEncoder(w).Encode(SessionState{
				Stage:      "quiz",
				Question:   first.Question,
				Options:    first.Options,
				Index:      1,
				Total:      len(p.CurrentQuiz),
				Message:    "quiz started",
				CoinsTotal: p.Coins,
				XPTotal:    p.XP,
			})
			return

		case "answer":
			if len(p.CurrentQuiz) == 0 {
				http.Error(w, "no active quiz", http.StatusBadRequest)
				return
			}
			var body struct {
				AnswerIndex int `json:"answerIndex"`
			}
			if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
				http.Error(w, "bad body", http.StatusBadRequest)
				return
			}
			cur := p.CurrentQuiz[p.QuizIndex]
			correct := body.AnswerIndex == cur.CorrectIndex
			earned := 0
			if correct {
				earned = 10
				p.Coins += earned
				p.Streak += 1
				p.QuizScore++ // Track correct answers server-side
			} else {
				p.Streak = 0
			}
			// advance
			p.QuizIndex++
			more := p.QuizIndex < len(p.CurrentQuiz)

			resp := SessionState{
				Stage:       "result",
				Correct:     correct,
				CoinsEarned: earned,
				CoinsTotal:  p.Coins,
				XPTotal:     p.XP,
				More:        more,
				Message:     map[bool]string{true: "Correct! +10 coins", false: "Not quite. Keep going!"}[correct],
			}
			// include next question if more
			if more {
				next := p.CurrentQuiz[p.QuizIndex]
				resp.Stage = "quiz"
				resp.Question = next.Question
				resp.Options = next.Options
				resp.Index = p.QuizIndex + 1
				resp.Total = len(p.CurrentQuiz)
			} else {
				// end of quiz; clear selection list but keep progress coins/streak
				p.CurrentQuiz = nil
				p.LessonsSeen = nil
				p.QuizIndex = 0
			}
			_ = json.NewEncoder(w).Encode(resp)
			return
		}
	}

	http.Error(w, "invalid request", http.StatusBadRequest)
}

func handleProChallenge(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}
	if len(proChallenges) == 0 {
		http.Error(w, "no challenges available", http.StatusServiceUnavailable)
		return
	}

	difficulty := strings.TrimSpace(strings.ToLower(r.URL.Query().Get("difficulty")))
	if difficulty == "" {
		difficulty = "advanced"
	}
	topic := strings.TrimSpace(strings.ToLower(r.URL.Query().Get("topic")))

	var pool []ProChallenge
	for _, ch := range proChallenges {
		if difficulty != "" && difficulty != "any" && !strings.EqualFold(ch.Difficulty, difficulty) {
			continue
		}
		if topic != "" && topic != "any" {
			match := false
			for _, tpc := range ch.Topics {
				if strings.EqualFold(tpc, topic) {
					match = true
					break
				}
			}
			if !match {
				continue
			}
		}
		pool = append(pool, ch)
	}

	if len(pool) == 0 {
		http.Error(w, "no challenge found for selection", http.StatusNotFound)
		return
	}

	selected := pool[mrand.Intn(len(pool))]
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	_ = json.NewEncoder(w).Encode(selected)
}

func handleProChallengeSubmit(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}
	p := getProfile(r)
	if p.HintIdx == nil {
		p.HintIdx = map[string]int{}
	}

	var body struct {
		ID   string `json:"id"`
		Code string `json:"code"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		http.Error(w, "invalid body", http.StatusBadRequest)
		return
	}
	ch, ok := proChallengesByID[body.ID]
	if !ok {
		http.Error(w, "challenge not found", http.StatusNotFound)
		return
	}

	res, err := runChallengeTests(r.Context(), ch, body.Code)
	if err != nil {
		http.Error(w, fmt.Sprintf("test execution failed: %v", err), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	if res.Passed {
		p.Coins += ch.Reward.Coins
		p.XP += ch.Reward.XP
		p.CodingScore += ch.Reward.XP // Track coding score for leaderboard
		resp := map[string]any{
			"passed":      true,
			"total":       res.Total,
			"coinsEarned": ch.Reward.Coins,
			"coinsTotal":  p.Coins,
			"xpEarned":    ch.Reward.XP,
			"xpTotal":     p.XP,
			"message":     fmt.Sprintf("All tests passed! +%d coins · +%d XP", ch.Reward.Coins, ch.Reward.XP),
			"stdout":      res.Stdout,
		}
		_ = json.NewEncoder(w).Encode(resp)
		return
	}

	resp := map[string]any{
		"passed":   false,
		"total":    res.Total,
		"failures": res.Failures,
		"stdout":   res.Stdout,
		"stderr":   res.Stderr,
	}
	_ = json.NewEncoder(w).Encode(resp)
}

func handleProChallengeHint(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}
	var body struct {
		ID string `json:"id"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil || body.ID == "" {
		http.Error(w, "invalid body", http.StatusBadRequest)
		return
	}
	ch, ok := proChallengesByID[body.ID]
	if !ok {
		http.Error(w, "challenge not found", http.StatusNotFound)
		return
	}

	p := getProfile(r)
	if p.HintIdx == nil {
		p.HintIdx = map[string]int{}
	}
	if p.Coins >= 2 {
		p.Coins -= 2
	} else {
		p.Coins = 0
	}

	index := p.HintIdx[ch.ID]
	var hint string
	hasMore := false
	if index < len(ch.Hints) {
		hint = ch.Hints[index]
		index++
		if index < len(ch.Hints) {
			hasMore = true
		}
	} else if len(ch.Hints) > 0 {
		// no more hints; repeat last hint
		hint = ch.Hints[len(ch.Hints)-1]
		index = len(ch.Hints)
	}
	p.HintIdx[ch.ID] = index

	resp := map[string]any{
		"hint":       hint,
		"index":      index,
		"hasMore":    hasMore,
		"coinsTotal": p.Coins,
		"xpTotal":    p.XP,
	}
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	_ = json.NewEncoder(w).Encode(resp)
}

// Build one MCQ for a lesson (correct = lesson explain/text; distractors from others)
func buildQuizForLesson(l Lesson) QuizQuestion {
	question := fmt.Sprintf("Which statement best matches the concept '%s'?", l.Title)
	correct := strings.TrimSpace(l.Explain)
	if correct == "" {
		correct = strings.TrimSpace(l.Text)
	}
	var pool []string
	for _, ls := range lessonsByCat {
		for _, x := range ls {
			if x.Title == l.Title {
				continue
			}
			cur := strings.TrimSpace(x.Explain)
			if cur == "" {
				cur = strings.TrimSpace(x.Text)
			}
			if cur != "" {
				pool = append(pool, cur)
			}
		}
	}
	mrand.Shuffle(len(pool), func(i, j int) { pool[i], pool[j] = pool[j], pool[i] })
	opts := []string{correct}
	for i := 0; i < 3 && i < len(pool); i++ {
		opts = append(opts, pool[i])
	}
	for len(opts) < 4 {
		opts = append(opts, "This option does not apply to the concept.")
	}
	mrand.Shuffle(len(opts), func(i, j int) { opts[i], opts[j] = opts[j], opts[i] })
	correctIdx := 0
	for i, o := range opts {
		if o == correct {
			correctIdx = i
			break
		}
	}
	return QuizQuestion{
		LessonTitle:  l.Title,
		Question:     question,
		Options:      opts,
		CorrectIndex: correctIdx,
	}
}

func runChallengeTests(parent context.Context, ch ProChallenge, source string) (challengeTestResult, error) {
	var result challengeTestResult
	if strings.Contains(ch.ID, "..") {
		return result, fmt.Errorf("invalid challenge id")
	}
	if strings.TrimSpace(source) == "" {
		result.Failures = []testFailure{{Name: "submission", Output: "no code submitted"}}
		return result, nil
	}

	tempDir, err := os.MkdirTemp("", "avid-pro-*")
	if err != nil {
		return result, err
	}
	defer os.RemoveAll(tempDir)

	mod := "module example.com/protmp\n\ngo 1.24\n"
	if err := os.WriteFile(filepath.Join(tempDir, "go.mod"), []byte(mod), 0o644); err != nil {
		return result, err
	}
	if err := os.WriteFile(filepath.Join(tempDir, "challenge.go"), []byte(source), 0o644); err != nil {
		return result, err
	}

	testSrc, err := resolveChallengeTestPath(ch.ID)
	if err != nil {
		return result, err
	}
	if err := writeChallengeTest(filepath.Join(tempDir, "challenge_test.go"), testSrc); err != nil {
		return result, err
	}

	runCtx, cancel := context.WithTimeout(parent, 5*time.Second)
	defer cancel()

	cmd := exec.CommandContext(runCtx, "go", "test", "-run", "Test", "-count=1", "-timeout=3s", "./...")
	cmd.Dir = tempDir
	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	runErr := cmd.Run()
	result.Stdout = strings.TrimSpace(stdout.String())
	result.Stderr = strings.TrimSpace(stderr.String())
	result.Total = countTests(result.Stdout + "\n" + result.Stderr)

	if runErr != nil {
		result.Failures = extractFailures(result.Stdout + "\n" + result.Stderr)
		if len(result.Failures) == 0 {
			if errors.Is(runErr, context.DeadlineExceeded) || errors.Is(runCtx.Err(), context.DeadlineExceeded) {
				result.Failures = []testFailure{{Name: "timeout", Output: "tests exceeded execution time limit"}}
			} else if result.Stderr != "" {
				result.Failures = []testFailure{{Name: "tests", Output: result.Stderr}}
			} else if result.Stdout != "" {
				result.Failures = []testFailure{{Name: "tests", Output: result.Stdout}}
			} else {
				result.Failures = []testFailure{{Name: "tests", Output: runErr.Error()}}
			}
		}
		return result, nil
	}

	result.Passed = true
	return result, nil
}

func resolveChallengeTestPath(id string) (string, error) {
	candidates := []string{
		filepath.Join("protests", id, "challenge_test.go"),
		filepath.Join("backend", "protests", id, "challenge_test.go"),
		filepath.Join("..", "backend", "protests", id, "challenge_test.go"),
	}
	for _, path := range candidates {
		if info, err := os.Stat(path); err == nil && !info.IsDir() {
			return path, nil
		}
	}
	return "", fmt.Errorf("hidden tests for %s not found", id)
}

func writeChallengeTest(dst, src string) error {
	data, err := os.ReadFile(src)
	if err != nil {
		return err
	}
	lines := strings.Split(string(data), "\n")
	if len(lines) > 0 && strings.HasPrefix(strings.TrimSpace(lines[0]), "//go:build") {
		lines = lines[1:]
		if len(lines) > 0 && strings.TrimSpace(lines[0]) == "" {
			lines = lines[1:]
		}
	}
	cleaned := strings.Join(lines, "\n")
	return os.WriteFile(dst, []byte(cleaned), 0o644)
}

func countTests(out string) int {
	count := strings.Count(out, "--- PASS:")
	count += strings.Count(out, "--- FAIL:")
	count += strings.Count(out, "--- SKIP:")
	if count == 0 && strings.Contains(out, "PASS\n") {
		// fallback when test output suppressed
		count = 1
	}
	return count
}

func extractFailures(out string) []testFailure {
	lines := strings.Split(out, "\n")
	var (
		current *testFailure
		result  []testFailure
	)
	flush := func() {
		if current != nil {
			result = append(result, testFailure{
				Name:   current.Name,
				Output: strings.TrimSpace(current.Output),
			})
			current = nil
		}
	}

	for _, raw := range lines {
		line := strings.TrimSpace(raw)
		if line == "" {
			continue
		}
		if strings.HasPrefix(line, "--- FAIL:") {
			flush()
			fields := strings.Fields(line)
			name := ""
			if len(fields) >= 3 {
				name = fields[2]
			}
			current = &testFailure{Name: name, Output: line}
			continue
		}
		if strings.HasPrefix(line, "--- PASS:") || strings.HasPrefix(line, "--- SKIP:") || strings.HasPrefix(line, "PASS") || strings.HasPrefix(line, "FAIL") || strings.HasPrefix(line, "ok ") {
			flush()
			continue
		}
		if current != nil {
			current.Output += "\n" + line
		}
	}
	flush()
	return result
}

// Liberal CORS so frontend dev server can call POST endpoints
func cors(next http.HandlerFunc) http.HandlerFunc {
	allowed := os.Getenv("ALLOWED_ORIGIN")
	if allowed == "" {
		allowed = "*" // dev-friendly
	}
	return func(w http.ResponseWriter, r *http.Request) {
		origin := r.Header.Get("Origin")
		if allowed == "*" && origin != "" {
			w.Header().Set("Access-Control-Allow-Origin", origin)
		} else if origin == allowed {
			w.Header().Set("Access-Control-Allow-Origin", origin)
		} else if allowed != "*" {
			w.Header().Set("Access-Control-Allow-Origin", allowed)
		}
		w.Header().Set("Vary", "Origin")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type")
		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusNoContent)
			return
		}
		withSession(next).ServeHTTP(w, r)
	}
}

// ---------- AI Lesson Generation ----------

// handleAIGenerate generates a lesson using AI based on topic and category
func handleAIGenerate(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")

	// Check if AI lessons are enabled
	flags := config.GetFeatureFlags()
	if !flags.IsAILessonsEnabled() {
		http.Error(w, `{"error":"AI lesson generation is currently disabled"}`, http.StatusServiceUnavailable)
		return
	}

	if r.Method != http.MethodPost {
		http.Error(w, `{"error":"method not allowed"}`, http.StatusMethodNotAllowed)
		return
	}

	var req struct {
		Category string `json:"category"`
		Topic    string `json:"topic"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, `{"error":"invalid request body"}`, http.StatusBadRequest)
		return
	}

	if req.Topic == "" {
		http.Error(w, `{"error":"topic is required"}`, http.StatusBadRequest)
		return
	}

	if req.Category == "" {
		req.Category = "general"
	}

	// Get the configured AI provider
	provider, err := ai.GetProvider(flags.GetAIProvider(), flags.GetAIModel())
	if err != nil {
		log.Printf("Error getting AI provider: %v", err)
		http.Error(w, `{"error":"AI provider not available"}`, http.StatusInternalServerError)
		return
	}

	// Generate lesson with timeout
	ctx, cancel := context.WithTimeout(r.Context(), 30*time.Second)
	defer cancel()

	lesson, err := provider.GenerateLesson(ctx, req.Category, req.Topic)
	if err != nil {
		log.Printf("Error generating lesson: %v", err)
		http.Error(w, fmt.Sprintf(`{"error":"failed to generate lesson: %v"}`, err), http.StatusInternalServerError)
		return
	}

	// Convert AI lesson to main Lesson type
	mainLesson := &Lesson{
		Title:    lesson.Title,
		Category: lesson.Category,
		Text:     lesson.Text,
		Explain:  lesson.Explain,
		UseCases: lesson.UseCases,
		Tips:     lesson.Tips,
		Source:   "ai",
	}

	// Return the generated lesson
	response := SessionState{
		Stage:   "lesson",
		Lesson:  mainLesson,
		Message: "AI-generated lesson",
	}

	_ = json.NewEncoder(w).Encode(response)
}

// handleAIConfig returns the current AI feature flag status
func handleAIConfig(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")

	flags := config.GetFeatureFlags()

	response := map[string]interface{}{
		"aiEnabled": flags.IsAILessonsEnabled(),
		"provider":  flags.GetAIProvider(),
		"maxPerDay": flags.GetMaxAILessonsPerDay(),
	}

	_ = json.NewEncoder(w).Encode(response)
}

// ---------- Leaderboard ----------

// handleLeaderboard returns the leaderboard, optionally filtered by mode
func handleLeaderboard(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")

	if r.Method != http.MethodGet {
		http.Error(w, `{"error":"method not allowed"}`, http.StatusMethodNotAllowed)
		return
	}

	mode := r.URL.Query().Get("mode")
	limit := 100 // default limit

	filtered := []LeaderboardEntry{}
	for _, entry := range leaderboard {
		if mode == "" || entry.Mode == mode {
			filtered = append(filtered, entry)
		}
	}

	// Sort by score descending
	sort.Slice(filtered, func(i, j int) bool {
		return filtered[i].Score > filtered[j].Score
	})

	// Apply limit
	if len(filtered) > limit {
		filtered = filtered[:limit]
	}

	_ = json.NewEncoder(w).Encode(filtered)
}

// handleLeaderboardSubmit submits a score to the leaderboard
func handleLeaderboardSubmit(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")

	if r.Method != http.MethodPost {
		http.Error(w, `{"error":"method not allowed"}`, http.StatusMethodNotAllowed)
		return
	}

	p := getProfile(r)

	var req struct {
		Name     string `json:"name"`
		Score    int    `json:"score"`
		Mode     string `json:"mode"`
		Category string `json:"category"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, `{"error":"invalid request body"}`, http.StatusBadRequest)
		return
	}

	if req.Name == "" {
		req.Name = "Anonymous"
	}
	if req.Mode == "" {
		http.Error(w, `{"error":"mode is required"}`, http.StatusBadRequest)
		return
	}
	if req.Score < 0 {
		http.Error(w, `{"error":"invalid score"}`, http.StatusBadRequest)
		return
	}

	// SERVER-SIDE VALIDATION: Check if score is legitimate
	var validatedScore int
	switch req.Mode {
	case "quiz":
		validatedScore = p.QuizScore
		if req.Score > validatedScore {
			http.Error(w, `{"error":"invalid score: server validation failed"}`, http.StatusForbidden)
			return
		}
	case "typing":
		validatedScore = p.TypingScore
		if req.Score > validatedScore {
			http.Error(w, `{"error":"invalid score: server validation failed"}`, http.StatusForbidden)
			return
		}
	case "coding":
		validatedScore = p.CodingScore
		if req.Score > validatedScore {
			http.Error(w, `{"error":"invalid score: server validation failed"}`, http.StatusForbidden)
			return
		}
	default:
		http.Error(w, `{"error":"invalid mode"}`, http.StatusBadRequest)
		return
	}

	// Prevent spam submissions (1 minute cooldown)
	if time.Since(p.LastScoreSubmit) < time.Minute {
		http.Error(w, `{"error":"please wait before submitting another score"}`, http.StatusTooManyRequests)
		return
	}
	p.LastScoreSubmit = time.Now()

	// Use validated score from server, not client-submitted score
	req.Score = validatedScore

	// Sanitize name
	if len(req.Name) > 30 {
		req.Name = req.Name[:30]
	}

	entry := LeaderboardEntry{
		Name:     req.Name,
		Score:    req.Score,
		Mode:     req.Mode,
		Category: req.Category,
		Date:     time.Now(),
	}

	leaderboard = append(leaderboard, entry)

	// Keep only top 1000 entries to prevent memory issues
	if len(leaderboard) > 1000 {
		sort.Slice(leaderboard, func(i, j int) bool {
			return leaderboard[i].Score > leaderboard[j].Score
		})
		leaderboard = leaderboard[:1000]
	}

	// Save to disk immediately
	leaderboardPath := os.Getenv("LEADERBOARD_FILE")
	if leaderboardPath == "" {
		leaderboardPath = filepath.Join("..", "data", "leaderboard.json")
		if _, err := os.Stat(filepath.Dir(leaderboardPath)); os.IsNotExist(err) {
			leaderboardPath = filepath.Join("data", "leaderboard.json")
		}
	}
	if err := saveLeaderboard(leaderboardPath); err != nil {
		log.Printf("Error saving leaderboard: %v", err)
	}

	response := map[string]interface{}{
		"success": true,
		"rank":    calculateRank(entry),
		"message": "Score submitted successfully!",
	}

	_ = json.NewEncoder(w).Encode(response)
}

// calculateRank determines the player's rank on the leaderboard
func calculateRank(entry LeaderboardEntry) int {
	rank := 1
	for _, e := range leaderboard {
		if e.Mode == entry.Mode && e.Score > entry.Score {
			rank++
		}
	}
	return rank
}

// handleTypingScore updates the typing score for the session
func handleTypingScore(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")

	if r.Method != http.MethodPost {
		http.Error(w, `{"error":"method not allowed"}`, http.StatusMethodNotAllowed)
		return
	}

	p := getProfile(r)

	var req struct {
		Score int `json:"score"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, `{"error":"invalid request body"}`, http.StatusBadRequest)
		return
	}

	if req.Score < 0 {
		http.Error(w, `{"error":"invalid score"}`, http.StatusBadRequest)
		return
	}

	// Update typing score (keep best)
	if req.Score > p.TypingScore {
		p.TypingScore = req.Score
	}

	response := map[string]interface{}{
		"success": true,
		"score":   p.TypingScore,
	}

	_ = json.NewEncoder(w).Encode(response)
}
