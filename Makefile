.PHONY: test bench build run clean

test:
	go test ./... -v -count=1 -timeout 60s

bench:
	go test ./internal/bench/ -v -run TestSearchLatency -timeout 30s
	go test ./internal/index/ -bench=. -benchmem -timeout 30s

build:
	go build -o bin/search ./cmd/search

run: build
	./bin/search -addr :8080

clean:
	rm -rf bin/
