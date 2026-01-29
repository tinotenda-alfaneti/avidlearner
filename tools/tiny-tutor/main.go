package main

import (
	"bufio"
	"encoding/json"
	"flag"
	"fmt"
	"math/rand"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"
)

type Lesson struct {
	Title    string   `json:"title"`
	Category string   `json:"category"`
	Text     string   `json:"text"`
	Explain  string   `json:"explain"`
	UseCases []string `json:"useCases"`
	Tips     []string `json:"tips"`
}

type ProgressEntry struct {
	Completed     bool      `json:"completed"`
	Attempts      int       `json:"attempts"`
	LastCompleted time.Time `json:"lastCompleted,omitempty"`
}

var (
	progress     map[string]ProgressEntry
	progressPath string
)

func loadLessons(path string) ([]Lesson, error) {
	b, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	var lessons []Lesson
	if err := json.Unmarshal(b, &lessons); err != nil {
		return nil, err
	}
	return lessons, nil
}

func listLessons(lessons []Lesson) {
	for i, l := range lessons {
		fmt.Printf("%3d) %s [%s]\n", i+1, l.Title, l.Category)
	}
}

func prompt(s *bufio.Reader, label string) (string, error) {
	fmt.Print(label)
	line, err := s.ReadString('\n')
	if err != nil {
		return "", err
	}
	return strings.TrimSpace(line), nil
}

func showLessonInteractive(lesson Lesson, all []Lesson) error {
	fmt.Printf("\n== %s ==\n\n", lesson.Title)
	fmt.Println(lesson.Text)
	fmt.Println()
	r := bufio.NewReader(os.Stdin)
	for {
		fmt.Println("Options: (h)int  (e)xplain  (q)uiz  (m)enu  (x)exit")
		ans, err := prompt(r, "> ")
		if err != nil {
			return err
		}
		switch strings.ToLower(ans) {
		case "h", "hint":
			if len(lesson.Tips) > 0 {
				fmt.Println("Hint:", lesson.Tips[0])
			} else {
				fmt.Println("No tips available for this lesson.")
			}
		case "e", "explain":
			fmt.Println("Explanation:")
			fmt.Println(lesson.Explain)
		case "q", "quiz":
			runQuickQuiz(lesson, all)
		case "m", "menu":
			return nil
		case "x", "exit":
			os.Exit(0)
		default:
			fmt.Println("Unknown option")
		}
		fmt.Println()
	}
}

func runQuickQuiz(lesson Lesson, all []Lesson) {
	// Simple quiz: pick a use case and mix with other random use cases
	if len(lesson.UseCases) == 0 {
		fmt.Println("No quiz available for this lesson.")
		return
	}
	correct := lesson.UseCases[0]
	// Collect distractors
	pool := []string{correct}
	for _, l := range all {
		if l.Title == lesson.Title {
			continue
		}
		for _, uc := range l.UseCases {
			if len(pool) >= 4 {
				break
			}
			if uc != correct {
				pool = append(pool, uc)
			}
		}
		if len(pool) >= 4 {
			break
		}
	}
	// Ensure at least 2 options
	if len(pool) < 2 {
		fmt.Println("Not enough data to generate a quiz for this lesson.")
		return
	}
	rand.Shuffle(len(pool), func(i, j int) { pool[i], pool[j] = pool[j], pool[i] })
	fmt.Println("Which of the following is a primary use case for this lesson?")
	for i, opt := range pool {
		fmt.Printf("%d) %s\n", i+1, opt)
	}
	fmt.Print("Answer (number): ")
	var choice int
	_, err := fmt.Scanf("%d\n", &choice)
	if err != nil || choice < 1 || choice > len(pool) {
		fmt.Println("Invalid choice.")
		return
	}
	picked := pool[choice-1]
	if picked == correct {
		fmt.Println("Correct! —", correct)
		recordAttempt(lesson.Title, true)
	} else {
		fmt.Println("Not quite.")
		recordAttempt(lesson.Title, false)
		fmt.Println("Expected:", correct)
		fmt.Println("Why: ")
		fmt.Println(lesson.Explain)
		if len(lesson.Tips) > 0 {
			fmt.Println("Tips:")
			for _, t := range lesson.Tips {
				fmt.Println(" -", t)
			}
		}
	}
}

func loadProgress(path string) (map[string]ProgressEntry, error) {
	m := make(map[string]ProgressEntry)
	b, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return m, nil
		}
		return nil, err
	}
	if err := json.Unmarshal(b, &m); err != nil {
		return nil, err
	}
	return m, nil
}

func saveProgress(path string, m map[string]ProgressEntry) error {
	if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
		return err
	}
	b, err := json.MarshalIndent(m, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(path, b, 0644)
}

func recordAttempt(title string, correct bool) {
	if progress == nil {
		progress = make(map[string]ProgressEntry)
	}
	p := progress[title]
	p.Attempts++
	if correct {
		p.Completed = true
		p.LastCompleted = time.Now()
	}
	progress[title] = p
	if progressPath != "" {
		_ = saveProgress(progressPath, progress)
	}
}

func printProgress() {
	if progress == nil || len(progress) == 0 {
		fmt.Println("No progress recorded yet.")
		return
	}
	for title, p := range progress {
		var s string
		if p.Completed {
			s = "completed"
		} else {
			s = "incomplete"
		}
		fmt.Printf("- %s: %s (attempts=%d", title, s, p.Attempts)
		if !p.LastCompleted.IsZero() {
			fmt.Printf(", last=%s", p.LastCompleted.Format(time.RFC3339))
		}
		fmt.Println(")")
	}
}

func main() {
	rand.Seed(time.Now().UnixNano())
	listFlag := flag.Bool("list", false, "List available lessons")
	lessonFlag := flag.Int("lesson", 0, "Start a lesson by number (1-based)")
	catFlag := flag.String("category", "", "Filter lessons by category")
	fileFlag := flag.String("file", filepath.Join("..", "data", "lessons.json"), "Path to lessons JSON")
	progressFlag := flag.String("progress", filepath.Join("..", "data", "tutor_progress.json"), "Path to progress JSON")
	flag.Parse()

	// Resolve path relative to executable if necessary
	exePath, _ := os.Executable()
	exeDir := filepath.Dir(exePath)
	lessonsPath := *fileFlag
	if !filepath.IsAbs(lessonsPath) {
		// try repo-relative: assume running from repo root
		lessonsPath = filepath.Clean(filepath.Join(exeDir, lessonsPath))
	}

	lessons, err := loadLessons(lessonsPath)
	if err != nil {
		// fallback to repo relative path
		lessonsPath = filepath.Join("..", "data", "lessons.json")
		lessons, err = loadLessons(lessonsPath)
		if err != nil {
			fmt.Fprintln(os.Stderr, "failed to load lessons:", err)
			os.Exit(1)
		}
	}

	// optional category filter
	var filtered []Lesson
	if *catFlag != "" {
		for _, l := range lessons {
			if strings.EqualFold(l.Category, *catFlag) {
				filtered = append(filtered, l)
			}
		}
	} else {
		filtered = lessons
	}

	if *listFlag {
		listLessons(filtered)
		return
	}

	// resolve and load progress
	progressPath = *progressFlag
	if !filepath.IsAbs(progressPath) {
		progressPath = filepath.Clean(filepath.Join(exeDir, progressPath))
	}
	prog, err := loadProgress(progressPath)
	if err != nil {
		// fallback to repo-relative
		progressPath = filepath.Join("..", "data", "tutor_progress.json")
		prog, _ = loadProgress(progressPath)
	}
	progress = prog

	if *lessonFlag > 0 {
		idx := *lessonFlag - 1
		if idx < 0 || idx >= len(filtered) {
			fmt.Fprintln(os.Stderr, "lesson number out of range")
			os.Exit(2)
		}
		if err := showLessonInteractive(filtered[idx], filtered); err != nil {
			fmt.Fprintln(os.Stderr, "error:", err)
		}
		return
	}

	// interactive menu
	r := bufio.NewReader(os.Stdin)
	for {
		fmt.Println("\nAvidLearner Tiny CLI Tutor")
		fmt.Println("Commands: list  start <n>  filter <category>  progress  reset-progress  exit")
		cmd, _ := prompt(r, "> ")
		parts := strings.Fields(cmd)
		if len(parts) == 0 {
			continue
		}
		switch parts[0] {
		case "list":
			listLessons(filtered)
		case "progress":
			printProgress()
		case "reset-progress":
			progress = make(map[string]ProgressEntry)
			if progressPath != "" {
				_ = saveProgress(progressPath, progress)
			}
			fmt.Println("progress reset")
		case "start":
			if len(parts) < 2 {
				fmt.Println("usage: start <number>")
				continue
			}
			n, err := strconv.Atoi(parts[1])
			if err != nil {
				fmt.Println("invalid number")
				continue
			}
			if n < 1 || n > len(filtered) {
				fmt.Println("number out of range")
				continue
			}
			_ = showLessonInteractive(filtered[n-1], filtered)
		case "filter":
			if len(parts) < 2 {
				filtered = lessons
				fmt.Println("cleared filter")
				continue
			}
			cat := parts[1]
			filtered = nil
			for _, l := range lessons {
				if strings.EqualFold(l.Category, cat) {
					filtered = append(filtered, l)
				}
			}
			fmt.Printf("filtered to %d lessons in '%s'\n", len(filtered), cat)
		case "exit", "quit":
			os.Exit(0)
		default:
			fmt.Println("unknown command")
		}
	}
}
