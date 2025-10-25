# String Analyzer API (Stage 1)

Simple REST API that analyzes strings and stores computed properties in-memory.

Features:

- POST /strings — analyze and store a string
- GET /strings/{value} — get analyzed string by original value
- GET /strings — list all strings with query filters
- GET /strings/filter-by-natural-language?query=... — small natural language filter
- DELETE /strings/{value} — delete stored string

## Run locally

Requirements: Go 1.20+ (module is configured)

Install dependencies and run:

```bash
go mod download
go run ./
```

Server listens on :3000

## Example curl

Create:

```bash
curl -X POST -H "Content-Type: application/json" -d '{"value":"racecar"}' http://localhost:3000/strings
```

Get:

```bash
curl http://localhost:3000/strings/racecar
```

List with filters:

```bash
curl "http://localhost:3000/strings?is_palindrome=true&min_length=1"
```

Natural language filter:

```bash
curl "http://localhost:3000/strings/filter-by-natural-language?query=all%20single%20word%20palindromic%20strings"
```

## Notes

- This implementation uses an in-memory store (not persistent) to remain environment-agnostic for the cohort. You can easily swap in a persistent DB by implementing the store methods in `database/storage.go`.
- The natural language parser is heuristic and supports a small set of example phrases described in the task.
