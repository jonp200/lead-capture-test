// Command leadcapture imports partner CSV files and serves the staff lead application.
package main

import (
	"context"
	"errors"
	"fmt"
	"io/fs"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"leadcapture/internal/leads"
	"leadcapture/web"
)

func main() {
	if err := run(); err != nil {
		// Underlying errors can contain PII or file paths; print safe messages only.
		fmt.Fprintln(os.Stderr, err.Error())
		os.Exit(1)
	}
}

func run() error {
	if len(os.Args) < 2 || (os.Args[1] != "serve" && os.Args[1] != "import") {
		return errors.New("usage: leads serve | leads import <file.csv>")
	}
	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer cancel()
	dbPath := os.Getenv("LEADS_DB")
	if dbPath == "" {
		dbPath = "leads.db"
	}
	store, err := leads.Open(ctx, dbPath)
	if err != nil {
		return errors.New("could not open database")
	}
	defer store.Close()
	if os.Args[1] == "import" {
		return importFile(ctx, store)
	}
	return serve(ctx, store)
}

func importFile(ctx context.Context, store *leads.Store) error {
	const importArgCount = 3
	if len(os.Args) != importArgCount {
		return errors.New("usage: leads import <file.csv>")
	}
	file, err := os.Open(os.Args[2]) //nolint:gosec // The local CLI intentionally accepts a user-selected file path.
	if err != nil {
		return errors.New("could not open CSV file")
	}
	defer file.Close()
	result, err := store.Import(ctx, file)
	for _, rejection := range result.Rejected {
		fmt.Fprintf(os.Stderr, "line %d: %v\n", rejection.Line, rejection.Errors)
	}
	fmt.Fprintf(os.Stdout, "Imported %d; rejected %d\n", result.Imported, len(result.Rejected))
	if err != nil {
		return errors.New("import stopped: check headers, file format, and database availability")
	}
	return nil
}

func serve(ctx context.Context, store *leads.Store) error {
	assets, err := fs.Sub(web.Assets, "dist")
	if err != nil {
		return errors.New("could not load frontend")
	}
	addr := os.Getenv("LEADS_ADDR")
	if addr == "" {
		addr = "127.0.0.1:8080"
	}
	const readTimeout = 15 * time.Second
	const idleTimeout = 60 * time.Second
	server := &http.Server{Addr: addr, Handler: leads.Handler(store, assets),
		ReadHeaderTimeout: readTimeout, ReadTimeout: readTimeout, WriteTimeout: idleTimeout, IdleTimeout: idleTimeout}
	done := make(chan struct{})
	go func() {
		defer close(done)
		<-ctx.Done()
		shutdownCtx, cancel := context.WithTimeout(context.WithoutCancel(ctx), readTimeout)
		defer cancel()
		if shutdownErr := server.Shutdown(shutdownCtx); shutdownErr != nil {
			_ = server.Close()
		}
	}()
	fmt.Fprintf(os.Stdout, "Lead capture listening on %s\n", addr)
	err = server.ListenAndServe()
	if errors.Is(err, http.ErrServerClosed) {
		<-done
		return nil
	}
	return errors.New("HTTP server could not start")
}
