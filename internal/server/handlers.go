package server

import (
	"encoding/json"
	"net/http"
	"strings"

	"mbsecli/internal/feedback"

	"github.com/gorilla/websocket"
)

func (s *Server) handleModel(w http.ResponseWriter, r *http.Request) {
	s.mu.RLock()
	g := s.graph
	s.mu.RUnlock()
	writeJSON(w, http.StatusOK, g)
}

func (s *Server) handleFeedback(w http.ResponseWriter, r *http.Request) {
	s.mu.RLock()
	store := s.feedback
	s.mu.RUnlock()

	switch r.Method {
	case http.MethodGet:
		writeJSON(w, http.StatusOK, feedbackAll(store))

	case http.MethodPost:
		if store == nil {
			http.Error(w, "no model loaded", http.StatusServiceUnavailable)
			return
		}
		var req struct {
			FQN    string `json:"fqn"`
			Author string `json:"author"`
			Text   string `json:"text"`
		}
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		if req.FQN == "" || req.Text == "" {
			http.Error(w, "fqn and text are required", http.StatusBadRequest)
			return
		}
		note, err := store.Add(req.FQN, req.Author, req.Text)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		s.PublishUpdate()
		writeJSON(w, http.StatusCreated, note)

	default:
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
	}
}

// handleFeedbackByID handles PATCH /api/feedback/{id} for status updates,
// e.g. {"status": "resolved"}.
func (s *Server) handleFeedbackByID(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPatch {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}
	id := strings.TrimPrefix(r.URL.Path, "/api/feedback/")
	if id == "" {
		http.Error(w, "missing note id", http.StatusBadRequest)
		return
	}

	s.mu.RLock()
	store := s.feedback
	s.mu.RUnlock()
	if store == nil {
		http.Error(w, "no model loaded", http.StatusServiceUnavailable)
		return
	}

	var req struct {
		Status feedback.Status `json:"status"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	note, err := store.SetStatus(id, req.Status)
	if err != nil {
		http.Error(w, err.Error(), http.StatusNotFound)
		return
	}
	s.PublishUpdate()
	writeJSON(w, http.StatusOK, note)
}

func (s *Server) handleWS(w http.ResponseWriter, r *http.Request) {
	conn, err := s.upgrader.Upgrade(w, r, nil)
	if err != nil {
		return
	}
	defer conn.Close()

	ch := s.hub.add()
	defer s.hub.remove(ch)

	// Send the current snapshot immediately so a newly-opened tab doesn't
	// wait for the next file change to see the model.
	s.mu.RLock()
	initial, _ := json.Marshal(map[string]any{
		"type":     "model-update",
		"graph":    s.graph,
		"feedback": feedbackAll(s.feedback),
	})
	s.mu.RUnlock()
	if err := conn.WriteMessage(websocket.TextMessage, initial); err != nil {
		return
	}

	// Reader goroutine: we don't expect client->server messages beyond pings,
	// but we must read to process control frames and detect disconnects.
	go func() {
		for {
			if _, _, err := conn.ReadMessage(); err != nil {
				s.hub.remove(ch)
				return
			}
		}
	}()

	for msg := range ch {
		if err := conn.WriteMessage(websocket.TextMessage, msg); err != nil {
			return
		}
	}
}

func writeJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(v)
}
