package main

import (
	"context"
	"flag"
	"fmt"
	"os"
	"os/signal"
	"sync"
	"time"
)

var (
	timeout time.Duration
)

func init() {
	flag.DurationVar(&timeout, "timeout", 10*time.Second, "timeout period (e.g., 300ms, 2h45m, 15s)")
}

func main() {
	flag.Parse()

	host := flag.Arg(0)
	port := flag.Arg(1)
	address := host + ":" + port

	client := NewTelnetClient(address, timeout, os.Stdin, os.Stdout)

	if err := client.Connect(); err != nil {
		fmt.Fprintf(os.Stderr, "Failed to connect: %v\n", err)
		os.Exit(1)
	}
	defer client.Close()

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt)
	defer stop()

	var wg sync.WaitGroup
	errors := make(chan error, 2)

	wg.Add(1)
	go func() {
		defer wg.Done()
		if err := client.Send(); err != nil {
			errors <- fmt.Errorf("send error: %w", err)
		}
	}()

	wg.Add(1)
	go func() {
		defer wg.Done()
		if err := client.Receive(); err != nil {
			errors <- fmt.Errorf("receive error: %w", err)
		}
	}()

	select {
	case <-ctx.Done():
		fmt.Fprintln(os.Stderr, "\nReceived interrupt signal")
	case err := <-errors:
		fmt.Fprintf(os.Stderr, "Connection closed: %v\n", err)
	}

	client.Close()
}
