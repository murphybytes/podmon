package service

import (
	"context"

	"k8s.io/apimachinery/pkg/watch"
)

type handlerFns struct {
	onAdd    EventHandlerFn
	onRemove EventHandlerFn
}

func handleEvent(ctx context.Context, evt watch.Event, fns handlerFns) {
	switch evt.Type {
	case watch.Added:
		handleAddEvent(ctx, evt, fns.onAdd)
	case watch.Deleted:
		handleDeleteEvent(ctx, evt, fns.onRemove)
	}
}

func handleAddEvent(ctx context.Context, evt watch.Event, onAdd EventHandlerFn) {

}

func handleDeleteEvent(ctx context.Context, evt watch.Event, onDelete EventHandlerFn) {

}
