package server_test

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync"
	"testing"
	"time"

	"quiz-race/engine"
	"quiz-race/server"
)

func TestSubmit_ValidYes(t *testing.T) {
	e := engine.NewGameEngine()
	s := server.NewServer(e)
	ts := httptest.NewServer(s)
	defer ts.Close()

	resp, err := http.Post(ts.URL+"/submit", "application/json", strings.NewReader(`{"user_id":"user-1","answer":"yes"}`))
	if err != nil {
		t.Fatalf("post submit failed: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Fatalf("expected status %d, got %d", http.StatusOK, resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		t.Fatalf("read response failed: %v", err)
	}

	if !strings.Contains(string(body), "received") {
		t.Fatalf("expected response to contain received, got %q", string(body))
	}

	if got := e.Winner(); got != "user-1" {
		t.Fatalf("expected winner user-1, got %q", got)
	}
}

func TestSubmit_ValidNo(t *testing.T) {
	e := engine.NewGameEngine()
	s := server.NewServer(e)
	ts := httptest.NewServer(s)
	defer ts.Close()

	resp, err := http.Post(ts.URL+"/submit", "application/json", strings.NewReader(`{"user_id":"user-1","answer":"no"}`))
	if err != nil {
		t.Fatalf("post submit failed: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Fatalf("expected status %d, got %d", http.StatusOK, resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		t.Fatalf("read response failed: %v", err)
	}

	if !strings.Contains(string(body), "received") {
		t.Fatalf("expected response to contain received, got %q", string(body))
	}

	if got := e.Winner(); got != "" {
		t.Fatalf("expected no winner, got %q", got)
	}
}

func TestSubmit_InvalidPayload(t *testing.T) {
	e := engine.NewGameEngine()
	s := server.NewServer(e)
	ts := httptest.NewServer(s)
	defer ts.Close()

	resp, err := http.Post(ts.URL+"/submit", "application/json", strings.NewReader("not-json"))
	if err != nil {
		t.Fatalf("post submit failed: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusBadRequest {
		t.Fatalf("expected status %d, got %d", http.StatusBadRequest, resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		t.Fatalf("read response failed: %v", err)
	}

	if !strings.Contains(string(body), "invalid payload") {
		t.Fatalf("expected response to contain invalid payload, got %q", string(body))
	}
}

func TestSubmit_EmptyBody(t *testing.T) {
	e := engine.NewGameEngine()
	s := server.NewServer(e)
	ts := httptest.NewServer(s)
	defer ts.Close()

	resp, err := http.Post(ts.URL+"/submit", "application/json", strings.NewReader(""))
	if err != nil {
		t.Fatalf("post submit failed: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusBadRequest {
		t.Fatalf("expected status %d, got %d", http.StatusBadRequest, resp.StatusCode)
	}
}

func TestConcurrentRequests_NoRace(t *testing.T) {
	e := engine.NewGameEngine()
	s := server.NewServer(e)
	ts := httptest.NewServer(s)
	defer ts.Close()

	const requests = 500
	var wg sync.WaitGroup
	wg.Add(requests)
	errCh := make(chan error, requests)

	client := &http.Client{}
	for i := 1; i <= requests; i++ {
		userID := fmt.Sprintf("user-%d", i)
		answer := "no"
		if i%2 == 0 {
			answer = "yes"
		}

		go func(id string, ans string) {
			defer wg.Done()
			body := fmt.Sprintf(`{"user_id":"%s","answer":"%s"}`, id, ans)
			var lastErr error
			for attempt := 0; attempt < 30; attempt++ {
				resp, err := client.Post(ts.URL+"/submit", "application/json", strings.NewReader(body))
				if err == nil {
					_ = resp.Body.Close()
					return
				}
				lastErr = err
				time.Sleep(20 * time.Millisecond)
			}
			errCh <- fmt.Errorf("post submit failed for %s: %w", id, lastErr)
		}(userID, answer)
	}

	wg.Wait()
	close(errCh)

	for err := range errCh {
		t.Error(err)
	}

	if got := e.Winner(); got == "" {
		t.Fatal("expected exactly one winner to be set, got empty winner")
	}
}

func TestMetrics_ReturnsCounters(t *testing.T) {
	e := engine.NewGameEngine()
	s := server.NewServer(e)
	ts := httptest.NewServer(s)
	defer ts.Close()

	_, _ = http.Post(ts.URL+"/submit", "application/json", strings.NewReader(`{"user_id":"user-1","answer":"yes"}`))
	_, _ = http.Post(ts.URL+"/submit", "application/json", strings.NewReader(`{"user_id":"user-2","answer":"no"}`))
	_, _ = http.Post(ts.URL+"/submit", "application/json", strings.NewReader(`{"user_id":"user-3","answer":"no"}`))

	resp, err := http.Get(ts.URL + "/metrics")
	if err != nil {
		t.Fatalf("get metrics failed: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Fatalf("expected status %d, got %d", http.StatusOK, resp.StatusCode)
	}

	var m struct {
		Total     uint64 `json:"total"`
		Correct   uint64 `json:"correct"`
		Incorrect uint64 `json:"incorrect"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&m); err != nil {
		t.Fatalf("decode metrics failed: %v", err)
	}

	if m.Total != 3 || m.Correct != 1 || m.Incorrect != 2 {
		t.Fatalf("unexpected metrics: %+v", m)
	}
}
