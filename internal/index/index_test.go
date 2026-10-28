package index

import (
	"fmt"
	"math"
	"sync"
	"testing"
)

func TestAddAndSearch(t *testing.T) {
	idx := New()
	idx.Add(1, "the quick brown fox jumps over the lazy dog")
	idx.Add(2, "the brown bear eats honey in the forest")
	idx.Add(3, "quick sort and merge sort algorithms")

	results := idx.Search("quick brown", 10)
	if len(results) == 0 {
		t.Fatal("expected results for 'quick brown'")
	}
	if results[0].DocID != 1 {
		t.Fatalf("expected doc 1 first, got %d", results[0].DocID)
	}
}

func TestBM25Ranking(t *testing.T) {
	idx := New()
	idx.Add(1, "database indexing and query optimization")
	idx.Add(2, "database database database performance tuning database indexing")
	idx.Add(3, "web application development framework")

	results := idx.Search("database indexing", 10)
	if len(results) < 2 {
		t.Fatalf("expected at least 2 results, got %d", len(results))
	}

	// Doc 2 has higher TF for "database", should score higher
	if results[0].DocID != 2 {
		t.Fatalf("expected doc 2 (higher TF) first, got doc %d", results[0].DocID)
	}
}

func TestIDFBoostsRareTerms(t *testing.T) {
	idx := New()
	// "common" appears in all docs, "unique" only in doc 3
	idx.Add(1, "common word here and common there")
	idx.Add(2, "common word everywhere common common")
	idx.Add(3, "unique term plus common word")

	results := idx.Search("unique common", 10)
	if len(results) == 0 {
		t.Fatal("expected results")
	}
	// Doc 3 should rank first because "unique" has high IDF
	if results[0].DocID != 3 {
		t.Fatalf("expected doc 3 (unique term) first, got doc %d", results[0].DocID)
	}
}

func TestSearchNoResults(t *testing.T) {
	idx := New()
	idx.Add(1, "hello world")

	results := idx.Search("nonexistent", 10)
	if len(results) != 0 {
		t.Fatalf("expected no results, got %d", len(results))
	}
}

func TestEmptyQuery(t *testing.T) {
	idx := New()
	idx.Add(1, "hello world")

	results := idx.Search("", 10)
	if results != nil {
		t.Fatalf("expected nil, got %v", results)
	}
}

func TestTopKLimit(t *testing.T) {
	idx := New()
	for i := 0; i < 100; i++ {
		idx.Add(i, fmt.Sprintf("document number %d about testing search", i))
	}

	results := idx.Search("document testing search", 5)
	if len(results) != 5 {
		t.Fatalf("expected 5 results, got %d", len(results))
	}
}

func TestDocCount(t *testing.T) {
	idx := New()
	if idx.DocCount() != 0 {
		t.Fatal("expected 0 docs")
	}

	idx.Add(1, "hello")
	idx.Add(2, "world")
	if idx.DocCount() != 2 {
		t.Fatalf("expected 2 docs, got %d", idx.DocCount())
	}
}

func TestGetDocument(t *testing.T) {
	idx := New()
	idx.Add(42, "test document content")

	doc, ok := idx.Get(42)
	if !ok {
		t.Fatal("expected document to exist")
	}
	if doc.Content != "test document content" {
		t.Fatalf("unexpected content: %s", doc.Content)
	}

	_, ok = idx.Get(999)
	if ok {
		t.Fatal("expected document not found")
	}
}

func TestConcurrentAddAndSearch(t *testing.T) {
	idx := New()
	var wg sync.WaitGroup

	// Concurrent adds
	for i := 0; i < 100; i++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()
			idx.Add(id, fmt.Sprintf("concurrent document %d about search engines", id))
		}(i)
	}
	wg.Wait()

	if idx.DocCount() != 100 {
		t.Fatalf("expected 100 docs after concurrent adds, got %d", idx.DocCount())
	}

	// Concurrent searches
	for i := 0; i < 50; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			results := idx.Search("search engines", 10)
			if len(results) == 0 {
				t.Error("expected results from concurrent search")
			}
		}()
	}
	wg.Wait()
}

func TestScoresArePositive(t *testing.T) {
	idx := New()
	idx.Add(1, "information retrieval systems")
	idx.Add(2, "data structures and algorithms")

	results := idx.Search("information retrieval", 10)
	for _, r := range results {
		if r.Score <= 0 {
			t.Fatalf("expected positive score, got %f for doc %d", r.Score, r.DocID)
		}
		if math.IsNaN(r.Score) || math.IsInf(r.Score, 0) {
			t.Fatalf("invalid score %f for doc %d", r.Score, r.DocID)
		}
	}
}

func TestResultsDescendingByScore(t *testing.T) {
	idx := New()
	for i := 0; i < 50; i++ {
		idx.Add(i, fmt.Sprintf("document about topic %d with search terms", i))
	}

	results := idx.Search("search document topic", 20)
	for i := 1; i < len(results); i++ {
		if results[i].Score > results[i-1].Score {
			t.Fatalf("results not sorted: score[%d]=%.4f > score[%d]=%.4f",
				i, results[i].Score, i-1, results[i-1].Score)
		}
	}
}

func BenchmarkAdd(b *testing.B) {
	idx := New()
	docs := []string{
		"The quick brown fox jumps over the lazy dog near the river bank",
		"Information retrieval systems use inverted indexes for fast lookup",
		"Database query optimization requires understanding of index structures",
		"Machine learning models process large amounts of training data",
		"Distributed systems handle concurrent requests across multiple nodes",
	}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		idx.Add(i, docs[i%len(docs)])
	}
}

func BenchmarkSearch(b *testing.B) {
	idx := New()
	for i := 0; i < 10000; i++ {
		idx.Add(i, fmt.Sprintf("document %d about various topics including search and retrieval systems with databases", i))
	}

	queries := []string{"search retrieval", "database systems", "topics various", "document search"}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		idx.Search(queries[i%len(queries)], 10)
	}
}
