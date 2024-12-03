package service

import (
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/watch"
)

type handlerFns struct {
	onAdd    chan<- PodModifiedEvent
	onRemove chan<- PodRemovedEvent
}

func handleEvent(evt watch.Event, fns handlerFns) {
	switch evt.Type {
	case watch.Added, watch.Modified:
		handleAddEvent(evt, fns.onAdd)
	case watch.Deleted:
		handleDeleteEvent(evt, fns.onRemove)
	}
}

func handleAddEvent(evt watch.Event, onAdd chan<- PodModifiedEvent) {
	pod, ok := evt.Object.(*corev1.Pod)
	if ok {
		if pod.Status.PodIP != "" {
			onAdd <- PodModifiedEvent{
				Name:      pod.Name,
				IpAddress: pod.Status.PodIP,
			}
		}
	}

}

func handleDeleteEvent(evt watch.Event, onDelete chan<- PodRemovedEvent) {
	pod, ok := evt.Object.(*corev1.Pod)
	if ok {
		onDelete <- PodRemovedEvent{
			Name: pod.Name,
		}
	}
}
