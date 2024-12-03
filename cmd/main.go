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

	monitor, err := service.New("foo", service.KeyValues{"app": "busybox"})
	if err != nil {
		log.Fatal(err)
	}

	var eg errgroup.Group
	eg.Go(func() error {
		for evt := range monitor.Modified() {
			fmt.Printf("pod modified name: %s ip: %s", evt.Name, evt.IpAddress)
		}
		return nil
	})

	eg.Go(func() error {
		for evt := range monitor.Removed() {
			fmt.Printf("pod removed name: %s", evt.Name)
		}
		return nil
	})

	if err := service.Start(ctx, monitor); err != nil {
		log.Fatalf("Program terminated with error: %v", err)
	}

	<-ctx.Done()
	fmt.Println(ctx.Err())

	eg.Wait()

}
