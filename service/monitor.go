package service

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/watch"
	k8s "k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/client-go/util/homedir"
)

type KeyValues map[string]string

func (kv *KeyValues) LabelSelector() string {
	if kv == nil {
		return ""
	}
	selectors := make([]string, 0, len(*kv))
	for k, v := range *kv {
		selectors = append(selectors, fmt.Sprintf("%s=%s", k, v))
	}
	return strings.Join(selectors, ",")
}

type PodModifiedEvent struct {
	Name      string
	IpAddress string
}

type PodRemovedEvent struct {
	Name string
}

type Monitorable interface {
	Modified() <-chan PodModifiedEvent
	Removed() <-chan PodRemovedEvent

	monitor(context.Context) (<-chan watch.Event, error)
	handlers() (chan<- PodModifiedEvent, chan<- PodRemovedEvent)
}

type Monitor struct {
	client     *k8s.Clientset
	selectors  KeyValues
	namespace  string
	modifiedCh chan PodModifiedEvent
	removedCh  chan PodRemovedEvent
}

// New returns a monitor that can be used to track removed and modified
// pods identified by the namespace and selectors arguments.
func New(namespace string, selectors KeyValues) (*Monitor, error) {
	client, err := getK8sClient()
	if err != nil {
		return nil, err
	}

	return &Monitor{
		client:     client,
		modifiedCh: make(chan PodModifiedEvent),
		removedCh:  make(chan PodRemovedEvent),
	}, err
}

// Modified returns a channel that will produce an event containing pod name
// and IP address with a pod is added or modified
func (pf *Monitor) Modified() <-chan PodModifiedEvent { return pf.modifiedCh }

// Removed returns a channel that will produce and event contain the pod name
// when the pod is removed
func (pf *Monitor) Removed() <-chan PodRemovedEvent { return pf.removedCh }

func (pf *Monitor) handlers() (chan<- PodModifiedEvent, chan<- PodRemovedEvent) {
	return pf.modifiedCh, pf.removedCh
}

func (pf *Monitor) monitor(ctx context.Context) (<-chan watch.Event, error) {
	opts := v1.ListOptions{
		LabelSelector: pf.selectors.LabelSelector(),
		Watch:         true,
	}

	client := pf.client.CoreV1()
	watch, err := client.Pods(pf.namespace).Watch(ctx, opts)
	if err != nil {
		return nil, err
	}

	return watch.ResultChan(), nil
}

func getK8sClient() (*k8s.Clientset, error) {
	var configFileName string
	if v, ok := os.LookupEnv("KUBECONFIG"); ok {
		configFileName = v
	} else {
		configFileName = filepath.Join(homedir.HomeDir(), ".kube", "config")
	}
	result, err := clientcmd.BuildConfigFromFlags("", configFileName)
	if err != nil {
		return nil, err
	}
	return k8s.NewForConfig(result)
}
