// Command leadcapture imports partner CSV files and serves the staff lead application.
package main

import (
	"context"
	"errors"
	"fmt"
	"io/fs"
	"log"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gin-gonic/gin"

	"leadcapture/internal/leads"
	"leadcapture/web"
)

func main() {
	os.Exit(execute())
}

func execute() int {
	logger, closeLog, err := setupLogging(os.Getenv("LEADS_ENV"), os.Getenv("LEADS_LOG_FILE"))
	if err != nil {
		slog.New(slog.NewJSONHandler(os.Stderr, nil)).Error("logging setup failed", "error", err.Error())
		return 1
	}
	// Always suppress Gin's unstructured debug output, including route registration.
	// LEADS_ENV independently selects slog console logging or production file logging.
	gin.SetMode(gin.ReleaseMode)
	exitCode := 0
	if runErr := run(logger); runErr != nil {
		// run returns safe messages; underlying errors can contain PII or file paths.
		logger.Error("command failed", "error", runErr.Error())
		exitCode = 1
	}
	if closeErr := closeLog(); closeErr != nil {
		slog.New(slog.NewJSONHandler(os.Stderr, nil)).Error("could not close log file")
		exitCode = 1
	}
	return exitCode
}

func run(logger *slog.Logger) error {
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
		return importFile(ctx, store, logger)
	}
	return serve(ctx, store, logger)
}

func importFile(ctx context.Context, store *leads.Store, logger *slog.Logger) error {
	const importArgCount = 3
	if len(os.Args) != importArgCount {
		return errors.New("usage: leads import <file.csv>")
	}
	name := os.Args[2]
	logger.InfoContext(ctx, fmt.Sprintf("import file [%v]", name))
	file, err := os.Open(name) //nolint:gosec // The local CLI intentionally accepts a user-selected file path.
	if err != nil {
		return errors.New("could not open CSV file")
	}
	defer file.Close()
	result, err := store.Import(ctx, file)
	for _, rejection := range result.Rejected {
		logger.WarnContext(ctx, "CSV row rejected", "line", rejection.Line, "errors", rejection.Errors)
	}
	logger.InfoContext(ctx, "CSV import processed", "imported", result.Imported, "rejected", len(result.Rejected))
	if err != nil {
		return errors.New("import stopped: check headers, file format, and database availability")
	}
	return nil
}

func serve(ctx context.Context, store *leads.Store, logger *slog.Logger) error {
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
	server := &http.Server{Addr: addr, Handler: leads.Handler(store, assets, logger),
		ErrorLog:          log.New(serverLogWriter{logger: logger}, "", 0),
		ReadHeaderTimeout: readTimeout, ReadTimeout: readTimeout, WriteTimeout: idleTimeout, IdleTimeout: idleTimeout}
	done := make(chan struct{})
	go func() {
		defer close(done)
		<-ctx.Done()
		shutdownCtx, cancel := context.WithTimeout(context.WithoutCancel(ctx), readTimeout)
		defer cancel()
		if shutdownErr := server.Shutdown(shutdownCtx); shutdownErr != nil {
			logger.ErrorContext(ctx, "HTTP server graceful shutdown failed")
			_ = server.Close()
		}
	}()
	logger.InfoContext(ctx, "HTTP server starting", "address", addr)
	err = server.ListenAndServe()
	if errors.Is(err, http.ErrServerClosed) {
		<-done
		logger.InfoContext(ctx, "HTTP server stopped")
		return nil
	}
	return errors.New("HTTP server could not start")
}
