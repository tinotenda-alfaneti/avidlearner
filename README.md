# avidlearner
my own tutor mvp

Run locally
1. Backend (Go):

	- Build and run (default port 8080):

	  Set PORT if you want a different port (e.g. 8081).

2. Frontend:

	- Open `frontend/index.html` in your browser or serve it from the backend (backend serves `../frontend` by default).

Examples (Windows PowerShell):

```powershell
# Build backend
cd .\backend
go build -o backend.exe .
# Run on a different port if needed
$env:PORT = '8081'
Start-Process -NoNewWindow -FilePath "$PWD\backend.exe" -WorkingDirectory "$PWD"
```

Then open http://localhost:8081/ in your browser (or the port you set).
