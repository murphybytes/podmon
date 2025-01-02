package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"os/signal"

	"github.com/murphybytes/pod-finder/service"
	"golang.org/x/sync/errgroup"
)

func main() {

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt)
	defer stop()
	// We want to monitor pods in namespace 'foo' with label 'app': 'busybox'
	// See deploy.yaml
	monitor, err := service.New("foo", service.Labels{"app": "busybox"})
	if err != nil {
		log.Fatal(err)
	}
	// Set up goroutines to listen for pods that are modified, or removed
	var eg errgroup.Group
	eg.Go(func() error {
		for evt := range monitor.Events() {
			fmt.Printf("pod event: %s name: %s ip: %s\n", evt.EventType, evt.Name, evt.IpAddress)
		}
		return nil
	})

	// Start listening for pod changes
	if err := service.Start(ctx, monitor); err != nil {
		log.Fatalf("Program terminated with error: %v", err)
	}

	<-ctx.Done()
	fmt.Println(ctx.Err())

	eg.Wait()

}
