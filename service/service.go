package service

import (
	"context"

	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/watch"
)

// Start is called to start the pod finder service.  Pass in
// a cancel context to communicate application shut down.
func Start(ctx context.Context, m Monitorable) error {

	// we use this channel to block until we've started up, returning an error or nil.
	responseCh := make(chan error, 1)

	go func() {
		eventCh := m.Events()
		defer func() {
			close(eventCh)
		}()

		// get list existing pods
		initialPods, err := m.listPods(ctx)
		if err != nil {
			responseCh <- err
			return
		}

		for _, pod := range initialPods {
			eventCh <- pod
		}

		podEvents, err := m.monitor(ctx)
		if err != nil {
			responseCh <- err
			return
		}

		responseCh <- nil

		for {
			select {
			case evt := <-podEvents:
				if pod, ok := evt.Object.(*corev1.Pod); ok {
					pe := PodEvent{
						Name:      pod.Name,
						EventType: evt.Type,
						IpAddress: pod.Status.PodIP,
					}
					switch evt.Type {
					case watch.Added, watch.Modified:
						if pe.IpAddress != "" {
							eventCh <- pe
						}
					case watch.Deleted:
						eventCh <- pe
					}
				}
			case <-ctx.Done():
				return
			}

		}

	}()

	return <-responseCh
}
