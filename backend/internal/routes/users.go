package routes

import (
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"sync"
	"time"

	"avidlearner/internal/models"
)

var (
	usersMu     sync.RWMutex
	usersByID   = map[string]*models.User{}
	usersByName = map[string]*models.User{}
	usersDirty  bool
	usersPath   string
)

func LoadUsers(path string) error {
	usersMu.Lock()
	defer usersMu.Unlock()

	usersPath = path
	usersByID = map[string]*models.User{}
	usersByName = map[string]*models.User{}
	usersDirty = false

	b, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return nil
		}
		return err
	}

	var list []models.User
	if err := json.Unmarshal(b, &list); err != nil {
		return err
	}

	for i := range list {
		u := &list[i]
		if u.ID == "" || u.Username == "" {
			continue
		}
		normalized := normalizeUsername(u.Username)
		usersByID[u.ID] = u
		usersByName[normalized] = u
		ensureProfileDefaults(&u.Profile)
	}

	return nil
}

func SaveUsers(path string) error {
	if strings.TrimSpace(path) == "" {
		return errors.New("users file path not set")
	}
	usersMu.RLock()
	if !usersDirty {
		usersMu.RUnlock()
		return nil
	}

	snapshot := make([]models.User, 0, len(usersByID))
	for _, u := range usersByID {
		snapshot = append(snapshot, *u)
	}
	usersMu.RUnlock()

	sort.Slice(snapshot, func(i, j int) bool {
		return snapshot[i].CreatedAt.Before(snapshot[j].CreatedAt)
	})

	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return err
	}

	b, err := json.MarshalIndent(snapshot, "", "  ")
	if err != nil {
		return err
	}

	if err := os.WriteFile(path, b, 0o644); err != nil {
		return err
	}

	usersMu.Lock()
	usersDirty = false
	usersMu.Unlock()

	return nil
}

func usersPathOrDefault() string {
	usersMu.RLock()
	defer usersMu.RUnlock()
	return usersPath
}

func markUsersDirty() {
	usersMu.Lock()
	usersDirty = true
	usersMu.Unlock()
}

func normalizeUsername(username string) string {
	return strings.ToLower(strings.TrimSpace(username))
}

func getUserByName(username string) *models.User {
	key := normalizeUsername(username)
	usersMu.RLock()
	defer usersMu.RUnlock()
	return usersByName[key]
}

func getUserByID(id string) *models.User {
	usersMu.RLock()
	defer usersMu.RUnlock()
	return usersByID[id]
}

func addUser(u *models.User) error {
	if u == nil {
		return errors.New("user required")
	}
	normalized := normalizeUsername(u.Username)
	if normalized == "" || u.ID == "" {
		return errors.New("invalid user")
	}

	usersMu.Lock()
	defer usersMu.Unlock()
	if _, exists := usersByName[normalized]; exists {
		return errors.New("username already exists")
	}
	usersByID[u.ID] = u
	usersByName[normalized] = u
	usersDirty = true
	return nil
}

func updateUserByID(id string, fn func(*models.User)) {
	if id == "" || fn == nil {
		return
	}
	usersMu.Lock()
	defer usersMu.Unlock()
	u := usersByID[id]
	if u == nil {
		return
	}
	fn(u)
	usersDirty = true
}

func ensureProfileDefaults(profile *models.UserProfile) {
	if profile == nil {
		return
	}
	if profile.LessonsSeen == nil {
		profile.LessonsSeen = []string{}
	}
	if profile.SavedLessons == nil {
		profile.SavedLessons = []models.SavedLesson{}
	}
	if profile.UpdatedAt.IsZero() {
		profile.UpdatedAt = time.Now()
	}
}
