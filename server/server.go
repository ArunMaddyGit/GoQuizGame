package server

import (
	"encoding/json"
	"net/http"

	"quiz-race/engine"
)

// Handler defines the behavior expected by the HTTP server layer.
type Handler interface {
	Start(addr string) error
}

// Server represents the quiz-race server.
type Server struct {
	engine *engine.GameEngine
	mux    *http.ServeMux
}

type submitRequest struct {
	UserID string `json:"user_id"`
	Answer string `json:"answer"`
}

// NewServer creates an API server with registered routes.
func NewServer(e *engine.GameEngine) *Server {
	s := &Server{
		engine: e,
		mux:    http.NewServeMux(),
	}

	s.mux.HandleFunc("/submit", s.submitHandler)
	s.mux.HandleFunc("/metrics", s.metricsHandler)
	return s
}

func (s *Server) submitHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}

	defer r.Body.Close()

	var req submitRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid payload"})
		return
	}

	s.engine.Submit(req.UserID, req.Answer)
	writeJSON(w, http.StatusOK, map[string]string{"status": "received"})
}

func (s *Server) metricsHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}

	writeJSONAny(w, http.StatusOK, s.engine.Metrics())
}

func writeJSON(w http.ResponseWriter, status int, payload map[string]string) {
	writeJSONAny(w, status, payload)
}

func writeJSONAny(w http.ResponseWriter, status int, payload interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(payload)
}

// ServeHTTP allows Server to be used directly with httptest.
func (s *Server) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	s.mux.ServeHTTP(w, r)
}

// Start begins listening for HTTP requests on the provided address.
func (s *Server) Start(addr string) error {
	return http.ListenAndServe(addr, s.mux)
}
