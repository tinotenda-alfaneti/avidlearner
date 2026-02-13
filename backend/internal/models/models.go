package models

import "time"

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