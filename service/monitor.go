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

// ModifiedBuffSz is the size of the modified channel, we want to leave a little
const ChannelSz int = 100

type Labels map[string]string

func (kv *Labels) Selector() string {
	if kv == nil {
		return ""
	}
	selectors := make([]string, 0, len(*kv))
	for k, v := range *kv {
		selectors = append(selectors, fmt.Sprintf("%s=%s", k, v))
	}
	return strings.Join(selectors, ",")
}

type PodEvent struct {
	EventType watch.EventType
	Name      string
	IpAddress string
}

type Pods []PodEvent
type Pod = PodEvent

type Monitorable interface {
	Events() chan PodEvent

	monitor(context.Context) (<-chan watch.Event, error)
	listPods(context.Context) (Pods, error)
}

// Monitor returns channels that will emit events when pods are modified or
// removed.
type Monitor struct {
	client    *k8s.Clientset
	selectors Labels
	namespace string
	eventCh   chan PodEvent
}

// New returns a monitor that can be used to track removed and modified
// pods identified by the namespace and labels arguments.
func New(namespace string, selectors Labels) (*Monitor, error) {
	client, err := getK8sClient()
	if err != nil {
		return nil, err
	}

	return &Monitor{
		client:    client,
		namespace: namespace,
		selectors: selectors,
		eventCh:   make(chan PodEvent, ChannelSz),
	}, err
}

// Events returns a channel that will produce an event containing pod name
// and IP address with a pod is added or modified. Expect to get multiple
// events for the same pod.
func (pf *Monitor) Events() chan PodEvent { return pf.eventCh }

func (pf *Monitor) monitor(ctx context.Context) (<-chan watch.Event, error) {
	opts := v1.ListOptions{
		LabelSelector: pf.selectors.Selector(),
		Watch:         true,
	}

	client := pf.client.CoreV1()
	watch, err := client.Pods(pf.namespace).Watch(ctx, opts)
	if err != nil {
		return nil, err
	}

	return watch.ResultChan(), nil
}

func (pf *Monitor) listPods(ctx context.Context) (Pods, error) {
	pods := pf.client.CoreV1().Pods(pf.namespace)
	podList, err := pods.List(ctx, v1.ListOptions{
		LabelSelector: pf.selectors.Selector(),
	})
	if err != nil {
		return nil, err
	}
	var results Pods
	for _, pod := range podList.Items {
		results = append(results, Pod{
			EventType: watch.Added,
			Name:      pod.Name,
			IpAddress: pod.Status.PodIP,
		})
	}
	return results, nil
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
