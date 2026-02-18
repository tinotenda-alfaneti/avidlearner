package main

import (
	"encoding/json"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"avidlearner/internal/routes"
	. "avidlearner/internal/models"
)

func writeTempLessons(t *testing.T, lessons []Lesson) string {
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
	sample := []Lesson{{Title: "T1", Category: "c1", Text: "txt"}}
	path := writeTempLessons(t, sample)
	defer os.Remove(path)

	loaded, err := routes.LoadLessons(path)
	if err != nil {
		t.Fatalf("loadLessons error: %v", err)
	}
	lessonsByCat := map[string][]Lesson{}
	var categories []string
	for _, l := range loaded {
		lessonsByCat[l.Category] = append(lessonsByCat[l.Category], l)
	}
	for cat := range lessonsByCat {
		categories = append(categories, cat)
	}
	routes.SetLessonsByCategory(lessonsByCat)
	routes.SetCategories(categories)

	rr := httptest.NewRecorder()
	req := httptest.NewRequest("GET", "/api/lessons", nil)
	handler := http.HandlerFunc(routes.HandleLessons)
	handler.ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("expected 200 got %d", rr.Code)
	}
	var resp LessonsResponse
	if err := json.Unmarshal(rr.Body.Bytes(), &resp); err != nil {
		t.Fatalf("unmarshal resp: %v", err)
	}
	if len(resp.Categories) == 0 {
		t.Fatalf("expected categories, got none")
	}
}

func TestHandleRandom(t *testing.T) {
	sample := []Lesson{{Title: "T1", Category: "c1", Text: "txt"}, {Title: "T2", Category: "c1", Text: "txt2"}}
	path := writeTempLessons(t, sample)
	defer os.Remove(path)

	loaded, err := routes.LoadLessons(path)
	if err != nil {
		t.Fatalf("loadLessons error: %v", err)
	}
	lessonsByCat := map[string][]Lesson{}
	var categories []string
	for _, l := range loaded {
		lessonsByCat[l.Category] = append(lessonsByCat[l.Category], l)
	}
	for cat := range lessonsByCat {
		categories = append(categories, cat)
	}
	routes.SetLessonsByCategory(lessonsByCat)
	routes.SetCategories(categories)

	// Request any
	rr := httptest.NewRecorder()
	req := httptest.NewRequest("GET", "/api/random?category=any", nil)
	handler := http.HandlerFunc(routes.HandleRandom)
	handler.ServeHTTP(rr, req)
	if rr.Code != http.StatusOK {
		t.Fatalf("expected 200 got %d", rr.Code)
	}
	var got Lesson
	if err := json.Unmarshal(rr.Body.Bytes(), &got); err != nil {
		t.Fatalf("unmarshal got: %v", err)
	}
	if got.Title == "" {
		t.Fatalf("expected lesson, got empty")
	}
}
