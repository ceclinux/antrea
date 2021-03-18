// Copyright 2019 Antrea Authors
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package egresspolicy

import (
	"context"

	"github.com/vmware-tanzu/antrea/pkg/apiserver/registry/networkpolicy"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/apis/meta/internalversion"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/watch"
	"k8s.io/apiserver/pkg/registry/rest"

	"k8s.io/klog"

	"github.com/vmware-tanzu/antrea/pkg/apis/controlplane"
	"github.com/vmware-tanzu/antrea/pkg/apiserver/storage"
	"github.com/vmware-tanzu/antrea/pkg/controller/networkpolicy/store"
	"github.com/vmware-tanzu/antrea/pkg/controller/types"
)

// REST implements rest.Storage for EgressPolicies.
type REST struct {
	egressPolicyStore storage.Interface
}

var (
	_ rest.Storage = &REST{}
	_ rest.Watcher = &REST{}
	_ rest.Scoper  = &REST{}
	_ rest.Lister  = &REST{}
	_ rest.Getter  = &REST{}
)

// NewREST returns a REST object that will work against API services.
func NewREST(egressPolicyStore storage.Interface) *REST {
	return &REST{egressPolicyStore}
}

func (r *REST) New() runtime.Object {
	return &controlplane.EgressPolicy{}
}

func (r *REST) NewList() runtime.Object {
	return &controlplane.EgressPolicyList{}
}

func (r *REST) Get(ctx context.Context, name string, options *metav1.GetOptions) (runtime.Object, error) {

	egressPolicy, exists, err := r.egressPolicyStore.Get(name)
	if err != nil {
		klog.Info("%#v", err)
		return nil, errors.NewInternalError(err)
	}
	if !exists {
		klog.Info("not exsitssss")

		return nil, errors.NewNotFound(controlplane.Resource("egresspolicy"), name)
	}
	obj := new(controlplane.EgressPolicy)
	store.ToEgressPolicyMsg(egressPolicy.(*types.EgressPolicy), obj)
	return obj, nil
}

func (r *REST) List(ctx context.Context, options *internalversion.ListOptions) (runtime.Object, error) {
	labelSelector := labels.Everything()
	if options != nil && options.LabelSelector != nil {
		labelSelector = options.LabelSelector
	}
	egressPolicies := r.egressPolicyStore.List()

	items := make([]controlplane.EgressPolicy, 0, len(egressPolicies))
	for i := range egressPolicies {
		var item controlplane.EgressPolicy
		store.ToEgressPolicyMsg(egressPolicies[i].(*types.EgressPolicy), &item)
		if labelSelector.Matches(labels.Set(item.Labels)) {
			items = append(items, item)
		}
	}
	list := &controlplane.EgressPolicyList{Items: items}
	return list, nil
}

func (r *REST) NamespaceScoped() bool {
	return false
}

func (r *REST) Create(ctx context.Context, obj runtime.Object, createValidation rest.ValidateObjectFunc, options *v1.CreateOptions) (runtime.Object, error) {
	klog.Infof("creating")
	t := &types.EgressPolicy{
		EgressGroup: "lol",
		EgressIP:    "192.168.1.1",
	}
	err := r.egressPolicyStore.Create(t)
	if err != nil {
		klog.Info("%#v", err)
	}
	egressPolicies := r.egressPolicyStore.List()
	klog.Info("%#v", egressPolicies)
	return &controlplane.EgressPolicy{}, nil
}

func (r *REST) Watch(ctx context.Context, options *internalversion.ListOptions) (watch.Interface, error) {
	klog.Infof("wfwfwfwfffewf")
	key, label, field := networkpolicy.GetSelectors(options)
	return r.egressPolicyStore.Watch(ctx, key, label, field)
}

func (r *REST) ConvertToTable(ctx context.Context, obj runtime.Object, tableOptions runtime.Object) (*metav1.Table, error) {
	return rest.NewDefaultTableConvertor(controlplane.Resource("egresspolicy")).ConvertToTable(ctx, obj, tableOptions)
}
