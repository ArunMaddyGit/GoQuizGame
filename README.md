# Quiz Race

Quiz Race is a concurrent quiz simulation in Go where N users submit answers at nearly the same time, and the system declares the first correct submission as the winner. It demonstrates safe concurrent winner selection under load, plus end-to-end metrics for total/correct/incorrect answers in both sender and receiver paths.

## Architecture

- **Mock User Engine (`mock/mock.go`)**: Simulates N concurrent users, generates randomized answers, submits them to the API, tracks sender metrics, and fetches `/metrics` at the end.
- **API Server (`server/server.go`)**: Exposes `POST /submit` and `GET /metrics`, forwards submissions to the game engine, and returns aggregated counters.
- **Game Engine (`engine/engine.go`)**: Decides the winner once under concurrency and tracks receiver metrics (total/correct/incorrect).

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
git clone https://github.com/ArunMaddyGit/GoQuizGame.git
cd GoQuizGame
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
Winner: user-342
Time to winner: 183.214ms
Mock total answers sent: 1000
Mock correct answers sent: 491
Mock incorrect answers sent: 509
Metrics response from engine.go: {"total":1000,"correct":491,"incorrect":509}
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
- **Endpoint**: `GET /metrics`
- **Request headers**: `Content-Type: application/json`

Request body schema:

```json
{
  "user_id": "string",
  "answer": "yes|no"
}
```

- `user_id`: user identifier (example: `"user-42"`)
- `answer`: raw string passed to engine; only exact `"yes"` is treated as correct (case-sensitive)

Response `200`:

```json
{ "status": "received" }
```

Response `400` (bad JSON):

```json
{ "error": "invalid payload" }
```

Response `200` for `GET /metrics`:

```json
{ "total": 1000, "correct": 491, "incorrect": 509 }
```

Response `405` (wrong HTTP method).

## Concurrency Design

`sync.Once` in the Game Engine guarantees that winner assignment happens exactly once, even if 1000 correct (`"yes"`) submissions arrive concurrently. The first goroutine to execute the `Once` block sets and prints the winner and elapsed time, and all later attempts are ignored.

The API server uses Go's built-in `net/http`, which handles each request in its own goroutine. Because winner and metrics state ownership is centralized in the Game Engine (using atomics and `sync.Once`), no additional locking is needed in the server layer.

## Known Limitations & Future Enhancements

- Currently ignores all submissions after the first correct answer for winner selection (metrics still continue).

