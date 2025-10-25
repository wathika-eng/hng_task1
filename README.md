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

### Using Postgres via .env

You can enable Postgres-backed persistence by providing a `DATABASE_URL` in a `.env` file at the project root. Example `.env`:

```env
DATABASE_URL="postgres://user:pass@localhost:5432/stringsdb"
```

The app will attempt to connect and auto-migrate the `Value` model. If the database does not exist, either create it beforehand or the app will fall back to the in-memory store. To create the database using `psql`:

```bash
PGPASSWORD="your_db_password" psql "postgres://user:pass@localhost:5432/postgres" -c "CREATE DATABASE stringsdb;"
```

Then start the server with the `.env` file in place:

```bash
go run ./
```

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
