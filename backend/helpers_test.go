package main

import (
	"os"
	"path/filepath"
	"testing"
)

func TestLoadLessons(t *testing.T) {
	t.Run("loads valid lessons file", func(t *testing.T) {
		content := `[
			{"title": "Test Lesson 1", "category": "Testing", "text": "Test content", "explain": "Explanation", "useCases": ["case1"], "tips": ["tip1"]},
			{"title": "Test Lesson 2", "category": "Testing", "text": "Test content 2", "explain": "Explanation 2", "useCases": ["case2"], "tips": ["tip2"]}
		]`

		tmpFile, err := os.CreateTemp("", "lessons-*.json")
		if err != nil {
			t.Fatalf("failed to create temp file: %v", err)
		}
		defer os.Remove(tmpFile.Name())

		if _, err := tmpFile.WriteString(content); err != nil {
			t.Fatalf("failed to write temp file: %v", err)
		}
		tmpFile.Close()

		lessons, err := loadLessons(tmpFile.Name())
		if err != nil {
			t.Fatalf("loadLessons failed: %v", err)
		}

		if len(lessons) != 2 {
			t.Errorf("expected 2 lessons, got %d", len(lessons))
		}

		if lessons[0].Title != "Test Lesson 1" {
			t.Errorf("expected title 'Test Lesson 1', got '%s'", lessons[0].Title)
		}

		if lessons[1].Category != "Testing" {
			t.Errorf("expected category 'Testing', got '%s'", lessons[1].Category)
		}
	})

	t.Run("returns error for non-existent file", func(t *testing.T) {
		_, err := loadLessons(filepath.Join("non", "existent", "file.json"))
		if err == nil {
			t.Error("expected error for non-existent file, got nil")
		}
	})

	t.Run("returns error for invalid JSON", func(t *testing.T) {
		tmpFile, err := os.CreateTemp("", "invalid-*.json")
		if err != nil {
			t.Fatalf("failed to create temp file: %v", err)
		}
		defer os.Remove(tmpFile.Name())

		tmpFile.WriteString("{invalid json")
		tmpFile.Close()

		_, err = loadLessons(tmpFile.Name())
		if err == nil {
			t.Error("expected error for invalid JSON, got nil")
		}
	})
}

func TestLoadProChallenges(t *testing.T) {
	t.Run("loads valid challenges file", func(t *testing.T) {
		content := `[
			{"id": "challenge1", "title": "Challenge 1", "difficulty": "medium", "topics": ["go"], "description": "Test", "starter": {"filename": "main.go", "code": "package main"}, "hints": ["hint1"], "reward": {"xp": 100, "coins": 50}},
			{"id": "challenge2", "title": "Challenge 2", "difficulty": "hard", "topics": ["algorithms"], "description": "Test 2", "starter": {"filename": "main.go", "code": "package main"}, "hints": ["hint2"], "reward": {"xp": 200, "coins": 100}}
		]`

		tmpFile, err := os.CreateTemp("", "challenges-*.json")
		if err != nil {
			t.Fatalf("failed to create temp file: %v", err)
		}
		defer os.Remove(tmpFile.Name())

		tmpFile.WriteString(content)
		tmpFile.Close()

		challenges, byID, err := loadProChallenges(tmpFile.Name())
		if err != nil {
			t.Fatalf("loadProChallenges failed: %v", err)
		}

		if len(challenges) != 2 {
			t.Errorf("expected 2 challenges, got %d", len(challenges))
		}

		if len(byID) != 2 {
			t.Errorf("expected 2 challenges in map, got %d", len(byID))
		}

		challenge, ok := byID["challenge1"]
		if !ok {
			t.Error("expected challenge1 in map")
		}

		if challenge.Title != "Challenge 1" {
			t.Errorf("expected title 'Challenge 1', got '%s'", challenge.Title)
		}

		if challenge.Difficulty != "medium" {
			t.Errorf("expected difficulty 'medium', got '%s'", challenge.Difficulty)
		}
	})

	t.Run("skips challenges without ID", func(t *testing.T) {
		content := `[
			{"id": "challenge1", "title": "Challenge 1", "difficulty": "easy", "topics": ["go"], "description": "Test", "starter": {"filename": "main.go", "code": ""}, "hints": [], "reward": {"xp": 50, "coins": 25}},
			{"title": "Challenge Without ID", "difficulty": "easy", "topics": ["go"], "description": "Test", "starter": {"filename": "main.go", "code": ""}, "hints": [], "reward": {"xp": 50, "coins": 25}}
		]`

		tmpFile, err := os.CreateTemp("", "challenges-*.json")
		if err != nil {
			t.Fatalf("failed to create temp file: %v", err)
		}
		defer os.Remove(tmpFile.Name())

		tmpFile.WriteString(content)
		tmpFile.Close()

		challenges, byID, err := loadProChallenges(tmpFile.Name())
		if err != nil {
			t.Fatalf("loadProChallenges failed: %v", err)
		}

		if len(challenges) != 2 {
			t.Errorf("expected 2 challenges in list, got %d", len(challenges))
		}

		if len(byID) != 1 {
			t.Errorf("expected 1 challenge in map (skipping one without ID), got %d", len(byID))
		}
	})
}

func TestPickRandomLesson(t *testing.T) {
	// Setup test data
	lessonsByCat = map[string][]Lesson{
		"Go":     {{Title: "Go Lesson 1", Category: "Go", Text: "text"}},
		"Python": {{Title: "Python Lesson 1", Category: "Python", Text: "text"}, {Title: "Python Lesson 2", Category: "Python", Text: "text"}},
	}

	t.Run("picks lesson from specific category", func(t *testing.T) {
		lesson := pickRandomLesson("Go")
		if lesson == nil {
			t.Fatal("expected lesson, got nil")
		}

		if lesson.Category != "Go" {
			t.Errorf("expected category 'Go', got '%s'", lesson.Category)
		}
	})

	t.Run("picks lesson from any category", func(t *testing.T) {
		lesson := pickRandomLesson("any")
		if lesson == nil {
			t.Fatal("expected lesson, got nil")
		}

		if lesson.Category != "Go" && lesson.Category != "Python" {
			t.Errorf("expected category 'Go' or 'Python', got '%s'", lesson.Category)
		}
	})

	t.Run("returns nil for non-existent category", func(t *testing.T) {
		lesson := pickRandomLesson("NonExistent")
		if lesson != nil {
			t.Error("expected nil for non-existent category, got lesson")
		}
	})

	t.Run("returns nil for empty lessons", func(t *testing.T) {
		originalLessons := lessonsByCat
		lessonsByCat = map[string][]Lesson{}

		lesson := pickRandomLesson("any")
		if lesson != nil {
			t.Error("expected nil for empty lessons, got lesson")
		}

		lessonsByCat = originalLessons
	})
}

func TestAllLessons(t *testing.T) {
	lessonsByCat = map[string][]Lesson{
		"Go":     {{Title: "Go 1", Category: "Go", Text: "text"}},
		"Python": {{Title: "Python 1", Category: "Python", Text: "text"}, {Title: "Python 2", Category: "Python", Text: "text"}},
	}

	lessons := allLessons()

	if len(lessons) != 3 {
		t.Errorf("expected 3 lessons, got %d", len(lessons))
	}
}

func TestFindLessonByTitle(t *testing.T) {
	lessonsByCat = map[string][]Lesson{
		"Go":     {{Title: "Go Basics", Category: "Go", Text: "text"}},
		"Python": {{Title: "Python Basics", Category: "Python", Text: "text"}},
	}

	t.Run("finds existing lesson", func(t *testing.T) {
		lesson := findLessonByTitle("Go Basics")
		if lesson == nil {
			t.Fatal("expected lesson, got nil")
		}

		if lesson.Title != "Go Basics" {
			t.Errorf("expected title 'Go Basics', got '%s'", lesson.Title)
		}

		if lesson.Category != "Go" {
			t.Errorf("expected category 'Go', got '%s'", lesson.Category)
		}
	})

	t.Run("returns nil for non-existent lesson", func(t *testing.T) {
		lesson := findLessonByTitle("Non Existent")
		if lesson != nil {
			t.Error("expected nil for non-existent lesson, got lesson")
		}
	})
}

func TestUniqueStrings(t *testing.T) {
	t.Run("removes duplicates", func(t *testing.T) {
		input := []string{"a", "b", "c", "a", "b", "d"}
		result := uniqueStrings(input)

		if len(result) != 4 {
			t.Errorf("expected 4 unique strings, got %d", len(result))
		}

		expected := map[string]bool{"a": true, "b": true, "c": true, "d": true}
		for _, s := range result {
			if !expected[s] {
				t.Errorf("unexpected string in result: %s", s)
			}
		}
	})

	t.Run("preserves order of first occurrence", func(t *testing.T) {
		input := []string{"z", "a", "z", "b", "a"}
		result := uniqueStrings(input)

		if len(result) != 3 {
			t.Errorf("expected 3 unique strings, got %d", len(result))
		}

		if result[0] != "z" || result[1] != "a" || result[2] != "b" {
			t.Errorf("expected order [z, a, b], got %v", result)
		}
	})

	t.Run("handles empty slice", func(t *testing.T) {
		result := uniqueStrings([]string{})
		if len(result) != 0 {
			t.Errorf("expected empty result, got %d items", len(result))
		}
	})
}

func TestPickLessonForProfile(t *testing.T) {
	lessonsByCat = map[string][]Lesson{
		"Go": {
			{Title: "Lesson 1", Category: "Go", Text: "text"},
			{Title: "Lesson 2", Category: "Go", Text: "text"},
			{Title: "Lesson 3", Category: "Go", Text: "text"},
		},
	}

	t.Run("picks lesson and updates recent history", func(t *testing.T) {
		p := newProfile()
		lesson := pickLessonForProfile(p, "Go", "")

		if lesson == nil {
			t.Fatal("expected lesson, got nil")
		}

		if len(p.RecentLessons) != 1 {
			t.Errorf("expected 1 recent lesson, got %d", len(p.RecentLessons))
		}

		if p.RecentLessons[0] != lesson.Title {
			t.Errorf("expected recent lesson to be '%s', got '%s'", lesson.Title, p.RecentLessons[0])
		}
	})

	t.Run("avoids recently seen lessons", func(t *testing.T) {
		p := newProfile()
		p.RecentLessons = []string{"Lesson 1", "Lesson 2"}

		seen := make(map[string]int)
		for i := 0; i < 10; i++ {
			lesson := pickLessonForProfile(p, "Go", "")
			if lesson != nil {
				seen[lesson.Title]++
			}
		}

		// Should prefer Lesson 3 since 1 and 2 are recent
		if seen["Lesson 3"] < seen["Lesson 1"] || seen["Lesson 3"] < seen["Lesson 2"] {
			t.Error("expected Lesson 3 to be picked more often than recent lessons")
		}
	})

	t.Run("trims history when it exceeds window", func(t *testing.T) {
		p := newProfile()
		// Fill with many lessons
		for i := 0; i < lessonRepeatWindow*3; i++ {
			pickLessonForProfile(p, "Go", "")
		}

		if len(p.RecentLessons) > lessonRepeatWindow*2 {
			t.Errorf("expected recent lessons to be trimmed to <= %d, got %d", lessonRepeatWindow*2, len(p.RecentLessons))
		}
	})
}
