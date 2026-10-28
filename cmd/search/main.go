package main

import (
	"flag"
	"log/slog"
	"net/http"
	"os"

	"github.com/siddhant-rajhans/fulltext-search-engine/internal/index"
	"github.com/siddhant-rajhans/fulltext-search-engine/internal/server"
)

func main() {
	addr := flag.String("addr", ":8080", "listen address")
	flag.Parse()

	idx := index.New()
	srv := server.New(idx)

	slog.Info("starting search service", "addr", *addr)
	if err := http.ListenAndServe(*addr, srv); err != nil {
		slog.Error("server error", "error", err)
		os.Exit(1)
	}
}
