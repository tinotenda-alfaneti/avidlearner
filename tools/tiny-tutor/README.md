# Tiny CLI Tutor

Minimal terminal-first tutor for AvidLearner lessons.

Usage (from repo root):

```powershell
pushd .\tools\tiny-tutor; go build -o tiny-tutor.exe .; .\tiny-tutor.exe -file ..\..\data\lessons.json
```

Commands inside the CLI:
- `list` — show lessons
- `start <n>` — start lesson number n
- `filter <category>` — restrict lessons by category
- `hint`, `explain`, `quiz` — available while a lesson is open
- `exit` — quit

This is intentionally small: it demonstrates a terminal-first flow and can be extended to store progress, fetch remote lessons, or add auto-graded exercises.
