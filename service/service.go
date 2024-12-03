package service

import (
	"context"
)

// Start is called to start the pod finder service.  Pass in
// a cancel context to communicate application shut down.
func Start(ctx context.Context, m Monitorable) error {

	// we use this channel to block until we've statrted up, returning an error or nil.
	responseCh := make(chan error, 1)

	go func() {
		modifiedCh, removedCh := m.handlers()
		defer func() {
			close(modifiedCh)
			close(removedCh)
		}()

		podEvents, err := m.monitor(ctx)
		if err != nil {
			responseCh <- err
			return
		}

		responseCh <- nil

		for {
			select {
			case evt := <-podEvents:
				handleEvent(evt, handlerFns{
					onAdd:    modifiedCh,
					onRemove: removedCh,
				})

			case <-ctx.Done():
				return
			}

		}

	}()

	return <-responseCh
}
