package service

import (
	"context"
	"time"
)

type CallbackFn func([]string)

// Config contains startup parameters
type Config struct {
	OnAddPods    CallbackFn
	OnRemovePads CallbackFn
}

// Start is called to start the pod finder service.  Pass in
// a cancel context to communicate application shut down.
func Start(ctx context.Context, cfg Config) error {
	// we use this channel to block until we've statrted up, returning an error or nil.
	responseCh := make(chan error, 1)

	go func() {
		ticker := time.Tick(time.Second)

		responseCh <- nil

		for {
			select {
			case <-ticker:
				// do work
			case <-ctx.Done():
				return

			}

		}

	}()

	return <-responseCh
}
