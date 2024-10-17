package service

import (
	"context"

	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/watch"
)

type handlerFns struct {
	onAdd    ModifiedEventHandlerFn
	onRemove DeleteEventHandlerFn
}

func handleEvent(ctx context.Context, evt watch.Event, fns handlerFns) {
	switch evt.Type {
	case watch.Added, watch.Modified:
		handleAddEvent(ctx, evt, fns.onAdd)
	case watch.Deleted:
		handleDeleteEvent(ctx, evt, fns.onRemove)
	}
}

func handleAddEvent(ctx context.Context, evt watch.Event, onAdd ModifiedEventHandlerFn) {
	pod, ok := evt.Object.(*corev1.Pod)
	if ok {
		if pod.Status.PodIP != "" {
			onAdd(ctx, pod.Name, pod.Status.PodIP)
		}
	}

}

func handleDeleteEvent(ctx context.Context, evt watch.Event, onDelete DeleteEventHandlerFn) {
	pod, ok := evt.Object.(*corev1.Pod)
	if ok {
		onDelete(ctx, pod.Name)
	}
}
