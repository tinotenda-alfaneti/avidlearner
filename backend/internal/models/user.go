package models

import "time"

type SavedLesson struct {
	Title    string    `json:"title"`
	Category string    `json:"category"`
	Source   string    `json:"source,omitempty"`
	SavedAt  time.Time `json:"savedAt"`
}

type UserStats struct {
	LessonsRead       int       `json:"lessonsRead"`
	QuizzesTaken      int       `json:"quizzesTaken"`
	QuizCorrect       int       `json:"quizCorrect"`
	TypingSessions    int       `json:"typingSessions"`
	CodingSubmissions int       `json:"codingSubmissions"`
	CodingPassed      int       `json:"codingPassed"`
	LastActive        time.Time `json:"lastActive"`
}

type UserProfile struct {
	Coins        int           `json:"coins"`
	XP           int           `json:"xp"`
	QuizStreak   int           `json:"quizStreak"`
	TypingStreak int           `json:"typingStreak"`
	TypingBest   int           `json:"typingBest"`
	CodingScore  int           `json:"codingScore"`
	LessonsSeen  []string      `json:"lessonsSeen"`
	SavedLessons []SavedLesson `json:"savedLessons"`
	Stats        UserStats     `json:"stats"`
	UpdatedAt    time.Time     `json:"updatedAt"`
}

type User struct {
	ID               string      `json:"id"`
	Username         string      `json:"username"`
	PasswordHash     string      `json:"passwordHash"`
	CreatedAt        time.Time   `json:"createdAt"`
	LeaderboardOptIn bool        `json:"leaderboardOptIn"`
	Profile          UserProfile `json:"profile"`
}

type UserPublic struct {
	ID               string      `json:"id"`
	Username         string      `json:"username"`
	CreatedAt        time.Time   `json:"createdAt"`
	LeaderboardOptIn bool        `json:"leaderboardOptIn"`
	Profile          UserProfile `json:"profile"`
}

func (u User) Public() UserPublic {
	return UserPublic{
		ID:               u.ID,
		Username:         u.Username,
		CreatedAt:        u.CreatedAt,
		LeaderboardOptIn: u.LeaderboardOptIn,
		Profile:          u.Profile,
	}
}
