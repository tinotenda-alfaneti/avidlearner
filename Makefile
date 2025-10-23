.PHONY: build run

build:
	cd backend && go build -o backend.exe .

run: build
	@echo "To run on Windows PowerShell, use: scripts\\run.ps1 [port]"
