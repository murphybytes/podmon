package service

import (
	"context"
	"os"
	"path/filepath"

	"k8s.io/apimachinery/pkg/watch"
	k8s "k8s.io/client-go/kubernetes"
	core "k8s.io/client-go/kubernetes/typed/core/v1"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/client-go/util/homedir"
)

type CallbackFn func(string)
type KeyValues map[string]string
type PodGetter func(context.Context, string, KeyValues,
	core.CoreV1Interface) (watch.Interface, error)

// Config contains startup parameters
type Config struct {
	OnAddPods    CallbackFn
	OnRemovePods CallbackFn
	// Namespace of the pods we are interested in
	Namespace string
	// LabelSelector key value pairs from the pod label
	LabelSelector KeyValues
}

type options struct {
	podMatcherFn PodGetter
}

func WithPodGetterFunction(fn PodGetter) func(*options) {
	return func(opt *options) {
		opt.podMatcherFn = fn
	}
}

// Start is called to start the pod finder service.  Pass in
// a cancel context to communicate application shut down.
func Start(ctx context.Context, cfg Config, opts ...func(*options)) error {
	opt := options{
		podMatcherFn: getMatchingPods,
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
		watcher, err := opt.podMatcherFn(ctx, cfg.Namespace, cfg.LabelSelector, client.CoreV1())
		if err != nil {
			responseCh <- err
			return
		}
		responseCh <- nil

		for {
			select {
			case <-watcher.ResultChan():
				// do work

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

func getMatchingPods(
	ctx context.Context,
	ns string,
	labels KeyValues,
	client core.CoreV1Interface,
) (watch.Interface, error) {
	return nil, nil
}
