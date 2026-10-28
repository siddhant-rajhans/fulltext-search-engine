package server

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/siddhant-rajhans/fulltext-search-engine/internal/index"
)

func setup(t *testing.T) (*Server, *index.InvertedIndex) {
	t.Helper()
	idx := index.New()
	return New(idx), idx
}

func TestIndexDocument(t *testing.T) {
	srv, _ := setup(t)

	body := `{"id": 1, "content": "hello world search engine"}`
	req := httptest.NewRequest("POST", "/documents", bytes.NewBufferString(body))
	rec := httptest.NewRecorder()
	srv.ServeHTTP(rec, req)

	if rec.Code != http.StatusCreated {
		t.Fatalf("expected 201, got %d: %s", rec.Code, rec.Body.String())
	}

	var resp IndexResponse
	json.NewDecoder(rec.Body).Decode(&resp)
	if !resp.Indexed {
		t.Fatal("expected indexed=true")
	}
}

func TestIndexEmptyContent(t *testing.T) {
	srv, _ := setup(t)

	body := `{"id": 1, "content": ""}`
	req := httptest.NewRequest("POST", "/documents", bytes.NewBufferString(body))
	rec := httptest.NewRecorder()
	srv.ServeHTTP(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Fatalf("expected 400 for empty content, got %d", rec.Code)
	}
}

func TestSearchEndpoint(t *testing.T) {
	srv, idx := setup(t)
	idx.Add(1, "information retrieval and search engines")
	idx.Add(2, "database management systems")

	req := httptest.NewRequest("GET", "/search?q=search+engines", nil)
	rec := httptest.NewRecorder()
	srv.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rec.Code)
	}

	var resp SearchResponse
	json.NewDecoder(rec.Body).Decode(&resp)
	if len(resp.Results) == 0 {
		t.Fatal("expected results")
	}
	if resp.Results[0].DocID != 1 {
		t.Fatalf("expected doc 1, got %d", resp.Results[0].DocID)
	}
}

func TestSearchMissingQuery(t *testing.T) {
	srv, _ := setup(t)

	req := httptest.NewRequest("GET", "/search", nil)
	rec := httptest.NewRecorder()
	srv.ServeHTTP(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", rec.Code)
	}
}

func TestGetDocument(t *testing.T) {
	srv, idx := setup(t)
	idx.Add(42, "test document")

	req := httptest.NewRequest("GET", "/documents/42", nil)
	rec := httptest.NewRecorder()
	srv.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rec.Code)
	}

	var resp DocResponse
	json.NewDecoder(rec.Body).Decode(&resp)
	if resp.ID != 42 {
		t.Fatalf("expected id 42, got %d", resp.ID)
	}
}

func TestGetDocumentNotFound(t *testing.T) {
	srv, _ := setup(t)

	req := httptest.NewRequest("GET", "/documents/999", nil)
	rec := httptest.NewRecorder()
	srv.ServeHTTP(rec, req)

	if rec.Code != http.StatusNotFound {
		t.Fatalf("expected 404, got %d", rec.Code)
	}
}

func TestStatsEndpoint(t *testing.T) {
	srv, idx := setup(t)
	idx.Add(1, "hello world")
	idx.Add(2, "test document")

	req := httptest.NewRequest("GET", "/stats", nil)
	rec := httptest.NewRecorder()
	srv.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rec.Code)
	}

	var resp StatsResponse
	json.NewDecoder(rec.Body).Decode(&resp)
	if resp.Documents != 2 {
		t.Fatalf("expected 2 documents, got %d", resp.Documents)
	}
}

func TestSearchWithTopK(t *testing.T) {
	srv, idx := setup(t)
	for i := 0; i < 20; i++ {
		idx.Add(i, "search engine document about various topics")
	}

	req := httptest.NewRequest("GET", "/search?q=search+engine&k=3", nil)
	rec := httptest.NewRecorder()
	srv.ServeHTTP(rec, req)

	var resp SearchResponse
	json.NewDecoder(rec.Body).Decode(&resp)
	if len(resp.Results) != 3 {
		t.Fatalf("expected 3 results with k=3, got %d", len(resp.Results))
	}
}
