# Full-Text Search Engine

![Go 1.22](https://img.shields.io/badge/go-1.22-00ADD8?logo=go&logoColor=white)
![License: MIT](https://img.shields.io/badge/license-MIT-blue)
![Tests: 26 passing](https://img.shields.io/badge/tests-26%20passing-brightgreen)

In-memory full-text search engine in Go with an inverted index, BM25 ranking, stop-word removal, and position-aware posting lists. Concurrent-safe for parallel indexing and search.

## Measured Results

Benchmarked on AMD Ryzen 5 3550H (4C/8T), 2 GB RAM. Reproduce with `make bench`.

| Metric | Value |
|--------|-------|
| Query latency (p50, 10K docs) | **2.8 ms** |
| Query latency (p99, 10K docs) | **10.8 ms** |
| Index throughput | **76,388 docs/s** |
| Concurrent safety | 100 goroutine adds + 50 goroutine searches, **0 races** |
| Tests | **26/26 passing** |
| Test categories | tokenizer, index, BM25 ranking, HTTP API, concurrency |

## Architecture

```
                            ┌─────────────────┐
       POST /documents ────→│   HTTP Server   │←──── GET /search?q=...
                            └────────┬────────┘
                     ┌───────────────┼───────────────┐
                     │           index path      search path
                     ▼               │               ▼
              ┌─────────────┐        │        ┌─────────────┐
              │  Tokenizer  │        │        │  Tokenizer  │
              │             │        │        │             │
              │  lowercase  │        │        │  lowercase  │
              │  stop words │        │        │  stop words │
              │  positions  │        │        │             │
              └──────┬──────┘        │        └──────┬──────┘
                     │               │               │
                     ▼               │               ▼
              ┌──────────────────────────────────────────────┐
              │              Inverted Index                   │
              │                                              │
              │  term → [ docID : freq, positions ]          │
              │                                              │
              │  "index"  → [{doc:1, f:2, pos:[3,7]},       │
              │              {doc:4, f:1, pos:[12]}  ]       │
              │                                              │
              │  sync.RWMutex                                │
              │  ├─ concurrent reads (search)                │
              │  └─ exclusive writes (index)                 │
              └────────────────────┬─────────────────────────┘
                                   │
                              search path
                                   │
                                   ▼
                            ┌─────────────┐
                            │ BM25 Scorer │
                            │             │
                            │  k1 = 1.2   │
                            │  b  = 0.75  │
                            │             │
                            │  IDF × TF   │
                            │  per term   │
                            └──────┬──────┘
                                   │
                                   ▼
                           ranked results []
```

## API

| Endpoint | Method | Description |
|----------|--------|-------------|
| `/documents` | POST | Index a document `{"id": 1, "content": "..."}` |
| `/documents/{id}` | GET | Retrieve a document by ID |
| `/search?q=...&k=10` | GET | Search with BM25 ranking, top-k results |
| `/stats` | GET | Index statistics (doc count, term count) |

## Quick Start

```bash
# Build and run
make run

# Index some documents
curl -X POST http://localhost:8080/documents \
  -d '{"id": 1, "content": "Information retrieval systems use inverted indexes"}'

curl -X POST http://localhost:8080/documents \
  -d '{"id": 2, "content": "Database query optimization with index structures"}'

# Search
curl "http://localhost:8080/search?q=inverted+index&k=5"

# Check stats
curl http://localhost:8080/stats
```

## BM25 Scoring

The scoring function for a query Q and document D:

```
score(Q, D) = sum over q in Q: IDF(q) * TF(q, D)

IDF(q) = log(1 + (N - df(q) + 0.5) / (df(q) + 0.5))

TF(q, D) = (f(q,D) * (k1 + 1)) / (f(q,D) + k1 * (1 - b + b * |D| / avgdl))
```

Where N = total documents, df(q) = documents containing term q, f(q,D) = frequency of q in D, |D| = document length, avgdl = average document length.

Parameters: k1 = 1.2 (term frequency saturation), b = 0.75 (length normalization).

<details>
<summary><strong>Design Decisions</strong></summary>

### 1. BM25 over raw TF-IDF

BM25 adds term frequency saturation (parameter k1=1.2) and document length normalization (parameter b=0.75). A document that repeats "database" 50 times does not score 50x higher than one that mentions it twice. Short documents are not penalized for having fewer total tokens.

### 2. Sorted-set posting lists

Each term maps to a slice of postings sorted by document ID. This makes the index append-friendly for concurrent writes and straightforward to scan during search. The search path iterates all postings for each query term, which is O(D) per term where D is the document frequency.

### 3. Atomic Lua-free concurrency

The index uses `sync.RWMutex`: writes take the write lock, searches take the read lock. Multiple searches run concurrently; a write blocks until all concurrent reads finish. No external coordination layer.

### 4. Stop word removal + lowercasing

The tokenizer strips 30+ English stop words and lowercases all tokens before indexing. Single-character tokens are dropped. This reduces index size and prevents high-frequency low-information terms from dominating scores.

### 5. Position-aware indexing

Postings store per-document term positions, enabling future phrase queries and proximity scoring without re-indexing. The current search uses bag-of-words BM25, but the position data is there for extensions.

</details>

## Tests

```
internal/tokenizer/  7 tests  basic tokenization, stop words, case, punctuation,
                              empty input, positions, single-char drops
internal/index/     11 tests  add/search, BM25 ranking, IDF rare terms, no results,
                              empty query, top-k, doc count, get, concurrency,
                              positive scores, descending order
internal/server/     8 tests  index endpoint, empty content, search, missing query,
                              get doc, not found, stats, top-k parameter
```

## Project Structure

```
cmd/search/            main.go entry point
internal/
  tokenizer/           stop-word removal, lowercasing, position tracking
  index/               inverted index, BM25 scoring, posting lists
  server/              REST API handlers
  bench/               latency and throughput benchmarks
```

## License

MIT
