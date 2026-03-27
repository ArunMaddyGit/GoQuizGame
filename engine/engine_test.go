package engine_test

import (
	"fmt"
	"sync"
	"testing"

	"quiz-race/engine"
)

func TestNoWinner_WrongAnswer(t *testing.T) {
	e := engine.NewGameEngine()

	for i := 1; i <= 10; i++ {
		e.Submit(fmt.Sprintf("user-%d", i), "no")
	}

	if got := e.Winner(); got != "" {
		t.Fatalf("expected no winner, got %q", got)
	}
}

func TestSingleWinner_FirstYesWins(t *testing.T) {
	e := engine.NewGameEngine()

	e.Submit("user-1", "yes")
	e.Submit("user-2", "yes")

	if got := e.Winner(); got != "user-1" {
		t.Fatalf("expected winner user-1, got %q", got)
	}
}

func TestWinner_YesBeatsNo(t *testing.T) {
	e := engine.NewGameEngine()

	e.Submit("user-1", "no")
	e.Submit("user-2", "yes")

	if got := e.Winner(); got != "user-2" {
		t.Fatalf("expected winner user-2, got %q", got)
	}
}

func TestConcurrentSubmissions_ExactlyOneWinner(t *testing.T) {
	e := engine.NewGameEngine()

	const users = 1000
	var wg sync.WaitGroup
	wg.Add(users)

	for i := 1; i <= users; i++ {
		userID := fmt.Sprintf("user-%d", i)
		go func(id string) {
			defer wg.Done()
			e.Submit(id, "yes")
		}(userID)
	}

	wg.Wait()

	if got := e.Winner(); got == "" {
		t.Fatal("expected exactly one winner, got empty winner")
	}
}

func TestWinnerImmutable_AfterSet(t *testing.T) {
	e := engine.NewGameEngine()

	e.Submit("user-A", "yes")
	firstWinner := e.Winner()
	e.Submit("user-B", "yes")

	if got := e.Winner(); got != firstWinner {
		t.Fatalf("expected winner %q to remain unchanged, got %q", firstWinner, got)
	}
	if firstWinner != "user-A" {
		t.Fatalf("expected initial winner user-A, got %q", firstWinner)
	}
}

func TestAnswerMetrics_Counts(t *testing.T) {
	e := engine.NewGameEngine()

	e.Submit("user-1", "no")
	e.Submit("user-2", "yes")
	e.Submit("user-3", "maybe")

	if got := e.TotalReceived(); got != 3 {
		t.Fatalf("expected total received 3, got %d", got)
	}
	if got := e.CorrectReceived(); got != 1 {
		t.Fatalf("expected correct received 1, got %d", got)
	}
	if got := e.IncorrectReceived(); got != 2 {
		t.Fatalf("expected incorrect received 2, got %d", got)
	}
}
