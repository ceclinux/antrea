// Copyright 2020 Antrea Authors
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
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strings"

	agentquerier "github.com/vmware-tanzu/antrea/pkg/agent/querier"
	cpv1beta "github.com/vmware-tanzu/antrea/pkg/apis/controlplane/v1beta2"
	"github.com/vmware-tanzu/antrea/pkg/querier"
	"k8s.io/klog"
)

// HandleFunc creates a http.HandlerFunc which uses an AgentNetworkPolicyInfoQuerier
// to query applied to groups in current agent. The HandlerFunc accepts `name` parameter
// in URL and returns the specific applied to group.
func HandleFunc(aq agentquerier.AgentQuerier) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		klog.Infof("eeeeeeeeeeeeeee")
		npFilter, err := newFilterFromURLQuery(r.URL.Query())
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
		}
		npq := aq.GetEgressPolicyInfoQuerier()
		nps := npq.GetEgressPolicies(npFilter)
		var obj interface{}

		obj = cpv1beta.EgressPolicyList{Items: nps}
		if err := json.NewEncoder(w).Encode(obj); err != nil {
			http.Error(w, "Failed to encode response: "+err.Error(), http.StatusInternalServerError)
		}
	}
}

func newFilterFromURLQuery(query url.Values) (*querier.EgressPolicyQueryFilter, error) {
	namespace := query.Get("namespace")
	pod := query.Get("pod")
	if pod != "" && namespace == "" {
		return nil, fmt.Errorf("with a pod name, namespace must be provided")
	}

	strSourceType := strings.ToUpper(query.Get("type"))

	source := query.Get("source")
	name := query.Get("name")
	if name != "" && (source != "" || namespace != "" || pod != "" || strSourceType != "") {
		return nil, fmt.Errorf("with a name, none of the other fields can be set")
	}

	return &querier.EgressPolicyQueryFilter{
		Name:       name,
		SourceName: source,
		Namespace:  namespace,
		Pod:        pod,
	}, nil
}
