package service

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestService(t *testing.T) {
	var cfg Config
	ctx, fn := context.WithCancel(context.Background())
	defer fn()
	assert.Nil(t, Start(ctx, cfg))

}
