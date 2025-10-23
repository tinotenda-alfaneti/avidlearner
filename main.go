package main

import (
	"log"
	"os"
	"os/exec"
	"path/filepath"
)

// Runner executable: prefer compiled backend binary (backend/backend.exe or backend/backend).
// If not present, fall back to `go run ./backend`.
func main() {
	// prefer a compiled backend binary
	exe := filepath.Join("backend", "backend.exe")
	if _, err := os.Stat(exe); os.IsNotExist(err) {
		// try without .exe for Unix builds
		exe = filepath.Join("backend", "backend")
	}

	if _, err := os.Stat(exe); err == nil {
		cmd := exec.Command(exe, os.Args[1:]...)
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
		cmd.Stdin = os.Stdin
		cmd.Env = os.Environ()
		if err := cmd.Run(); err != nil {
			log.Fatalf("failed to run backend binary %s: %v", exe, err)
		}
		return
	}

	// no compiled binary found; run go run ./backend
	cmd := exec.Command("go", "run", "./backend")
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Stdin = os.Stdin
	cmd.Env = os.Environ()
	if err := cmd.Run(); err != nil {
		log.Fatalf("failed to run 'go run ./backend': %v", err)
	}
}
