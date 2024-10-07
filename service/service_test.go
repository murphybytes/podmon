package service

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"k8s.io/apimachinery/pkg/watch"
	core "k8s.io/client-go/kubernetes/typed/core/v1"
)

type mockWatch struct {
	ch chan watch.Event
}

func (m *mockWatch) ResultChan() <-chan watch.Event {
	return m.ch
}

func (m *mockWatch) Stop() {}

func mockPodGetter(_ context.Context, _ string, _ KeyValues, _ core.CoreV1Interface) (watch.Interface, error) {
	return nil, nil
}

func getMockPodGetter(itf watch.Interface) PodGetter {
	return func(_ context.Context, _ string, _ KeyValues, _ core.CoreV1Interface) (watch.Interface, error) {
		return itf, nil
	}
}

func TestService(t *testing.T) {
	watcher := &mockWatch{
		ch: make(chan watch.Event, 1),
	}
	var cfg Config
	ctx, fn := context.WithCancel(context.Background())
	defer fn()
	assert.Nil(t, Start(
		ctx,
		cfg,
		WithPodGetterFunction(getMockPodGetter(watcher)),
	))

}
