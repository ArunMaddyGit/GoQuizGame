# Quiz Race

Quiz Race is a concurrent quiz simulation in Go where N users submit answers at nearly the same time, and the system declares the first correct submission as the winner. It is designed to demonstrate safe concurrent winner selection under load using Go's standard library primitives and HTTP stack.

## Architecture

- **Mock User Engine (`mock/mock.go`)**: Simulates N concurrent users, generates randomized answers, and submits them to the API.
- **API Server (`server/server.go`)**: Exposes `POST /submit`, validates and forwards submissions to the game engine.
- **Game Engine (`engine/engine.go`)**: Decides the winner once and only once under concurrent submissions.

```text
[Mock Users] --HTTP POST /submit--> [API Server] --> [Game Engine] --> prints Winner
```

## Project Structure

```text
quiz-race/
├── main.go
├── go.mod
├── mock/
│   ├── mock.go
│   └── mock_test.go
├── server/
│   ├── server.go
│   └── server_test.go
└── engine/
    ├── engine.go
    └── engine_test.go
```

## Prerequisites

- Go 1.21 or higher

Verify installation:

```bash
go version
```

## Installation

Step 1: Clone the repository

```bash
git clone https://github.com/your-username/quiz-race.git
cd quiz-race
```

Step 2: Download dependencies

```bash
go mod tidy
```

## Running the Application

Default run (1000 users):

```bash
go run .
```

Custom number of users (example: 500):

```bash
go run . -n 500
```

Expected console output:

```text
Server started on :8080
Winner: user-342
All users responded.
```

## Running the Tests

Run all tests:

```bash
go test ./...
```

Run with race detector:

```bash
go test -race ./...
```

Run specific packages:

```bash
go test ./engine/...
go test ./server/...
go test ./mock/...
```

Run verbose with race detector:

```bash
go test -v -race ./...
```

Expected result:

```text
PASS
# no DATA RACE warnings
```

## API Reference

- **Endpoint**: `POST /submit`
- **Request headers**: `Content-Type: application/json`

Request body schema:

```json
{
  "user_id": "string",
  "answer": "yes|no"
}
```

- `user_id`: user identifier (example: `"user-42"`)
- `answer`: `"yes"` or `"no"` (case-insensitive)

Response `200`:

```json
{ "status": "received" }
```

Response `400` (bad JSON):

```json
{ "error": "invalid payload" }
```

Response `400` (bad answer value):

```json
{ "error": "invalid answer value" }
```

Response `405` (wrong HTTP method).

## Concurrency Design

`sync.Once` in the Game Engine guarantees that winner assignment happens exactly once, even if 1000 correct (`"yes"`) submissions arrive concurrently. The first goroutine to execute the `Once` block sets and prints the winner, and all later attempts are ignored.

The API server uses Go's built-in `net/http`, which handles each request in its own goroutine. Because winner state ownership is centralized in the Game Engine, no additional locking is needed in the server layer.

## Known Limitations & Future Enhancements

- Currently ignores all submissions after the first correct answer (no counter).
- Future: track count of correct answers received after winner is declared.
- Future: support multiple rounds.
- Future: persist results to a database.

## License

MIT
