package main

import (
	crand "crypto/rand"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"log"
	mrand "math/rand"
	"net/http"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"
)

// ---------- Models ----------

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

const lessonRepeatWindow = 100

type SessionState struct {
	Stage       string   `json:"stage"`
	// lesson/reading
	Lesson      *Lesson  `json:"lesson,omitempty"`
	// quiz
	Question    string   `json:"question,omitempty"`
	Options     []string `json:"options,omitempty"`
	Index       int      `json:"index,omitempty"`
	Total       int      `json:"total,omitempty"`
	// result/answer
	Correct     bool     `json:"correct,omitempty"`
	CoinsEarned int      `json:"coinsEarned,omitempty"`
	CoinsTotal  int      `json:"coinsTotal,omitempty"`
	More        bool     `json:"more,omitempty"`
	Message     string   `json:"message,omitempty"`
}

// One generated MCQ
type QuizQuestion struct {
	LessonTitle  string
	Question     string
	Options      []string
	CorrectIndex int
}

// Per-session state
type profile struct {
	Coins         int
	Streak        int
	LessonsSeen   []string

	CurrentQuiz   []QuizQuestion
	QuizIndex     int
	LastLesson    *Lesson
	RecentLessons []string
}

// ---------- Globals ----------
var (
	lessonsByCat map[string][]Lesson
	categories   []string
	sessions     = map[string]*profile{} // sid -> profile
)

// ---------- Main ----------

func main() {
	mrand.Seed(time.Now().UnixNano())

	// Load lessons
	dataPath := os.Getenv("LESSONS_FILE")
	if dataPath == "" {
		dataPath = filepath.Join("..","data", "lessons.json")
	}
	loaded, err := loadLessons(dataPath)
	if err != nil {
		log.Fatalf("failed to load lessons from %s: %v", dataPath, err)
	}
	lessonsByCat = map[string][]Lesson{}
	for _, l := range loaded {
		lessonsByCat[l.Category] = append(lessonsByCat[l.Category], l)
	}
	for cat := range lessonsByCat {
		categories = append(categories, cat)
	}
	sort.Strings(categories)

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

	port := os.Getenv("PORT")
	if port == "" {
		port = "8081"
	}
	log.Printf("Server listening on :%s", port)
	log.Fatal(http.ListenAndServe(":"+port, nil))
}

// ---------- Helpers ----------

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

func pickLessonForProfile(p *profile, cat string) *Lesson {
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
	if len(pool) == 0 {
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
			sessions[sid] = &profile{}
			r.AddCookie(&http.Cookie{Name: "sid", Value: sid})
		}
		next.ServeHTTP(w, r)
	})
}

func getProfile(r *http.Request) *profile {
	c, err := r.Cookie("sid")
	if err != nil {
		return &profile{}
	}
	p, ok := sessions[c.Value]
	if !ok {
		p = &profile{}
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
	lesson := pickLessonForProfile(p, cat)
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
			lesson := pickLessonForProfile(p, r.URL.Query().Get("category"))
			if lesson == nil {
				http.Error(w, "no lessons", http.StatusInternalServerError)
				return
			}
			p.LastLesson = lesson
			_ = json.NewEncoder(w).Encode(SessionState{
				Stage:  "lesson",
				Lesson: lesson,
			})
			return

		case "quiz":
			if len(p.CurrentQuiz) == 0 {
				http.Error(w, "no active quiz", http.StatusBadRequest)
				return
			}
			q := p.CurrentQuiz[p.QuizIndex]
			_ = json.NewEncoder(w).Encode(SessionState{
				Stage:    "quiz",
				Question: q.Question,
				Options:  q.Options,
				Index:    p.QuizIndex + 1,
				Total:    len(p.CurrentQuiz),
			})
			return
		}

	case http.MethodPost:
		switch stage {
		case "add":
			var body struct{ Title string `json:"title"` }
			if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
				http.Error(w, "bad body", http.StatusBadRequest)
				return
			}
			if body.Title != "" {
				p.LessonsSeen = uniqueStrings(append(p.LessonsSeen, body.Title))
			}
			_ = json.NewEncoder(w).Encode(map[string]any{
				"stage":        "added",
				"lessonsSeen":  p.LessonsSeen,
				"count":        len(p.LessonsSeen),
				"message":      "lesson added to study list",
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
			first := p.CurrentQuiz[0]
			_ = json.NewEncoder(w).Encode(SessionState{
				Stage:    "quiz",
				Question: first.Question,
				Options:  first.Options,
				Index:    1,
				Total:    len(p.CurrentQuiz),
				Message:  "quiz started",
			})
			return

		case "answer":
			if len(p.CurrentQuiz) == 0 {
				http.Error(w, "no active quiz", http.StatusBadRequest)
				return
			}
			var body struct{ AnswerIndex int `json:"answerIndex"` }
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
	for i := 0; i < 3 && i < len(pool); i++ { opts = append(opts, pool[i]) }
	for len(opts) < 4 { opts = append(opts, "This option does not apply to the concept.") }
	mrand.Shuffle(len(opts), func(i, j int) { opts[i], opts[j] = opts[j], opts[i] })
	correctIdx := 0
	for i, o := range opts {
		if o == correct { correctIdx = i; break }
	}
	return QuizQuestion{
		LessonTitle:  l.Title,
		Question:     question,
		Options:      opts,
		CorrectIndex: correctIdx,
	}
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
