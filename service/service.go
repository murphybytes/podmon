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
	core "k8s.io/client-go/kubernetes/typed/core/v1"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/client-go/util/homedir"
)

// ModfifiedEventHandlerFn recieves the pod name and an IP address
type ModifiedEventHandlerFn func(context.Context, string, string)

// DeleteEventHandler is called when a pod is removed, takes the pod name
type DeleteEventHandlerFn func(context.Context, string)
type StartHandlerFn func(context.Context, []string)
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

type EventChannelGetter func(context.Context, string, KeyValues,
	core.CoreV1Interface) (<-chan watch.Event, error)

type PodLister func(context.Context, string, KeyValues,
	core.CoreV1Interface) ([]string, error)

// Config contains startup parameters
type Config struct {
	OnModifiedPod ModifiedEventHandlerFn
	OnRemovePods  DeleteEventHandlerFn
	OnStart       StartHandlerFn
	// Namespace of the pods we are interested in
	Namespace string
	// LabelSelector key value pairs from the pod label
	LabelSelector KeyValues
}

type options struct {
	eventChannelGetterFn EventChannelGetter
	podListerFn          PodLister
}

func WithPodGetterFunction(fn EventChannelGetter) func(*options) {
	return func(opt *options) {
		opt.eventChannelGetterFn = fn
	}
}

func WithPodListerFunction(fn PodLister) func(*options) {
	return func(opt *options) {
		opt.podListerFn = fn
	}
}

// Start is called to start the pod finder service.  Pass in
// a cancel context to communicate application shut down.
func Start(ctx context.Context, cfg *Config, opts ...func(*options)) error {
	opt := options{
		eventChannelGetterFn: getPodEventChannel,
		podListerFn:          listPodIps,
	}

	for _, fn := range opts {
		fn(&opt)
	}

	// we use this channel to block until we've statrted up, returning an error or nil.
	responseCh := make(chan error, 1)

	go func() {
		client, err := getK8sClient()
		if err != nil {
			responseCh <- err
			return
		}
		podEventCh, err := opt.eventChannelGetterFn(ctx, cfg.Namespace, cfg.LabelSelector, client.CoreV1())
		if err != nil {
			responseCh <- err
			return
		}
		responseCh <- nil

		for {
			select {
			case evt := <-podEventCh:
				handleEvent(ctx, evt, handlerFns{
					onAdd:    cfg.OnModifiedPod,
					onRemove: cfg.OnRemovePods,
				})

			case <-ctx.Done():
				return
			}

		}

	}()

	return <-responseCh
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

func getPodEventChannel(
	ctx context.Context,
	ns string,
	labels KeyValues,
	client core.CoreV1Interface,
) (<-chan watch.Event, error) {
	opts := v1.ListOptions{
		LabelSelector: labels.LabelSelector(),
		Watch:         true,
	}
	watch, err := client.Pods(ns).Watch(ctx, opts)
	if err != nil {
		return nil, err
	}
	return watch.ResultChan(), nil
}

func listPodIps(
	ctx context.Context,
	ns string,
	labels KeyValues,
	client core.CoreV1Interface,
) ([]string, error) {
	return nil, nil
}
