package mock

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"math/rand"
	"net/http"
	"os"
	"sync"
	"sync/atomic"
	"time"
)

// EngineMock is a placeholder mock for engine behaviors in tests.
type EngineMock struct {
	// TODO: add configurable mock behavior.
}

// ServerMock is a placeholder mock for server-side collaborators in tests.
type ServerMock struct {
	// TODO: add configurable mock behavior.
}

var seedOnce sync.Once
var sentTotalCount atomic.Uint64
var sentCorrectCount atomic.Uint64
var sentIncorrectCount atomic.Uint64

// Run simulates n users submitting answers to the API server.
func Run(n int, serverURL string) {
	seedOnce.Do(func() {
		rand.Seed(time.Now().UnixNano())
	})

	var wg sync.WaitGroup
	sentTotalCount.Store(0)
	sentCorrectCount.Store(0)
	sentIncorrectCount.Store(0)
	wg.Add(n)

	for i := 1; i <= n; i++ {
		userID := fmt.Sprintf("user-%d", i)

		go func(id string) {
			defer wg.Done()

			answer := "no"
			if rand.Intn(2) == 0 {
				answer = "yes"
			}

			delayMs := rand.Intn(991) + 10
			time.Sleep(time.Duration(delayMs) * time.Millisecond)

			payload, err := json.Marshal(map[string]string{
				"user_id": id,
				"answer":  answer,
			})
			if err != nil {
				fmt.Fprintf(os.Stderr, "request marshal error for %s: %v\n", id, err)
				return
			}

			resp, err := http.Post(serverURL+"/submit", "application/json", bytes.NewReader(payload))
			if err != nil {
				fmt.Fprintf(os.Stderr, "request error for %s: %v\n", id, err)
				return
			}
			_ = resp.Body.Close()

			total := sentTotalCount.Add(1)
			if answer == "yes" {
				_ = sentCorrectCount.Add(1)
				// fmt.Printf("Mock : sent correct answers: %d\n", correct)
			} else {
				_ = sentIncorrectCount.Add(1)
				// fmt.Printf("Mock sent incorrect answers: %d\n", incorrect)
			}
			if total == uint64(n) {
				fmt.Printf("Mock sent answers: %d\n", total)
			}
		}(userID)
	}

	wg.Wait()
	fmt.Printf("Mock : total number of answers sent: %d\n", sentTotalCount.Load())
	fmt.Printf("Mock : correct answers sent: %d\n", sentCorrectCount.Load())
	fmt.Printf("Mock : incorrect answers sent: %d\n", sentIncorrectCount.Load())

	metricsResp, err := http.Get(serverURL + "/metrics")
	if err != nil {
		fmt.Fprintf(os.Stderr, "metrics request error: %v\n", err)
	} else {
		body, readErr := io.ReadAll(metricsResp.Body)
		_ = metricsResp.Body.Close()
		if readErr != nil {
			fmt.Fprintf(os.Stderr, "metrics read error: %v\n", readErr)
		} else {
			fmt.Printf("Final Metrics response from engine.go: %s\n", string(body))
		}
	}
	fmt.Println("All users responded.")
}

// TotalSent returns the total number of successful submissions sent by mock users.
func TotalSent() uint64 {
	return sentTotalCount.Load()
}

// CorrectSent returns the number of successful "yes" submissions sent by mock users.
func CorrectSent() uint64 {
	return sentCorrectCount.Load()
}

// IncorrectSent returns the number of successful "no" submissions sent by mock users.
func IncorrectSent() uint64 {
	return sentIncorrectCount.Load()
}
