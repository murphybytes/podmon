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
	modifiedCh chan PodEvent
	monitorErr error
}

func newMock() *mockMonitor {
	return &mockMonitor{
		eventCh:    make(chan watch.Event),
		modifiedCh: make(chan PodEvent),
	}
}

func (m *mockMonitor) Events() chan PodEvent {
	return m.modifiedCh
}

func (m *mockMonitor) monitor(_ context.Context) (<-chan watch.Event, error) {
	return m.eventCh, m.monitorErr
}

func (m *mockMonitor) listPods(_ context.Context) (Pods, error) {
	return Pods{}, nil
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
		modified     []PodEvent
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
			modified: []PodEvent{
				{Name: "pod1", EventType: watch.Added, IpAddress: "10.10.10.10"},
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
			modified: []PodEvent{
				{Name: "pod1", EventType: watch.Added, IpAddress: "10.10.10.10"},
				{Name: "pod1", EventType: watch.Deleted, IpAddress: ""},
			},
		},
	}

	for _, tc := range tt {
		t.Run(tc.name, func(t *testing.T) {
			ctx, stop := context.WithCancel(context.Background())
			defer stop()

			mon := newMock()
			var modified []PodEvent

			var g errgroup.Group
			g.Go(func() error {
				for evt := range mon.Events() {
					modified = append(modified, evt)
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
