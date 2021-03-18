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

package store

import (
	"fmt"
	"reflect"

	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/watch"
	"k8s.io/client-go/tools/cache"

	"github.com/vmware-tanzu/antrea/pkg/apis/controlplane"
	"github.com/vmware-tanzu/antrea/pkg/apiserver/storage"
	"github.com/vmware-tanzu/antrea/pkg/apiserver/storage/ram"
	"github.com/vmware-tanzu/antrea/pkg/controller/types"
)

const (
	EgressIPIndex = "EgressIP"
)

// networkPolicyEvent implements storage.InternalEvent.
type egressPolicyEvent struct {
	// The current version of the stored NetworkPolicy.
	CurrPolicy *types.EgressPolicy
	// The previous version of the stored NetworkPolicy.
	PrevPolicy *types.EgressPolicy
	// The current version of the transferred NetworkPolicy, which will be used in Added and Modified events.
	CurrObject *controlplane.EgressPolicy
	// The previous version of the transferred NetworkPolicy, which will be used in Deleted events.
	// Note that only metadata will be set in Deleted events for efficiency.
	PrevObject *controlplane.EgressPolicy
	// The key of this NetworkPolicy.
	Key             string
	ResourceVersion uint64
}

// ToWatchEvent converts the networkPolicyEvent to *watch.Event based on the provided Selectors. It has the following features:
// 1. Added event will be generated if the Selectors was not interested in the object but is now.
// 2. Modified event will be generated if the Selectors was and is interested in the object.
// 3. Deleted event will be generated if the Selectors was interested in the object but is not now.
func (event *egressPolicyEvent) ToWatchEvent(selectors *storage.Selectors, isInitEvent bool) *watch.Event {
	prevObjSelected, currObjSelected := isSelected(event.Key, event.PrevPolicy, event.CurrPolicy, selectors, isInitEvent)

	switch {
	case !currObjSelected && !prevObjSelected:
		return nil
	case currObjSelected && !prevObjSelected:
		return &watch.Event{Type: watch.Added, Object: event.CurrObject}
	case currObjSelected && prevObjSelected:
		return &watch.Event{Type: watch.Modified, Object: event.CurrObject}
	case !currObjSelected && prevObjSelected:
		return &watch.Event{Type: watch.Deleted, Object: event.PrevObject}
	}
	return nil
}

func (event *egressPolicyEvent) GetResourceVersion() uint64 {
	return event.ResourceVersion
}

var _ storage.GenEventFunc = genEgressPolicyEvent

// genNetworkPolicyEvent generates InternalEvent from the given versions of a NetworkPolicy.
// It converts the stored NetworkPolicy to its message form.
func genEgressPolicyEvent(key string, prevObj, currObj interface{}, rv uint64) (storage.InternalEvent, error) {
	if reflect.DeepEqual(prevObj, currObj) {
		return nil, nil
	}

	event := &egressPolicyEvent{Key: key, ResourceVersion: rv}

	if prevObj != nil {
		event.PrevPolicy = prevObj.(*types.EgressPolicy)
		event.PrevObject = new(controlplane.EgressPolicy)
		ToEgressPolicyMsg(event.PrevPolicy, event.PrevObject)
	}

	if currObj != nil {
		event.CurrPolicy = currObj.(*types.EgressPolicy)
		event.CurrObject = new(controlplane.EgressPolicy)
		ToEgressPolicyMsg(event.CurrPolicy, event.CurrObject)
	}

	return event, nil
}

// ToNetworkPolicyMsg converts the stored NetworkPolicy to its message form.
// If includeBody is true, Rules and AppliedToGroups will be copied.
func ToEgressPolicyMsg(in *types.EgressPolicy, out *controlplane.EgressPolicy) {
	out.EgressGroup = in.EgressGroup
	out.EgressIP = in.EgressIP
}

// NetworkPolicyKeyFunc knows how to get the key of a NetworkPolicy.
func EgressPolicyKeyFunc(obj interface{}) (string, error) {
	policy, ok := obj.(*types.EgressPolicy)
	if !ok {
		return "", fmt.Errorf("object is not *types.EgressPolicy: %v", obj)
	}
	return policy.EgressIP, nil
}

// NewNetworkPolicyStore creates a store of NetworkPolicy.
func NewEgressPolicyStore() storage.Interface {
	// Build indices with the appliedToGroups and the addressGroups so that
	// it's efficient to get network policies that have references to specified
	// appliedToGroups or addressGroups.
	indexers := cache.Indexers{
		EgressIPIndex: func(obj interface{}) ([]string, error) {
			fp, ok := obj.(*types.EgressPolicy)
			if !ok {
				return []string{}, nil
			}
			if len(fp.EgressGroup) == 0 {
				return []string{}, nil
			}
			return []string{fp.EgressGroup}, nil
		},
	}
	return ram.NewStore(EgressPolicyKeyFunc, indexers, genEgressPolicyEvent, keyAndSpanSelectFunc, func() runtime.Object { return new(controlplane.EgressPolicy) })
}
