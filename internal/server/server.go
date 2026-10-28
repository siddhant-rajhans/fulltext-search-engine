package server

import (
	"encoding/json"
	"log/slog"
	"net/http"
	"strconv"
	"time"

	"github.com/siddhant-rajhans/fulltext-search-engine/internal/index"
)

type Server struct {
	idx *index.InvertedIndex
	mux *http.ServeMux
}

func New(idx *index.InvertedIndex) *Server {
	s := &Server{idx: idx, mux: http.NewServeMux()}
	s.mux.HandleFunc("POST /documents", s.handleIndex)
	s.mux.HandleFunc("GET /search", s.handleSearch)
	s.mux.HandleFunc("GET /documents/{id}", s.handleGet)
	s.mux.HandleFunc("GET /stats", s.handleStats)
	return s
}

func (s *Server) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	s.mux.ServeHTTP(w, r)
}

type IndexRequest struct {
	ID      int    `json:"id"`
	Content string `json:"content"`
}

type IndexResponse struct {
	ID      int `json:"id"`
	Indexed bool `json:"indexed"`
}

func (s *Server) handleIndex(w http.ResponseWriter, r *http.Request) {
	var req IndexRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, `{"error":"invalid JSON"}`, http.StatusBadRequest)
		return
	}
	if req.Content == "" {
		http.Error(w, `{"error":"content required"}`, http.StatusBadRequest)
		return
	}

	s.idx.Add(req.ID, req.Content)
	writeJSON(w, http.StatusCreated, IndexResponse{ID: req.ID, Indexed: true})
}

type SearchResponse struct {
	Query   string         `json:"query"`
	Results []ResultItem   `json:"results"`
	Took    string         `json:"took"`
}

type ResultItem struct {
	DocID   int     `json:"doc_id"`
	Score   float64 `json:"score"`
	Snippet string  `json:"snippet,omitempty"`
}

func (s *Server) handleSearch(w http.ResponseWriter, r *http.Request) {
	query := r.URL.Query().Get("q")
	if query == "" {
		http.Error(w, `{"error":"q parameter required"}`, http.StatusBadRequest)
		return
	}

	topK := 10
	if k := r.URL.Query().Get("k"); k != "" {
		if n, err := strconv.Atoi(k); err == nil && n > 0 {
			topK = n
		}
	}

	start := time.Now()
	results := s.idx.Search(query, topK)
	took := time.Since(start)

	items := make([]ResultItem, len(results))
	for i, res := range results {
		item := ResultItem{DocID: res.DocID, Score: res.Score}
		if doc, ok := s.idx.Get(res.DocID); ok {
			item.Snippet = truncate(doc.Content, 200)
		}
		items[i] = item
	}

	writeJSON(w, http.StatusOK, SearchResponse{
		Query:   query,
		Results: items,
		Took:    took.String(),
	})

	slog.Info("search", "query", query, "results", len(items), "took", took)
}

type DocResponse struct {
	ID      int    `json:"id"`
	Content string `json:"content"`
	Length  int    `json:"length"`
}

func (s *Server) handleGet(w http.ResponseWriter, r *http.Request) {
	idStr := r.PathValue("id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		http.Error(w, `{"error":"invalid id"}`, http.StatusBadRequest)
		return
	}

	doc, ok := s.idx.Get(id)
	if !ok {
		http.Error(w, `{"error":"not found"}`, http.StatusNotFound)
		return
	}

	writeJSON(w, http.StatusOK, DocResponse{ID: doc.ID, Content: doc.Content, Length: doc.Length})
}

type StatsResponse struct {
	Documents int `json:"documents"`
	Terms     int `json:"terms"`
}

func (s *Server) handleStats(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, http.StatusOK, StatsResponse{
		Documents: s.idx.DocCount(),
		Terms:     s.idx.TermCount(),
	})
}

func writeJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(v)
}

func truncate(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}
	return s[:maxLen] + "..."
}
