// Package server exposes the parsed model, feedback notes, and live-reload
// events over HTTP + WebSocket, and serves the embedded React build.
package server

import (
	"encoding/json"
	"io/fs"
	"log"
	"net/http"
	"sync"

	"mbsecli/internal/feedback"
	"mbsecli/internal/model"

	"github.com/gorilla/websocket"
)

// Server holds the current model state and serves it to the UI.
type Server struct {
	mu       sync.RWMutex
	graph    *model.Graph
	feedback *feedback.Store

	hub      *hub
	upgrader websocket.Upgrader
	assets   fs.FS // embedded React build (web/dist), or nil in dev mode
}

// New creates a Server. assets may be nil during `npm run dev`, in which
// case the Vite dev server (not this Go server) serves the frontend and only
// the /api and /ws routes are used.
func New(assets fs.FS) *Server {
	return &Server{
		hub:      newHub(),
		upgrader: websocket.Upgrader{CheckOrigin: func(r *http.Request) bool { return true }},
		assets:   assets,
	}
}

// SetGraph updates the current model snapshot and, if changed, notifies the
// hub is left to the caller (see PublishUpdate) so the watcher loop controls
// exactly when a broadcast happens.
func (s *Server) SetGraph(g *model.Graph) {
	s.mu.Lock()
	s.graph = g
	s.mu.Unlock()
}

func (s *Server) SetFeedback(store *feedback.Store) {
	s.mu.Lock()
	s.feedback = store
	s.mu.Unlock()
}

// PublishUpdate broadcasts the current graph + feedback to all connected
// clients. Call this after SetGraph following a re-parse.
func (s *Server) PublishUpdate() {
	s.mu.RLock()
	payload := map[string]any{
		"type":     "model-update",
		"graph":    s.graph,
		"feedback": feedbackAll(s.feedback),
	}
	s.mu.RUnlock()

	data, err := json.Marshal(payload)
	if err != nil {
		log.Printf("server: marshal update: %v", err)
		return
	}
	s.hub.broadcast(data)
}

func feedbackAll(store *feedback.Store) map[string][]*feedback.Note {
	if store == nil {
		return map[string][]*feedback.Note{}
	}
	return store.All()
}

// Routes builds the http.Handler for the whole server.
func (s *Server) Routes() http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("/api/model", s.handleModel)
	mux.HandleFunc("/api/feedback", s.handleFeedback)
	mux.HandleFunc("/api/feedback/", s.handleFeedbackByID)
	mux.HandleFunc("/ws", s.handleWS)

	if s.assets != nil {
		mux.Handle("/", spaHandler(s.assets))
	}
	return withCORS(mux)
}

func withCORS(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PATCH, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type")
		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusNoContent)
			return
		}
		next.ServeHTTP(w, r)
	})
}
