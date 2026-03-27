package mock_test

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"regexp"
	"strconv"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"quiz-race/mock"
)

func TestRun_AllRequestsReceived(t *testing.T) {
	const users = 50
	var count int64

	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/metrics" {
			w.WriteHeader(http.StatusOK)
			return
		}
		if r.Method != http.MethodPost {
			t.Errorf("expected POST, got %s", r.Method)
		}
		if r.URL.Path != "/submit" {
			t.Errorf("expected /submit path, got %s", r.URL.Path)
		}
		atomic.AddInt64(&count, 1)
		_ = r.Body.Close()
		w.WriteHeader(http.StatusOK)
	}))
	defer ts.Close()

	mock.Run(users, ts.URL)

	if got := atomic.LoadInt64(&count); got != users {
		t.Fatalf("expected %d requests, got %d", users, got)
	}
	if got := mock.TotalSent(); got != users {
		t.Fatalf("expected total sent %d, got %d", users, got)
	}
	if got := mock.CorrectSent() + mock.IncorrectSent(); got != users {
		t.Fatalf("expected correct+incorrect %d, got %d", users, got)
	}
}

func TestRun_PayloadShape(t *testing.T) {
	payloadCh := make(chan map[string]string, 1)

	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/metrics" {
			w.WriteHeader(http.StatusOK)
			return
		}
		var payload map[string]string
		if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
			t.Errorf("decode payload failed: %v", err)
		}
		_ = r.Body.Close()
		select {
		case payloadCh <- payload:
		default:
		}
		w.WriteHeader(http.StatusOK)
	}))
	defer ts.Close()

	mock.Run(1, ts.URL)

	var payload map[string]string
	select {
	case payload = <-payloadCh:
	default:
		t.Fatal("expected at least one payload")
	}

	userID := payload["user_id"]
	answer := payload["answer"]

	if userID == "" {
		t.Fatal("expected non-empty user_id")
	}
	if answer != "yes" && answer != "no" {
		t.Fatalf("expected answer yes or no, got %q", answer)
	}
}

func TestRun_UserIDFormat(t *testing.T) {
	const users = 20
	re := regexp.MustCompile(`^user-(\d+)$`)

	var mu sync.Mutex
	seen := make(map[string]struct{}, users)

	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/metrics" {
			w.WriteHeader(http.StatusOK)
			return
		}
		var payload map[string]string
		if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
			t.Errorf("decode payload failed: %v", err)
		}
		_ = r.Body.Close()

		userID := payload["user_id"]
		matches := re.FindStringSubmatch(userID)
		if len(matches) != 2 {
			t.Errorf("user_id does not match pattern user-<N>: %q", userID)
		} else {
			n, err := strconv.Atoi(matches[1])
			if err != nil {
				t.Errorf("invalid user_id suffix %q: %v", matches[1], err)
			} else if n < 1 || n > users {
				t.Errorf("user_id out of expected range 1..%d: %q", users, userID)
			}
		}

		mu.Lock()
		seen[userID] = struct{}{}
		mu.Unlock()

		w.WriteHeader(http.StatusOK)
	}))
	defer ts.Close()

	mock.Run(users, ts.URL)

	mu.Lock()
	defer mu.Unlock()
	if len(seen) != users {
		t.Fatalf("expected %d unique user IDs, got %d", users, len(seen))
	}
}

func TestRun_CompletesWithoutHang(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/metrics" {
			w.WriteHeader(http.StatusOK)
			return
		}
		_ = r.Body.Close()
		w.WriteHeader(http.StatusOK)
	}))
	defer ts.Close()

	done := make(chan struct{})
	go func() {
		mock.Run(10, ts.URL)
		close(done)
	}()

	select {
	case <-done:
	case <-time.After(5 * time.Second):
		t.Fatal("mock.Run did not complete before timeout")
	}
}
