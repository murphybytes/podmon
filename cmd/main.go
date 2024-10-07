package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"os/signal"

	"github.com/murphybytes/pod-finder/service"
)

func main() {
	fmt.Println("Hello")

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt)
	defer stop()

	var cfg service.Config
	cfg.OnAddPods = func(_ context.Context, ip string) {
		fmt.Printf("Added IP: %s\n", ip)
	}
	cfg.OnRemovePods = func(_ context.Context, ip string) {
		fmt.Printf("Removed IP: %s\n", ip)
	}
	cfg.OnStart = func(_ context.Context, ips []string) {
		fmt.Printf("On Start: %v\n", ips)
	}
	cfg.LabelSelector = service.KeyValues{
		"app": "busybox",
	}
	cfg.Namespace = "foo"
	if err := service.Start(ctx, &cfg); err != nil {
		log.Fatalf("Program terminated with error: %v", err)
	}

	<-ctx.Done()
	fmt.Println(ctx.Err())

}
