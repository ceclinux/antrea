package egress

import (
	"context"
	"sync"

	"github.com/vmware-tanzu/antrea/pkg/agent"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"k8s.io/apimachinery/pkg/fields"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/watch"
	clientset "k8s.io/client-go/kubernetes"
)

func NewEgressController(
	kubeClient clientset.Interface,
	nodeName string,
) *Controller {
	c := &Controller{
		kubeClient: kubeClient,
	}
	// Use nodeName to filter resources when watching resources.
	options := metav1.ListOptions{
		FieldSelector: fields.OneTermEqualSelector("nodeName", nodeName).String(),
	}
	c.networkPolicyWatcher = &watcher{
		objectType: "EgressPolicy",
		watchFunc: func() (watch.Interface, error) {
			antreaClient, err := c.antreaClientProvider.GetAntreaClient()
			if err != nil {
				return nil, err
			}
			return antreaClient.ControlplaneV1beta2().EgressPolicy().Watch(context.TODO(), options)
		}}
	return c
}

type Controller struct {
	kubeClient           clientset.Interface
	antreaClientProvider agent.AntreaClientProvider
	networkPolicyWatcher *watcher
}

type watcher struct {
	// objectType is the type of objects being watched, used for logging.
	objectType string
	// watchFunc is the function that starts the watch.
	watchFunc func() (watch.Interface, error)
	// AddFunc is the function that handles added event.
	AddFunc func(obj runtime.Object) error
	// UpdateFunc is the function that handles modified event.
	UpdateFunc func(obj runtime.Object) error
	// DeleteFunc is the function that handles deleted event.
	DeleteFunc func(obj runtime.Object) error
	// ReplaceFunc is the function that handles init events.
	ReplaceFunc func(objs []runtime.Object) error
	// connected represents whether the watch has connected to apiserver successfully.
	connected bool
	// lock protects connected.
	lock sync.RWMutex
	// group to be notified when each watcher receives bookmark event
	fullSyncWaitGroup *sync.WaitGroup
	// fullSynced indicates if the resource has been synced at least once since agent started.
	fullSynced bool
}
