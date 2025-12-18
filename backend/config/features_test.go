package config

import (
	"os"
	"testing"
)

func TestGetFeatureFlags(t *testing.T) {
	t.Run("returns singleton instance", func(t *testing.T) {
		ff1 := GetFeatureFlags()
		ff2 := GetFeatureFlags()

		if ff1 != ff2 {
			t.Error("expected same instance, got different instances")
		}
	})

	t.Run("initializes with default values when env not set", func(t *testing.T) {
		// This test just verifies the singleton works
		ff := GetFeatureFlags()

		// Check defaults (actual values depend on environment)
		if ff.GetAIProvider() == "" {
			t.Error("expected non-empty provider")
		}

		if ff.GetAIModel() == "" {
			t.Error("expected non-empty model")
		}

		if ff.GetMaxAILessonsPerDay() <= 0 {
			t.Error("expected positive max lessons")
		}
	})
}

func TestIsAILessonsEnabled(t *testing.T) {
	ff := GetFeatureFlags()

	// Test initial state
	initialState := ff.IsAILessonsEnabled()

	// Test toggling
	ff.SetAILessonsEnabled(!initialState)
	if ff.IsAILessonsEnabled() == initialState {
		t.Error("expected state to change after SetAILessonsEnabled")
	}

	ff.SetAILessonsEnabled(initialState)
	if ff.IsAILessonsEnabled() != initialState {
		t.Error("expected state to return to initial value")
	}
}

func TestSetAILessonsEnabled(t *testing.T) {
	ff := GetFeatureFlags()

	t.Run("enables AI lessons", func(t *testing.T) {
		ff.SetAILessonsEnabled(true)
		if !ff.IsAILessonsEnabled() {
			t.Error("expected AI lessons to be enabled")
		}
	})

	t.Run("disables AI lessons", func(t *testing.T) {
		ff.SetAILessonsEnabled(false)
		if ff.IsAILessonsEnabled() {
			t.Error("expected AI lessons to be disabled")
		}
	})
}

func TestGetAIProvider(t *testing.T) {
	ff := GetFeatureFlags()
	provider := ff.GetAIProvider()

	if provider == "" {
		t.Error("expected non-empty provider")
	}

	// Should be one of the valid providers
	validProviders := map[string]bool{"openai": true, "anthropic": true, "local": true}
	if !validProviders[provider] {
		t.Errorf("unexpected provider '%s'", provider)
	}
}

func TestGetAIModel(t *testing.T) {
	ff := GetFeatureFlags()
	model := ff.GetAIModel()

	if model == "" {
		t.Error("expected non-empty model")
	}
}

func TestGetMaxAILessonsPerDay(t *testing.T) {
	ff := GetFeatureFlags()
	maxLessons := ff.GetMaxAILessonsPerDay()

	if maxLessons <= 0 {
		t.Errorf("expected positive max lessons, got %d", maxLessons)
	}
}

func TestGetEnvBool(t *testing.T) {
	t.Run("returns default when env var not set", func(t *testing.T) {
		os.Unsetenv("TEST_BOOL")
		result := getEnvBool("TEST_BOOL", true)
		if !result {
			t.Error("expected default value true")
		}
	})

	t.Run("parses true value", func(t *testing.T) {
		os.Setenv("TEST_BOOL", "true")
		defer os.Unsetenv("TEST_BOOL")

		result := getEnvBool("TEST_BOOL", false)
		if !result {
			t.Error("expected true")
		}
	})

	t.Run("parses false value", func(t *testing.T) {
		os.Setenv("TEST_BOOL", "false")
		defer os.Unsetenv("TEST_BOOL")

		result := getEnvBool("TEST_BOOL", true)
		if result {
			t.Error("expected false")
		}
	})

	t.Run("returns default on parse error", func(t *testing.T) {
		os.Setenv("TEST_BOOL", "invalid")
		defer os.Unsetenv("TEST_BOOL")

		result := getEnvBool("TEST_BOOL", true)
		if !result {
			t.Error("expected default value true on parse error")
		}
	})

	t.Run("parses 1 as true", func(t *testing.T) {
		os.Setenv("TEST_BOOL", "1")
		defer os.Unsetenv("TEST_BOOL")

		result := getEnvBool("TEST_BOOL", false)
		if !result {
			t.Error("expected true for value '1'")
		}
	})

	t.Run("parses 0 as false", func(t *testing.T) {
		os.Setenv("TEST_BOOL", "0")
		defer os.Unsetenv("TEST_BOOL")

		result := getEnvBool("TEST_BOOL", true)
		if result {
			t.Error("expected false for value '0'")
		}
	})
}

func TestGetEnvString(t *testing.T) {
	t.Run("returns default when env var not set", func(t *testing.T) {
		os.Unsetenv("TEST_STRING")
		result := getEnvString("TEST_STRING", "default")
		if result != "default" {
			t.Errorf("expected 'default', got '%s'", result)
		}
	})

	t.Run("returns env value when set", func(t *testing.T) {
		os.Setenv("TEST_STRING", "custom")
		defer os.Unsetenv("TEST_STRING")

		result := getEnvString("TEST_STRING", "default")
		if result != "custom" {
			t.Errorf("expected 'custom', got '%s'", result)
		}
	})

	t.Run("returns empty string when env is empty", func(t *testing.T) {
		os.Setenv("TEST_STRING", "")
		defer os.Unsetenv("TEST_STRING")

		result := getEnvString("TEST_STRING", "default")
		if result != "default" {
			t.Errorf("expected default for empty env var, got '%s'", result)
		}
	})
}

func TestGetEnvInt(t *testing.T) {
	t.Run("returns default when env var not set", func(t *testing.T) {
		os.Unsetenv("TEST_INT")
		result := getEnvInt("TEST_INT", 42)
		if result != 42 {
			t.Errorf("expected 42, got %d", result)
		}
	})

	t.Run("parses valid integer", func(t *testing.T) {
		os.Setenv("TEST_INT", "100")
		defer os.Unsetenv("TEST_INT")

		result := getEnvInt("TEST_INT", 42)
		if result != 100 {
			t.Errorf("expected 100, got %d", result)
		}
	})

	t.Run("returns default on parse error", func(t *testing.T) {
		os.Setenv("TEST_INT", "invalid")
		defer os.Unsetenv("TEST_INT")

		result := getEnvInt("TEST_INT", 42)
		if result != 42 {
			t.Errorf("expected default 42 on parse error, got %d", result)
		}
	})

	t.Run("parses negative integer", func(t *testing.T) {
		os.Setenv("TEST_INT", "-10")
		defer os.Unsetenv("TEST_INT")

		result := getEnvInt("TEST_INT", 42)
		if result != -10 {
			t.Errorf("expected -10, got %d", result)
		}
	})

	t.Run("parses zero", func(t *testing.T) {
		os.Setenv("TEST_INT", "0")
		defer os.Unsetenv("TEST_INT")

		result := getEnvInt("TEST_INT", 42)
		if result != 0 {
			t.Errorf("expected 0, got %d", result)
		}
	})
}

func TestThreadSafety(t *testing.T) {
	ff := GetFeatureFlags()

	// Test concurrent reads and writes
	done := make(chan bool)
	iterations := 100

	// Concurrent readers
	for i := 0; i < 10; i++ {
		go func() {
			for j := 0; j < iterations; j++ {
				_ = ff.IsAILessonsEnabled()
				_ = ff.GetAIProvider()
				_ = ff.GetAIModel()
				_ = ff.GetMaxAILessonsPerDay()
			}
			done <- true
		}()
	}

	// Concurrent writers
	for i := 0; i < 5; i++ {
		go func(id int) {
			for j := 0; j < iterations; j++ {
				ff.SetAILessonsEnabled(id%2 == 0)
			}
			done <- true
		}(i)
	}

	// Wait for all goroutines
	for i := 0; i < 15; i++ {
		<-done
	}

	// If we get here without race detector errors, the test passes
}
