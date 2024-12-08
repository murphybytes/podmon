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
	fmt.Println("Hello")

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt)
	defer stop()
	// We want to monitor pods in namespace 'foo' with label 'app': 'busybox'
	// See deploy.yaml
	monitor, err := service.New("foo", service.KeyValues{"app": "busybox"})
	if err != nil {
		log.Fatal(err)
	}
	// Set up goroutines to listen for pods that are modified, or removed
	var eg errgroup.Group
	eg.Go(func() error {
		for evt := range monitor.Modified() {
			fmt.Printf("pod modified name: %s ip: %s\n", evt.Name, evt.IpAddress)
		}
		return nil
	})

	eg.Go(func() error {
		for evt := range monitor.Removed() {
			fmt.Printf("pod removed name: %s\n", evt.Name)
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
