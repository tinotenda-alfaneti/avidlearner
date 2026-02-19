package routes

import (
	"encoding/json"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"avidlearner/internal/models"
)

func writeTempLessons(t *testing.T, lessons []models.Lesson) string {
	t.Helper()
	b, err := json.Marshal(lessons)
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}
	tmp, err := ioutil.TempFile("", "lessons-*.json")
	if err != nil {
		t.Fatalf("tempfile: %v", err)
	}
	if _, err := tmp.Write(b); err != nil {
		t.Fatalf("write tmp: %v", err)
	}
	if err := tmp.Close(); err != nil {
		t.Fatalf("close tmp: %v", err)
	}
	return tmp.Name()
}

func TestHandleLessons(t *testing.T) {
	sample := []models.Lesson{{Title: "T1", Category: "c1", Text: "txt"}}
	path := writeTempLessons(t, sample)
	defer os.Remove(path)

	loaded, err := loadLessons(path)
	if err != nil {
		t.Fatalf("loadLessons error: %v", err)
	}
	lessonsByCat = map[string][]models.Lesson{}
	categories = nil
	for _, l := range loaded {
		lessonsByCat[l.Category] = append(lessonsByCat[l.Category], l)
	}
	for cat := range lessonsByCat {
		categories = append(categories, cat)
	}

	rr := httptest.NewRecorder()
	req := httptest.NewRequest("GET", "/api/lessons", nil)
	handler := http.HandlerFunc(handleLessons)
	handler.ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("expected 200 got %d", rr.Code)
	}
	var resp models.LessonsResponse
	if err := json.Unmarshal(rr.Body.Bytes(), &resp); err != nil {
		t.Fatalf("unmarshal resp: %v", err)
	}
	if len(resp.Categories) == 0 {
		t.Fatalf("expected categories, got none")
	}
}

func TestHandleRandom(t *testing.T) {
	sample := []models.Lesson{{Title: "T1", Category: "c1", Text: "txt"}, {Title: "T2", Category: "c1", Text: "txt2"}}
	path := writeTempLessons(t, sample)
	defer os.Remove(path)

	loaded, err := loadLessons(path)
	if err != nil {
		t.Fatalf("loadLessons error: %v", err)
	}
	lessonsByCat = map[string][]models.Lesson{}
	categories = nil
	for _, l := range loaded {
		lessonsByCat[l.Category] = append(lessonsByCat[l.Category], l)
	}
	for cat := range lessonsByCat {
		categories = append(categories, cat)
	}

	// Request any
	rr := httptest.NewRecorder()
	req := httptest.NewRequest("GET", "/api/random?category=any", nil)
	handler := http.HandlerFunc(handleRandom)
	handler.ServeHTTP(rr, req)
	if rr.Code != http.StatusOK {
		t.Fatalf("expected 200 got %d", rr.Code)
	}
	var got models.Lesson
	if err := json.Unmarshal(rr.Body.Bytes(), &got); err != nil {
		t.Fatalf("unmarshal got: %v", err)
	}
	if got.Title == "" {
		t.Fatalf("expected lesson, got empty")
	}
}
