package main

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"avidlearner/internal/routes"
)

func TestWithSession(t *testing.T) {
	handler := routes.WithSession(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("ok"))
	}))

	t.Run("creates new session when no cookie", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/", nil)
		rr := httptest.NewRecorder()

		handler.ServeHTTP(rr, req)

		if rr.Code != http.StatusOK {
			t.Errorf("expected status 200, got %d", rr.Code)
		}

		cookies := rr.Result().Cookies()
		if len(cookies) == 0 {
			t.Fatal("expected session cookie to be set")
		}

		var sessionCookie *http.Cookie
		for _, c := range cookies {
			if c.Name == "sid" {
				sessionCookie = c
				break
			}
		}

		if sessionCookie == nil {
			t.Fatal("expected sid cookie")
		}

		if sessionCookie.Value == "" {
			t.Error("session cookie value should not be empty")
		}

		if sessionCookie.HttpOnly != true {
			t.Error("session cookie should be HttpOnly")
		}

		if sessionCookie.Path != "/" {
			t.Errorf("expected cookie path /, got %s", sessionCookie.Path)
		}
	})

	t.Run("preserves existing session", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/", nil)
		req.AddCookie(&http.Cookie{Name: "sid", Value: "existing-session-id"})
		rr := httptest.NewRecorder()

		routes.SetSession("existing-session-id", routes.NewProfile())

		handler.ServeHTTP(rr, req)

		if rr.Code != http.StatusOK {
			t.Errorf("expected status 200, got %d", rr.Code)
		}
	})
}

func TestGetProfile(t *testing.T) {
	// Clear sessions
	routes.ResetSessions()

	t.Run("returns new profile when no cookie", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/", nil)
		p := routes.GetProfile(req)

		if p == nil {
			t.Fatal("expected profile, got nil")
		}

		if p.Coins != 0 {
			t.Errorf("expected 0 coins, got %d", p.Coins)
		}

		if p.HintIdx == nil {
			t.Error("expected HintIdx to be initialized")
		}
	})

	t.Run("returns existing profile", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/", nil)
		req.AddCookie(&http.Cookie{Name: "sid", Value: "test-session"})

		expectedProfile := routes.NewProfile()
		expectedProfile.Coins = 100
		expectedProfile.XP = 50
		routes.SetSession("test-session", expectedProfile)

		p := routes.GetProfile(req)

		if p == nil {
			t.Fatal("expected profile, got nil")
		}

		if p.Coins != 100 {
			t.Errorf("expected 100 coins, got %d", p.Coins)
		}

		if p.XP != 50 {
			t.Errorf("expected 50 XP, got %d", p.XP)
		}
	})

	t.Run("creates profile for unknown session", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/", nil)
		req.AddCookie(&http.Cookie{Name: "sid", Value: "unknown-session"})

		p := routes.GetProfile(req)

		if p == nil {
			t.Fatal("expected profile, got nil")
		}

		// Should create and store new profile
		if _, ok := routes.Sessions()["unknown-session"]; !ok {
			t.Error("expected profile to be stored in sessions")
		}
	})
}

func TestNewProfile(t *testing.T) {
	p := routes.NewProfile()

	if p == nil {
		t.Fatal("expected profile, got nil")
	}

	if p.Coins != 0 {
		t.Errorf("expected 0 coins, got %d", p.Coins)
	}

	if p.Streak != 0 {
		t.Errorf("expected 0 streak, got %d", p.Streak)
	}

	if p.XP != 0 {
		t.Errorf("expected 0 XP, got %d", p.XP)
	}

	if p.HintIdx == nil {
		t.Error("expected HintIdx to be initialized")
	}

	if len(p.HintIdx) != 0 {
		t.Errorf("expected empty HintIdx map, got %d items", len(p.HintIdx))
	}

	if p.LessonsSeen != nil {
		t.Error("expected nil LessonsSeen slice initially")
	}

	if p.RecentLessons != nil {
		t.Error("expected nil RecentLessons slice initially")
	}
}
