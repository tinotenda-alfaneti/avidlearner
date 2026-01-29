package main

import (
	"bytes"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strings"
	"time"
)

type Challenge struct {
	ID          string `json:"id"`
	Title       string `json:"title"`
	Description string `json:"description"`
	StarterCode string `json:"starterCode"`
	TestCode    string `json:"testCode"`
}

func loadChallenges(path string) ([]Challenge, error) {
	b, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	var cs []Challenge
	if err := json.Unmarshal(b, &cs); err != nil {
		return nil, err
	}
	return cs, nil
}

func listChallenges(cs []Challenge) {
	for i, c := range cs {
		fmt.Printf("%2d) %s — %s\n", i+1, c.Title, c.ID)
	}
}

func explainOutput(out string) string {
	// Simple heuristics to provide step-by-step explanations.
	var b strings.Builder
	if strings.Contains(out, "undefined: ") {
		re := regexp.MustCompile(`undefined: ([^\n]+)`)
		m := re.FindStringSubmatch(out)
		if len(m) > 1 {
			fmt.Fprintf(&b, "Error: `%s` is undefined.\n", m[1])
			fmt.Fprintln(&b, "Step: Did you declare the function/variable with the exact name and package scope?")
		}
	}
	if strings.Contains(out, "cannot use") && strings.Contains(out, "as type") {
		fmt.Fprintln(&b, "Type mismatch: a value was used where a different type was expected.")
		fmt.Fprintln(&b, "Step: Check function signatures, and ensure return/parameter types match the test harness.")
	}
	if strings.Contains(out, "panic:") {
		fmt.Fprintln(&b, "Runtime panic occurred.")
		// try to extract file:line
		re := regexp.MustCompile(`(?m)^\s*([^:\n]+:\d+):`)
		if m := re.FindStringSubmatch(out); len(m) > 1 {
			fmt.Fprintf(&b, "Location: %s\n", m[1])
		}
		fmt.Fprintln(&b, "Step: Inspect the indicated line for nil dereferences, index errors, or out-of-bounds access.")
	}
	if strings.Contains(out, "FAIL") && strings.Contains(out, "Test") {
		fmt.Fprintln(&b, "Test assertions failed.")
		// show lines from test failure
		lines := strings.Split(out, "\n")
		for _, l := range lines {
			if strings.Contains(l, "--- FAIL:") || strings.Contains(l, "Error: ") || strings.Contains(l, "got=") || strings.Contains(l, "want=") {
				fmt.Fprintln(&b, l)
			}
		}
		fmt.Fprintln(&b, "Step: Compare expected vs actual values in the failing assertion and add printing/logging to reproduce locally.")
	}
	if b.Len() == 0 {
		fmt.Fprintln(&b, "Output:\n"+out)
	}
	return b.String()
}

func runChallenge(ch Challenge, codeFile string) (string, error) {
	tmp, err := os.MkdirTemp("", "autograder-")
	if err != nil {
		return "", err
	}
	defer os.RemoveAll(tmp)

	// write user code
	userPath := filepath.Join(tmp, "user.go")
	b, err := os.ReadFile(codeFile)
	if err != nil {
		return "", err
	}
	if err := os.WriteFile(userPath, b, 0644); err != nil {
		return "", err
	}

	// write test file
	testPath := filepath.Join(tmp, "challenge_test.go")
	if err := os.WriteFile(testPath, []byte(ch.TestCode), 0644); err != nil {
		return "", err
	}

	// run `go test` inside tmp
	cmd := exec.Command("go", "test", "-v")
	cmd.Dir = tmp
	var outb bytes.Buffer
	cmd.Stdout = &outb
	cmd.Stderr = &outb

	// set a timeout
	done := make(chan error)
	go func() { done <- cmd.Run() }()
	select {
	case err := <-done:
		return outb.String(), err
	case <-time.After(8 * time.Second):
		_ = cmd.Process.Kill()
		return outb.String(), errors.New("timeout")
	}
}

func printDetail(out string, err error) {
	if err == nil {
		fmt.Println("✅ Passed all tests")
		fmt.Println(out)
		return
	}
	fmt.Println("❌ Tests failed or errors occurred")
	fmt.Println("--- Raw Output ---")
	fmt.Println(out)
	fmt.Println("--- Explanation ---")
	fmt.Println(explainOutput(out))
	if err != nil {
		if exitErr, ok := err.(*exec.ExitError); ok {
			_ = exitErr
			// already included in out
		} else {
			fmt.Println("Runner error:", err)
		}
	}
}

func main() {
	listFlag := flag.Bool("list", false, "List challenges")
	idFlag := flag.String("id", "", "Challenge id to attempt")
	codeFlag := flag.String("code", "", "Path to user code file (required for attempt)")
	fileFlag := flag.String("file", filepath.Join("..", "data", "challenges.json"), "Path to challenges JSON")
	flag.Parse()

	cs, err := loadChallenges(*fileFlag)
	if err != nil {
		// try repo-relative
		cs, err = loadChallenges(filepath.Join("..", "data", "challenges.json"))
		if err != nil {
			fmt.Fprintln(os.Stderr, "failed to load challenges:", err)
			os.Exit(1)
		}
	}

	if *listFlag {
		listChallenges(cs)
		return
	}

	if *idFlag != "" {
		if *codeFlag == "" {
			fmt.Fprintln(os.Stderr, "-code is required for attempts")
			os.Exit(2)
		}
		var found *Challenge
		for _, c := range cs {
			if c.ID == *idFlag {
				found = &c
				break
			}
		}
		if found == nil {
			fmt.Fprintln(os.Stderr, "challenge not found")
			os.Exit(3)
		}
		out, runErr := runChallenge(*found, *codeFlag)
		printDetail(out, runErr)
		return
	}

	// interactive small menu
	fmt.Println("AvidLearner Autograder CLI")
	fmt.Println("Commands: list, attempt <id> <codefile>, exit")
	var cmd string
	for {
		fmt.Print("> ")
		if _, err := fmt.Scan(&cmd); err != nil {
			if err == io.EOF {
				return
			}
			fmt.Fprintln(os.Stderr, "read error:", err)
			return
		}
		switch cmd {
		case "list":
			listChallenges(cs)
		case "attempt":
			var id, file string
			if _, err := fmt.Scan(&id); err != nil {
				fmt.Println("usage: attempt <id> <codefile>")
				continue
			}
			if _, err := fmt.Scan(&file); err != nil {
				fmt.Println("usage: attempt <id> <codefile>")
				continue
			}
			var found *Challenge
			for _, c := range cs {
				if c.ID == id {
					found = &c
					break
				}
			}
			if found == nil {
				fmt.Println("challenge not found")
				continue
			}
			out, runErr := runChallenge(*found, file)
			printDetail(out, runErr)
		case "exit", "quit":
			return
		default:
			fmt.Println("unknown command")
		}
	}
}
