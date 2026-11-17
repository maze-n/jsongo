package main

import (
	"context"
	"errors"
	"flag"
	"fmt"
	"log"
	"net"
	"net/http"
	"os"
	"os/signal"
	"path/filepath"
	"sync"
	"syscall"
	"time"

	"github.com/maze-n/jsongo/internal/api"
)

type Store struct {
	my   sync.Mutex
	data map[string][]map[string]any
}

func main() {
	port := flag.Int("port", 8080, "Server port number")
	flag.Usage = func() {
		fmt.Fprintf(os.Stderr, "Usage: %s [options] <filename>\n\nOptions:\n", os.Args[0])
		flag.PrintDefaults()
	}

	flag.Parse()

	if flag.NArg() < 1 {
		log.Fatalf("Error: missing required file argument")
		flag.Usage()
		os.Exit(1)
	}

	filename := flag.Arg(0)

	if filepath.Base(filename) != filename {
		fmt.Fprintf(os.Stderr, "Error: %q must be a file in the current directory\n", filename)
		os.Exit(1)
	}

	info, err := os.Stat(filename)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: cannot access file %q: %v\n", filename, err)
		os.Exit(1)
	}
	if info.IsDir() {
		fmt.Fprintf(os.Stderr, "Error: %q is a directory, not a file\n", filename)
		os.Exit(1)
	}

	run(*port)
}

func run(port int) error {
	httpServer := &http.Server{
		Addr:              fmt.Sprintf(":%d", port),
		Handler:           api.HandleRoutes(),
		ReadHeaderTimeout: 5 * time.Second,
	}
	ctx, stop := signal.NotifyContext(
		context.Background(),
		syscall.SIGINT,
		syscall.SIGTERM,
	)
	defer stop()

	serverErrors := make(chan error, 1)
	httpListener, err := net.Listen("tcp", httpServer.Addr)
	if err != nil {
		return err
	}

	go func() {
		log.Printf("JsonGo server listening on :%d", port)

		if err := httpServer.Serve(httpListener); err != nil &&
			!errors.Is(err, http.ErrServerClosed) {
			serverErrors <- err
		}
	}()

	var runErr error

	select {
	case <-ctx.Done():
		log.Println("Shutdown signal received")

	case runErr = <-serverErrors:
		log.Printf("Server failed: %v", runErr)
	}

	return runErr
}
