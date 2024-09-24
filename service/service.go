package service

import (
	"context"
	"os"
	"path/filepath"
	"time"

	k8s "k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/client-go/util/homedir"
)

type CallbackFn func([]string)

// Config contains startup parameters
type Config struct {
	OnAddPods    CallbackFn
	OnRemovePads CallbackFn
}

// Start is called to start the pod finder service.  Pass in
// a cancel context to communicate application shut down.
func Start(ctx context.Context, cfg Config) error {
	// we use this channel to block until we've statrted up, returning an error or nil.
	responseCh := make(chan error, 1)

	go func() {
		_, err := getK8sClient()
		if err != nil {
			responseCh <- err
			return
		}
		responseCh <- nil

		ticker := time.Tick(time.Second)

		for {
			select {
			case <-ticker:
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
