# Autograder CLI

Simple Go-based autograder for coding challenges. It runs tests in a temporary directory and provides heuristic, step-by-step explanations for common compiler/runtime/test failures.

Build:

```powershell
Set-Location tools\autograder
go build -o autograder.exe .
```

Usage examples:

List challenges:

```powershell
.\autograder.exe -file ..\..\data\challenges.json -list
```

Attempt a challenge (supply a Go source file implementing required functions):

```powershell
.\autograder.exe -file ..\..\data\challenges.json -id sum-ints -code mycode.go
```

Or use the interactive menu: run without flags and follow prompts.
