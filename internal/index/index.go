package index

import (
	"math"
	"sort"
	"sync"

	"github.com/siddhant-rajhans/fulltext-search-engine/internal/tokenizer"
)

type Posting struct {
	DocID     int
	Frequency int
	Positions []int
}

type Document struct {
	ID      int
	Content string
	Length  int
}

type InvertedIndex struct {
	mu        sync.RWMutex
	index     map[string][]Posting
	docs      map[int]Document
	docCount  int
	totalLen  int
	avgDocLen float64
}

func New() *InvertedIndex {
	return &InvertedIndex{
		index: make(map[string][]Posting),
		docs:  make(map[int]Document),
	}
}

func (idx *InvertedIndex) Add(docID int, content string) {
	terms := tokenizer.TokenizeWithPositions(content)

	termFreq := make(map[string][]int)
	for _, tp := range terms {
		termFreq[tp.Term] = append(termFreq[tp.Term], tp.Position)
	}

	idx.mu.Lock()
	defer idx.mu.Unlock()

	docLen := len(terms)
	idx.docs[docID] = Document{ID: docID, Content: content, Length: docLen}
	idx.docCount++
	idx.totalLen += docLen
	idx.avgDocLen = float64(idx.totalLen) / float64(idx.docCount)

	for term, positions := range termFreq {
		idx.index[term] = append(idx.index[term], Posting{
			DocID:     docID,
			Frequency: len(positions),
			Positions: positions,
		})
	}
}

func (idx *InvertedIndex) Search(query string, topK int) []SearchResult {
	queryTerms := tokenizer.Tokenize(query)
	if len(queryTerms) == 0 {
		return nil
	}

	idx.mu.RLock()
	defer idx.mu.RUnlock()

	scores := make(map[int]float64)

	for _, term := range queryTerms {
		postings, ok := idx.index[term]
		if !ok {
			continue
		}

		idf := idx.idf(len(postings))
		for _, p := range postings {
			doc := idx.docs[p.DocID]
			tf := bm25TF(p.Frequency, doc.Length, idx.avgDocLen)
			scores[p.DocID] += idf * tf
		}
	}

	results := make([]SearchResult, 0, len(scores))
	for docID, score := range scores {
		results = append(results, SearchResult{
			DocID: docID,
			Score: score,
		})
	}

	sort.Slice(results, func(i, j int) bool {
		return results[i].Score > results[j].Score
	})

	if topK > 0 && len(results) > topK {
		results = results[:topK]
	}

	return results
}

func (idx *InvertedIndex) Get(docID int) (Document, bool) {
	idx.mu.RLock()
	defer idx.mu.RUnlock()
	doc, ok := idx.docs[docID]
	return doc, ok
}

func (idx *InvertedIndex) DocCount() int {
	idx.mu.RLock()
	defer idx.mu.RUnlock()
	return idx.docCount
}

func (idx *InvertedIndex) TermCount() int {
	idx.mu.RLock()
	defer idx.mu.RUnlock()
	return len(idx.index)
}

// BM25 parameters
const (
	k1 = 1.2
	b  = 0.75
)

func (idx *InvertedIndex) idf(docFreq int) float64 {
	n := float64(idx.docCount)
	df := float64(docFreq)
	return math.Log(1 + (n-df+0.5)/(df+0.5))
}

func bm25TF(freq int, docLen int, avgDocLen float64) float64 {
	tf := float64(freq)
	dl := float64(docLen)
	return (tf * (k1 + 1)) / (tf + k1*(1-b+b*(dl/avgDocLen)))
}

type SearchResult struct {
	DocID int
	Score float64
}
