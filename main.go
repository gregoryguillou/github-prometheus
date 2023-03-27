package main

import (
	"context"
	"os"

	"golang.org/x/sync/errgroup"
)

func main() {
	metrics, config, err := ParseArgs(os.Args)
	if err != nil {
		os.Exit(2)
	}
	g, ctx := errgroup.WithContext(context.Background())
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()
	g.Go(Listen(ctx, cancel, config.Listener))
	g.Go(ControlC(ctx, cancel))
	g.Go(runMetrics(ctx, cancel, *metrics))
	g.Wait()
}
