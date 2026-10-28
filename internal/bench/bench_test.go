package bench

import (
	"fmt"
	"sort"
	"testing"
	"time"

	"github.com/siddhant-rajhans/fulltext-search-engine/internal/index"
)

func TestSearchLatency(t *testing.T) {
	idx := index.New()

	docs := []string{
		"Information retrieval is the activity of obtaining information system resources",
		"A search engine is a software system designed to carry out web searches",
		"An inverted index is an index data structure storing a mapping from content",
		"BM25 is a bag of words retrieval function that ranks documents based on query terms",
		"The term frequency inverse document frequency is a numerical statistic",
	}

	for i := 0; i < 10000; i++ {
		idx.Add(i, fmt.Sprintf("%s document number %d with additional context about search engines and information retrieval systems used in production", docs[i%len(docs)], i))
	}

	queries := []string{
		"search engine",
		"information retrieval",
		"inverted index",
		"BM25 ranking",
		"document frequency",
	}

	n := 1000
	latencies := make([]time.Duration, n)
	for i := 0; i < n; i++ {
		q := queries[i%len(queries)]
		start := time.Now()
		results := idx.Search(q, 10)
		latencies[i] = time.Since(start)
		_ = results
	}

	sort.Slice(latencies, func(i, j int) bool { return latencies[i] < latencies[j] })

	var total time.Duration
	for _, l := range latencies {
		total += l
	}

	t.Logf("Search latency over 10K docs (1000 queries):")
	t.Logf("  avg: %v", total/time.Duration(n))
	t.Logf("  p50: %v", latencies[n/2])
	t.Logf("  p99: %v", latencies[n*99/100])
	t.Logf("  Documents: %d", idx.DocCount())
	t.Logf("  Terms: %d", idx.TermCount())

	// Index speed
	idx2 := index.New()
	start := time.Now()
	for i := 0; i < 50000; i++ {
		idx2.Add(i, fmt.Sprintf("document %d about search engines and information retrieval systems with various terms", i))
	}
	indexTime := time.Since(start)
	t.Logf("Index 50K docs: %v (%.0f docs/s)", indexTime, 50000.0/indexTime.Seconds())
}
