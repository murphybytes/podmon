package service

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"golang.org/x/sync/errgroup"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/watch"
)

type mockMonitor struct {
	eventCh    chan watch.Event
	modifiedCh chan PodModifiedEvent
	removedCh  chan PodRemovedEvent
	monitorErr error
}

func newMock() *mockMonitor {
	return &mockMonitor{
		eventCh:    make(chan watch.Event),
		modifiedCh: make(chan PodModifiedEvent),
		removedCh:  make(chan PodRemovedEvent),
	}
}

func (m *mockMonitor) Modified() <-chan PodModifiedEvent {
	return m.modifiedCh
}

func (m *mockMonitor) Removed() <-chan PodRemovedEvent {
	return m.removedCh
}

func (m *mockMonitor) handlers() (chan<- PodModifiedEvent, chan<- PodRemovedEvent) {
	return m.modifiedCh, m.removedCh
}

func (m *mockMonitor) monitor(_ context.Context) (<-chan watch.Event, error) {
	return m.eventCh, m.monitorErr
}

func makePod(name, ip string) *corev1.Pod {
	var pod corev1.Pod
	pod.Name = name
	pod.Status.PodIP = ip
	return &pod
}

func TestService(t *testing.T) {
	tt := []struct {
		name         string
		input        []watch.Event
		modified     []PodModifiedEvent
		removed      []PodRemovedEvent
		monitorError error
	}{
		{
			name: "happy path",
			input: []watch.Event{
				{
					Type:   watch.Added,
					Object: makePod("pod1", "10.10.10.10"),
				},
			},
			modified: []PodModifiedEvent{
				{Name: "pod1", IpAddress: "10.10.10.10"},
			},
		},
		{
			name: "simple removal",
			input: []watch.Event{
				{
					Type:   watch.Added,
					Object: makePod("pod1", "10.10.10.10"),
				},
				{
					Type:   watch.Deleted,
					Object: makePod("pod1", ""),
				},
			},
			modified: []PodModifiedEvent{
				{Name: "pod1", IpAddress: "10.10.10.10"},
			},
			removed: []PodRemovedEvent{
				{Name: "pod1"},
			},
		},
	}

	for _, tc := range tt {
		t.Run(tc.name, func(t *testing.T) {
			ctx, stop := context.WithCancel(context.Background())
			defer stop()

			mon := newMock()
			var modified []PodModifiedEvent
			var removed []PodRemovedEvent

			var g errgroup.Group
			g.Go(func() error {
				for evt := range mon.Modified() {
					modified = append(modified, evt)
				}
				return nil
			})
			g.Go(func() error {
				for evt := range mon.Removed() {
					removed = append(removed, evt)
				}
				return nil
			})

			assert.Nil(t, Start(ctx, mon))

			for _, evt := range tc.input {
				mon.eventCh <- evt
			}

			stop()

			<-ctx.Done()
			g.Wait()

			assert.ElementsMatch(t, modified, tc.modified)

		})
	}
	assert.Nil(t, nil)
}
