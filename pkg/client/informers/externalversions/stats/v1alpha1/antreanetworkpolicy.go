package v1alpha1

import (
	"context"
	time "time"

	v1alpha1 "github.com/ceclinux/antrea/pkg/client/listers/stats/v1alpha1"
	statsv1alpha1 "github.com/vmware-tanzu/antrea/pkg/apis/stats/v1alpha1"
	versioned "github.com/vmware-tanzu/antrea/pkg/client/clientset/versioned"
	internalinterfaces "github.com/vmware-tanzu/antrea/pkg/client/informers/externalversions/internalinterfaces"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	runtime "k8s.io/apimachinery/pkg/runtime"
	watch "k8s.io/apimachinery/pkg/watch"
	cache "k8s.io/client-go/tools/cache"
)

// AntreaNetworkPolicyInformer provides access to a shared informer and lister for
// NetworkPolicies.
type AntreaNetworkPolicyStatsInformer interface {
	Informer() cache.SharedIndexInformer
	Lister() v1alpha1.NetworkPolicyLister
}

type antreaNetworkPolicyStatsInformer struct {
	factory          internalinterfaces.SharedInformerFactory
	tweakListOptions internalinterfaces.TweakListOptionsFunc
	namespace        string
}

// AntreaNetworkPolicyStatsInformer constructs a new informer for NetworkPolicy type.
// Always prefer using an informer factory to get a shared informer instead of getting an independent
// one. This reduces memory footprint and number of connections to the server.
func NewAntreaNetworkPolicyStatsInformer(client versioned.Interface, namespace string, resyncPeriod time.Duration, indexers cache.Indexers) cache.SharedIndexInformer {
	return NewFilteredAntreaNetworkPolicyStatsInformer(client, namespace, resyncPeriod, indexers, nil)
}

// NewFilteredAntreaNetworkPolicyStatsInformer constructs a new informer for NetworkPolicy type.
// Always prefer using an informer factory to get a shared informer instead of getting an independent
// one. This reduces memory footprint and number of connections to the server.
func NewFilteredAntreaNetworkPolicyStatsInformer(client versioned.Interface, namespace string, resyncPeriod time.Duration, indexers cache.Indexers, tweakListOptions internalinterfaces.TweakListOptionsFunc) cache.SharedIndexInformer {
	return cache.NewSharedIndexInformer(
		&cache.ListWatch{
			ListFunc: func(options v1.ListOptions) (runtime.Object, error) {
				if tweakListOptions != nil {
					tweakListOptions(&options)
				}
				return client.StatsV1alpha1().AntreaClusterNetworkPolicyStats().List(context.TODO(), options)
			},
			WatchFunc: func(options v1.ListOptions) (watch.Interface, error) {
				if tweakListOptions != nil {
					tweakListOptions(&options)
				}
				return client.StatsV1alpha1().AntreaClusterNetworkPolicyStats().Watch(context.TODO(), options)
			},
		},
		&statsv1alpha1.AntreaClusterNetworkPolicyStats{},
		resyncPeriod,
		indexers,
	)
}

func (f *antreaNetworkPolicyStatsInformer) defaultInformer(client versioned.Interface, resyncPeriod time.Duration) cache.SharedIndexInformer {
	return NewFilteredAntreaNetworkPolicyStatsInformer(client, f.namespace, resyncPeriod, cache.Indexers{cache.NamespaceIndex: cache.MetaNamespaceIndexFunc}, f.tweakListOptions)
}

func (f *antreaNetworkPolicyStatsInformer) Informer() cache.SharedIndexInformer {
	return f.factory.InformerFor(&statsv1alpha1.AntreaClusterNetworkPolicyStats{}, f.defaultInformer)
}

func (f *antreaNetworkPolicyStatsInformer) Lister() v1alpha1.AntreaNetworkPolicyStatsLister {
	return v1alpha1.NewAntreaNetworkPolicyStatsLister(f.Informer().GetIndexer())
}
