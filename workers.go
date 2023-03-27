package main

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"os/signal"

	"github.com/prometheus/client_golang/prometheus/promhttp"
)

// Listen
func Listen(ctx context.Context, cancel context.CancelFunc, address string) func() error {
	return func() error {
		defer cancel()
		http.Handle("/metrics", promhttp.Handler())
		server := &http.Server{Addr: address}
		go func() {
			fmt.Printf("listener started on %s\n", address)
			if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
				fmt.Println("listener error:", err)
			}
		}()
		<-ctx.Done()
		err := server.Shutdown(context.Background())
		if err != nil {
			fmt.Println("listener shutdown error:", err)
		}
		return err
	}
}

func ControlC(ctx context.Context, cancel context.CancelFunc) func() error {
	return func() error {
		defer cancel()
		c := make(chan os.Signal, 1)
		signal.Notify(c, os.Interrupt)
		for {
			select {
			case <-c:
				fmt.Println("\nCtrl+C stopped!")
				return nil
			case <-ctx.Done():
				return nil
			}
		}
	}
}
