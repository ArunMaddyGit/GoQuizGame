package engine

import (
	"fmt"
	"sync"
	"sync/atomic"
	"time"
)

// QuizEngine defines the behavior required by the quiz game engine.
type QuizEngine interface {
	Submit(userID string, answer string)
	Winner() string
}

// Metrics represents aggregate answer counters.
type Metrics struct {
	Total     uint64 `json:"total"`
	Correct   uint64 `json:"correct"`
	Incorrect uint64 `json:"incorrect"`
}

// Engine is the default quiz engine implementation.
type Engine struct {
	// TODO: add engine dependencies and state.
}

// GameEngine records exactly one winner across concurrent submissions.
type GameEngine struct {
	once           sync.Once
	winner         string
	startedAt      time.Time
	receivedCount  atomic.Uint64
	correctCount   atomic.Uint64
	incorrectCount atomic.Uint64
}

// NewGameEngine returns a new game engine instance.
func NewGameEngine() *GameEngine {
	return &GameEngine{
		startedAt: time.Now(),
	}
}

// Submit records the first user who answers "yes".
func (g *GameEngine) Submit(userID string, answer string) {
	count := g.receivedCount.Add(1)
	if count == 1000 {
		fmt.Printf("Engine : Total number of received answers: %d\n", count)
	}
	if answer != "yes" {
		_ = g.incorrectCount.Add(1)
		// fmt.Printf("Engine incorrect answers: %d\n", incorrect)
		return
	}
	_ = g.correctCount.Add(1)
	// fmt.Printf("Engine correct answers: %d\n", correct)

	g.once.Do(func() {
		g.winner = userID
		fmt.Printf("Winner: %s\n", userID)
		fmt.Printf("Time to winner: %s\n", time.Since(g.startedAt))
	})
}

// Winner returns the current winner or an empty string.
func (g *GameEngine) Winner() string {
	return g.winner
}

// TotalReceived returns the total number of answers seen by the engine.
func (g *GameEngine) TotalReceived() uint64 {
	return g.receivedCount.Load()
}

// CorrectReceived returns the number of "yes" answers seen by the engine.
func (g *GameEngine) CorrectReceived() uint64 {
	return g.correctCount.Load()
}

// IncorrectReceived returns the number of non-"yes" answers seen by the engine.
func (g *GameEngine) IncorrectReceived() uint64 {
	return g.incorrectCount.Load()
}

// Metrics returns total, correct, and incorrect answer counters.
func (g *GameEngine) Metrics() Metrics {
	return Metrics{
		Total:     g.receivedCount.Load(),
		Correct:   g.correctCount.Load(),
		Incorrect: g.incorrectCount.Load(),
	}
}
