package service

import (
	"context"
	"sync"
	"testing"

	"github.com/murphybytes/pod-finder/service"
	"github.com/stretchr/testify/assert"
)

func TestService(t *testing.T) {
	podMap := map[string]string{}
	var lock sync.Mutex

	ctx, fn := context.WithCancel(context.Background())

	cfg := &service.Config{
		OnModifiedPod: func(_ context.Context, podName, podIp string) {
			lock.Lock()
			defer lock.Unlock()
			podMap[podName] = podIp
		},
	}

	defer fn()
	assert.Nil(t, Start(
		ctx,
		&cfg,
	))

}
