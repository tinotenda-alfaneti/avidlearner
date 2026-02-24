package routes

import (
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"errors"
	"net/http"
	"regexp"
	"strings"
	"time"

	"avidlearner/internal/auth"
	"avidlearner/internal/models"
)

var (
	authManager *auth.Manager
)

var usernamePattern = regexp.MustCompile(`^[a-zA-Z0-9._-]{3,30}$`)

func SetAuthConfig(secret string, ttl time.Duration) error {
	manager, err := auth.NewManager(secret, ttl)
	if err != nil {
		return err
	}
	authManager = manager
	return nil
}

func handleSignup(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	if r.Method != http.MethodPost {
		http.Error(w, `{"error":"method not allowed"}`, http.StatusMethodNotAllowed)
		return
	}
	if authManager == nil {
		http.Error(w, `{"error":"auth not configured"}`, http.StatusServiceUnavailable)
		return
	}

	var req struct {
		Username         string `json:"username"`
		Password         string `json:"password"`
		LeaderboardOptIn bool   `json:"leaderboardOptIn"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, `{"error":"invalid request body"}`, http.StatusBadRequest)
		return
	}

	req.Username = strings.TrimSpace(req.Username)
	if !usernamePattern.MatchString(req.Username) {
		http.Error(w, `{"error":"username must be 3-30 chars and use letters, numbers, '.', '_' or '-' only"}`, http.StatusBadRequest)
		return
	}
	if len(req.Password) < 8 {
		http.Error(w, `{"error":"password must be at least 8 characters"}`, http.StatusBadRequest)
		return
	}
	if !req.LeaderboardOptIn {
		http.Error(w, `{"error":"leaderboard opt-in is required to create an account"}`, http.StatusBadRequest)
		return
	}

	if existing := getUserByName(req.Username); existing != nil {
		http.Error(w, `{"error":"username already exists"}`, http.StatusConflict)
		return
	}

	hash, err := auth.HashPassword(req.Password)
	if err != nil {
		http.Error(w, `{"error":"unable to create account"}`, http.StatusInternalServerError)
		return
	}

	userID, err := randomID()
	if err != nil {
		http.Error(w, `{"error":"unable to create account"}`, http.StatusInternalServerError)
		return
	}

	now := time.Now()
	user := &models.User{
		ID:               userID,
		Username:         req.Username,
		PasswordHash:     hash,
		CreatedAt:        now,
		LeaderboardOptIn: req.LeaderboardOptIn,
		Profile: models.UserProfile{
			Coins:        0,
			XP:           0,
			QuizStreak:   0,
			TypingStreak: 0,
			TypingBest:   0,
			CodingScore:  0,
			LessonsSeen:  []string{},
			SavedLessons: []models.SavedLesson{},
			Stats: models.UserStats{
				LastActive: now,
			},
			UpdatedAt: now,
		},
	}

	if err := addUser(user); err != nil {
		http.Error(w, `{"error":"username already exists"}`, http.StatusConflict)
		return
	}

	token, err := authManager.IssueToken(user.ID, user.Username)
	if err != nil {
		http.Error(w, `{"error":"unable to issue token"}`, http.StatusInternalServerError)
		return
	}

	_ = SaveUsers(usersPathOrDefault())

	resp := map[string]any{
		"token": token,
		"user":  user.Public(),
	}
	_ = json.NewEncoder(w).Encode(resp)
}

func handleLogin(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	if r.Method != http.MethodPost {
		http.Error(w, `{"error":"method not allowed"}`, http.StatusMethodNotAllowed)
		return
	}
	if authManager == nil {
		http.Error(w, `{"error":"auth not configured"}`, http.StatusServiceUnavailable)
		return
	}

	var req struct {
		Username string `json:"username"`
		Password string `json:"password"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, `{"error":"invalid request body"}`, http.StatusBadRequest)
		return
	}

	req.Username = strings.TrimSpace(req.Username)
	user := getUserByName(req.Username)
	if user == nil || !auth.CheckPassword(user.PasswordHash, req.Password) {
		http.Error(w, `{"error":"invalid credentials"}`, http.StatusUnauthorized)
		return
	}

	token, err := authManager.IssueToken(user.ID, user.Username)
	if err != nil {
		http.Error(w, `{"error":"unable to issue token"}`, http.StatusInternalServerError)
		return
	}

	updateUserByID(user.ID, func(u *models.User) {
		u.Profile.Stats.LastActive = time.Now()
		u.Profile.UpdatedAt = time.Now()
	})

	resp := map[string]any{
		"token": token,
		"user":  user.Public(),
	}
	_ = json.NewEncoder(w).Encode(resp)
}

func handleMe(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	if r.Method != http.MethodGet {
		http.Error(w, `{"error":"method not allowed"}`, http.StatusMethodNotAllowed)
		return
	}
	user, err := requireAuthUser(w, r)
	if err != nil {
		return
	}
	_ = json.NewEncoder(w).Encode(user.Public())
}

func handleProfile(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	user, err := requireAuthUser(w, r)
	if err != nil {
		return
	}

	switch r.Method {
	case http.MethodGet:
		_ = json.NewEncoder(w).Encode(user.Public())
		return
	case http.MethodPatch:
		var req struct {
			Coins        *int               `json:"coins"`
			XP           *int               `json:"xp"`
			QuizStreak   *int               `json:"quizStreak"`
			TypingStreak *int               `json:"typingStreak"`
			TypingBest   *int               `json:"typingBest"`
			CodingScore  *int               `json:"codingScore"`
			LessonsSeen  []string           `json:"lessonsSeen"`
			Stats        *models.UserStats  `json:"stats"`
		}
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			http.Error(w, `{"error":"invalid request body"}`, http.StatusBadRequest)
			return
		}

		updateUserByID(user.ID, func(u *models.User) {
			if req.Coins != nil {
				u.Profile.Coins = *req.Coins
			}
			if req.XP != nil {
				u.Profile.XP = *req.XP
			}
			if req.QuizStreak != nil {
				u.Profile.QuizStreak = *req.QuizStreak
			}
			if req.TypingStreak != nil {
				u.Profile.TypingStreak = *req.TypingStreak
			}
			if req.TypingBest != nil {
				if *req.TypingBest > u.Profile.TypingBest {
					u.Profile.TypingBest = *req.TypingBest
				}
			}
			if req.CodingScore != nil {
				u.Profile.CodingScore = *req.CodingScore
			}
			if req.LessonsSeen != nil {
				u.Profile.LessonsSeen = dedupeStrings(req.LessonsSeen)
				u.Profile.Stats.LessonsRead = len(u.Profile.LessonsSeen)
			}
			if req.Stats != nil {
				u.Profile.Stats.QuizzesTaken = req.Stats.QuizzesTaken
				u.Profile.Stats.QuizCorrect = req.Stats.QuizCorrect
				u.Profile.Stats.TypingSessions = req.Stats.TypingSessions
				u.Profile.Stats.CodingSubmissions = req.Stats.CodingSubmissions
				u.Profile.Stats.CodingPassed = req.Stats.CodingPassed
				u.Profile.Stats.LastActive = time.Now()
			} else {
				u.Profile.Stats.LastActive = time.Now()
			}
			u.Profile.UpdatedAt = time.Now()
		})

		updated := getUserByID(user.ID)
		if updated == nil {
			http.Error(w, `{"error":"user not found"}`, http.StatusNotFound)
			return
		}
		_ = json.NewEncoder(w).Encode(updated.Public())
		return
	default:
		http.Error(w, `{"error":"method not allowed"}`, http.StatusMethodNotAllowed)
		return
	}
}

func handleSaveLesson(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	user, err := requireAuthUser(w, r)
	if err != nil {
		return
	}
	if r.Method != http.MethodPost {
		http.Error(w, `{"error":"method not allowed"}`, http.StatusMethodNotAllowed)
		return
	}

	var req struct {
		Title    string `json:"title"`
		Category string `json:"category"`
		Source   string `json:"source"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, `{"error":"invalid request body"}`, http.StatusBadRequest)
		return
	}
	req.Title = strings.TrimSpace(req.Title)
	req.Category = strings.TrimSpace(req.Category)
	if req.Title == "" || req.Category == "" {
		http.Error(w, `{"error":"title and category required"}`, http.StatusBadRequest)
		return
	}

	updateUserByID(user.ID, func(u *models.User) {
		ensureProfileDefaults(&u.Profile)
		for _, saved := range u.Profile.SavedLessons {
			if strings.EqualFold(saved.Title, req.Title) && strings.EqualFold(saved.Category, req.Category) {
				return
			}
		}
		u.Profile.SavedLessons = append(u.Profile.SavedLessons, models.SavedLesson{
			Title:    req.Title,
			Category: req.Category,
			Source:   req.Source,
			SavedAt:  time.Now(),
		})
		u.Profile.UpdatedAt = time.Now()
		u.Profile.Stats.LastActive = time.Now()
	})

	updated := getUserByID(user.ID)
	if updated == nil {
		http.Error(w, `{"error":"user not found"}`, http.StatusNotFound)
		return
	}
	_ = json.NewEncoder(w).Encode(updated.Public())
}

func handleRemoveLesson(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	user, err := requireAuthUser(w, r)
	if err != nil {
		return
	}
	if r.Method != http.MethodPost {
		http.Error(w, `{"error":"method not allowed"}`, http.StatusMethodNotAllowed)
		return
	}

	var req struct {
		Title    string `json:"title"`
		Category string `json:"category"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, `{"error":"invalid request body"}`, http.StatusBadRequest)
		return
	}
	req.Title = strings.TrimSpace(req.Title)
	req.Category = strings.TrimSpace(req.Category)
	if req.Title == "" {
		http.Error(w, `{"error":"title required"}`, http.StatusBadRequest)
		return
	}

	updateUserByID(user.ID, func(u *models.User) {
		ensureProfileDefaults(&u.Profile)
		filtered := u.Profile.SavedLessons[:0]
		for _, saved := range u.Profile.SavedLessons {
			if strings.EqualFold(saved.Title, req.Title) {
				if req.Category == "" || strings.EqualFold(saved.Category, req.Category) {
					continue
				}
			}
			filtered = append(filtered, saved)
		}
		u.Profile.SavedLessons = filtered
		u.Profile.UpdatedAt = time.Now()
		u.Profile.Stats.LastActive = time.Now()
	})

	updated := getUserByID(user.ID)
	if updated == nil {
		http.Error(w, `{"error":"user not found"}`, http.StatusNotFound)
		return
	}
	_ = json.NewEncoder(w).Encode(updated.Public())
}

func requireAuthUser(w http.ResponseWriter, r *http.Request) (*models.User, error) {
	user, err := authUserFromRequest(r)
	if err != nil {
		http.Error(w, `{"error":"unauthorized"}`, http.StatusUnauthorized)
		return nil, err
	}
	return user, nil
}

func authUserFromRequest(r *http.Request) (*models.User, error) {
	if authManager == nil {
		return nil, errors.New("auth not configured")
	}
	token := bearerToken(r)
	if token == "" {
		return nil, errors.New("missing token")
	}
	claims, err := authManager.ParseToken(token)
	if err != nil {
		return nil, err
	}
	user := getUserByID(claims.Sub)
	if user == nil {
		return nil, errors.New("user not found")
	}
	if !strings.EqualFold(user.Username, claims.Username) {
		return nil, errors.New("user mismatch")
	}
	return user, nil
}

func bearerToken(r *http.Request) string {
	authHeader := strings.TrimSpace(r.Header.Get("Authorization"))
	if authHeader == "" {
		return ""
	}
	parts := strings.SplitN(authHeader, " ", 2)
	if len(parts) != 2 {
		return ""
	}
	if !strings.EqualFold(parts[0], "Bearer") {
		return ""
	}
	return strings.TrimSpace(parts[1])
}

func randomID() (string, error) {
	buf := make([]byte, 16)
	if _, err := rand.Read(buf); err != nil {
		return "", err
	}
	return hex.EncodeToString(buf), nil
}

func dedupeStrings(values []string) []string {
	seen := map[string]struct{}{}
	out := make([]string, 0, len(values))
	for _, v := range values {
		val := strings.TrimSpace(v)
		if val == "" {
			continue
		}
		key := strings.ToLower(val)
		if _, ok := seen[key]; ok {
			continue
		}
		seen[key] = struct{}{}
		out = append(out, val)
	}
	return out
}
