package egress

import (
	"context"

	"fmt"
	"sync"
	"time"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"k8s.io/apimachinery/pkg/util/wait"

	"github.com/vmware-tanzu/antrea/pkg/agent"
	"github.com/vmware-tanzu/antrea/pkg/apis/controlplane/v1beta2"
	"github.com/vmware-tanzu/antrea/pkg/querier"
	"k8s.io/klog"

	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/watch"
	clientset "k8s.io/client-go/kubernetes"
	"k8s.io/client-go/util/workqueue"
)

const (
	// How long to wait before retrying the processing of a network policy change.
	minRetryDelay = 5 * time.Second
	maxRetryDelay = 300 * time.Second
	// Default number of workers processing a rule change.
	defaultWorkers = 4
)

func NewEgressController(
	kubeClient clientset.Interface,
	nodeName string,
	antreaClientGetter agent.AntreaClientProvider,

) *Controller {
	c := &Controller{
		kubeClient:           kubeClient,
		antreaClientProvider: antreaClientGetter,
	}
	// Use nodeName to filter resources when watching resources.
	// options := metav1.ListOptions{
	// 	FieldSelector: fields.OneTermEqualSelector("nodeName", nodeName).String(),
	// }
	c.fullSyncGroup.Add(3)
	// options := metav1.ListOptions{
	// 	FieldSelector: fields.OneTermEqualSelector("nodeName", nodeName).String(),
	// }
	c.EgressGroupPatchWatcher = &watcher{
		objectType: "EgressGroup",
		watchFunc: func() (watch.Interface, error) {
			antreaClient, err := c.antreaClientProvider.GetAntreaClient()
			if err != nil {
				klog.Infof("%#v", err)
				return nil, err
			}
			tt, err := antreaClient.ControlplaneV1beta2().EgressPolicies().Watch(context.TODO(), metav1.ListOptions{})
			if err != nil {
				klog.Infof("%#v", err)
				return nil, err
			}
			return tt, err
		}, AddFunc: func(obj runtime.Object) error {
			policy, ok := obj.(*v1beta2.EgressGroup)

			if !ok {
				return fmt.Errorf("cannot convert to *v1beta1.EgressPolicy: %v", obj)
			}
			klog.Infof("EgressPolicy %#v\n applied to Pods on this Node", policy)
			return nil
		},
		UpdateFunc: func(obj runtime.Object) error {
			_, ok := obj.(*v1beta2.EgressPolicy)
			if !ok {
				return fmt.Errorf("cannot convert to *v1beta1.EgressPolicy: %v", obj)
			}
			return nil
		},
		DeleteFunc: func(obj runtime.Object) error {
			policy, ok := obj.(*v1beta2.EgressPolicy)
			if !ok {
				return fmt.Errorf("cannot convert to *v1beta1.EgressPolicy: %v", obj)
			}
			klog.Infof("EgressPolicy %#v\n no longer applied to Pods on this Node", policy)
			return nil
		},
		ReplaceFunc: func(objs []runtime.Object) error {
			policies := make([]*v1beta2.EgressPolicy, len(objs))
			var ok bool
			for i := range objs {
				policies[i], ok = objs[i].(*v1beta2.EgressPolicy)
				if !ok {
					return fmt.Errorf("cannot convert to *v1beta1.EgressPolicy: %v", objs[i])
				}

			}
			return nil
		},
		fullSyncWaitGroup: &c.fullSyncGroup,
		fullSynced:        false,
	}
	c.egressPolicyWatcher = &watcher{
		objectType: "EgressPolicy",
		watchFunc: func() (watch.Interface, error) {
			antreaClient, err := c.antreaClientProvider.GetAntreaClient()
			if err != nil {
				klog.Infof("%#v", err)
				return nil, err
			}
			// tt1, err1 := antreaClient.ControlplaneV1beta2().EgressPolicies().List(context.TODO(), metav1.ListOptions{})
			// if err1 != nil {
			// 	klog.Infof("%#v", err1)
			// 	return nil, err1
			// }
			// klog.Infof("%#v", tt1)

			tt, err := antreaClient.ControlplaneV1beta2().EgressPolicies().Watch(context.TODO(), metav1.ListOptions{})
			if err != nil {
				klog.Infof("%#v", err)
				return nil, err
			}
			return tt, err
		},
		AddFunc: func(obj runtime.Object) error {
			policy, ok := obj.(*v1beta2.EgressPolicy)

			if !ok {
				return fmt.Errorf("cannot convert to *v1beta1.EgressPolicy: %v", obj)
			}
			klog.Infof("EgressPolicy %#v\n applied to Pods on this Node", policy)
			return nil
		},
		UpdateFunc: func(obj runtime.Object) error {
			_, ok := obj.(*v1beta2.EgressPolicy)
			if !ok {
				return fmt.Errorf("cannot convert to *v1beta1.EgressPolicy: %v", obj)
			}
			return nil
		},
		DeleteFunc: func(obj runtime.Object) error {
			policy, ok := obj.(*v1beta2.EgressPolicy)
			if !ok {
				return fmt.Errorf("cannot convert to *v1beta1.EgressPolicy: %v", obj)
			}
			klog.Infof("EgressPolicy %#v\n no longer applied to Pods on this Node", policy)
			return nil
		},
		ReplaceFunc: func(objs []runtime.Object) error {
			policies := make([]*v1beta2.EgressPolicy, len(objs))
			var ok bool
			for i := range objs {
				policies[i], ok = objs[i].(*v1beta2.EgressPolicy)
				if !ok {
					return fmt.Errorf("cannot convert to *v1beta1.EgressPolicy: %v", objs[i])
				}

			}
			return nil
		},
		fullSyncWaitGroup: &c.fullSyncGroup,
		fullSynced:        false,
	}
	return c
}

func (c *Controller) Run(stopCh <-chan struct{}) {
	attempts := 0
	if err := wait.PollImmediateUntil(200*time.Millisecond, func() (bool, error) {
		if attempts%10 == 0 {
			klog.Info("Waiting for Antrea client to be ready")
		}
		if _, err := c.antreaClientProvider.GetAntreaClient(); err != nil {
			attempts++
			return false, nil
		}
		return true, nil
	}, stopCh); err != nil {
		klog.Info("Stopped waiting for Antrea client")
		return
	}
	klog.Infof("ssssssssssssss")
	ac, _ := c.antreaClientProvider.GetAntreaClient()

	egressPolicy := &v1beta2.EgressPolicy{
		EgressGroup: "fwf",
		EgressIP:    "1.1.1.1",
	}
	go func() {
		for {
			time.Sleep(10 * time.Second)
			t1, err1 := ac.ControlplaneV1beta2().EgressPolicies().Create(context.TODO(), egressPolicy, metav1.CreateOptions{})
			if err1 != nil {
				klog.Infof("%#v", err1)
			} else {
				klog.Infof("%#v", t1)
			}
			t, err := ac.ControlplaneV1beta2().EgressPolicies().List(context.TODO(), metav1.ListOptions{})
			if err != nil {
				klog.Infof("%#v", err)
			} else {
				klog.Infof("%#v", t)
			}
		}
	}()
	klog.Info("Antrea client is ready")
	go wait.NonSlidingUntil(c.egressPolicyWatcher.watch, 5*time.Second, stopCh)
	klog.Infof("Waiting for all watchers to complete full sync")
	c.fullSyncGroup.Wait()
	klog.Infof("All watchers have completed full sync, installing flows for init events")
	// Batch install all rules in queue after fullSync is finished.
	for i := 0; i < defaultWorkers; i++ {
		go wait.Until(c.worker, time.Second, stopCh)
	}
	<-stopCh
}

func (c *Controller) enqueueRule(ruleID string) {
	c.queue.Add(ruleID)
}

// worker runs a worker thread that just dequeues items, processes them, and
// marks them done. You may run as many of these in parallel as you wish; the
// workqueue guarantees that they will not end up processing the same rule at
// the same time.
func (c *Controller) worker() {
	for c.processNextWorkItem() {
	}
}

func (c *Controller) processNextWorkItem() bool {
	key, quit := c.queue.Get()
	if quit {
		return false
	}
	klog.Infof("%+v", key)
	defer c.queue.Done(key)
	return true
}

func (w *watcher) setConnected(connected bool) {
	w.lock.Lock()
	defer w.lock.Unlock()
	w.connected = connected
}

func (w *watcher) watch() {
	klog.Infof("Starting watch for %s", w.objectType)
	watcher, err := w.watchFunc()
	if err != nil {
		klog.Warningf("Failed to start watch for %s: %v", w.objectType, err)
		return
	}

	klog.Infof("Started watch for %s", w.objectType)
	w.setConnected(true)
	eventCount := 0
	defer func() {
		klog.Infof("Stopped watch for %s, total items received: %d", w.objectType, eventCount)
		w.setConnected(false)
		watcher.Stop()
	}()

	// First receive init events from the result channel and buffer them until
	// a Bookmark event is received, indicating that all init events have been
	// received.
	var initObjects []runtime.Object
loop:
	for {
		select {
		case event, ok := <-watcher.ResultChan():
			if !ok {
				klog.Warningf("Result channel for %s was closed", w.objectType)
				return
			}
			switch event.Type {
			case watch.Added:
				klog.Infof("Added %s (%#v)", w.objectType, event.Object)
				initObjects = append(initObjects, event.Object)
			case watch.Bookmark:
				break loop
			}
		}
	}
	klog.Infof("Received %d init events for %s", len(initObjects), w.objectType)

	eventCount += len(initObjects)
	if err := w.ReplaceFunc(initObjects); err != nil {
		klog.Errorf("Failed to handle init events: %v", err)
		return
	}
	if !w.fullSynced {
		w.fullSynced = true
		// Notify fullSyncWaitGroup that all events before bookmark is handled
		w.fullSyncWaitGroup.Done()
	}

	for {
		select {
		case event, ok := <-watcher.ResultChan():
			if !ok {
				return
			}
			switch event.Type {
			case watch.Added:
				if err := w.AddFunc(event.Object); err != nil {
					klog.Errorf("Failed to handle added event: %v", err)
					return
				}
				klog.Infof("Added %s (%#v)", w.objectType, event.Object)
			case watch.Modified:
				if err := w.UpdateFunc(event.Object); err != nil {
					klog.Errorf("Failed to handle modified event: %v", err)
					return
				}
				klog.Infof("Updated %s (%#v)", w.objectType, event.Object)
			case watch.Deleted:
				if err := w.DeleteFunc(event.Object); err != nil {
					klog.Errorf("Failed to handle deleted event: %v", err)
					return
				}
				klog.Infof("Removed %s (%#v)", w.objectType, event.Object)
			default:
				klog.Errorf("Unknown event: %v", event)
				return
			}
			eventCount++
		}
	}
}

func (c *Controller) GetEgressPolicies(npFilter *querier.EgressPolicyQueryFilter) []v1beta2.EgressPolicy {
	s := []v1beta2.EgressPolicy{
		{
			EgressGroup: "wowowo",
			EgressIP:    "2,2,2,2",
		},
	}
	return s
}

type Controller struct {
	kubeClient           clientset.Interface
	antreaClientProvider agent.AntreaClientProvider
	egressPolicyWatcher  *watcher
	EgressGroupPatchWatcher   *watcher
	fullSyncGroup        sync.WaitGroup
	queue                workqueue.RateLimitingInterface
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
