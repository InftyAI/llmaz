/*
Copyright 2025 The InftyAI Team.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package inference

import (
	"context"

	aigv1a1 "github.com/envoyproxy/ai-gateway/api/v1alpha1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

// gateway related utils: currently only support envoy ai gateway

// IsAIGatewayRouteExist check if the AIGatewayRoute exist
func IsAIGatewayRouteExist(ctx context.Context, client client.Client) (bool, error) {
	var route aigv1a1.AIGatewayRoute
	err := client.Get(ctx, types.NamespacedName{
		Name:      "envoy-ai-gateway-basic",
		Namespace: "default",
	}, &route)
	if err != nil {
		return false, err
	}
	return true, nil
}

// create a AIServiceBackend using playground model name
// example like below:
// apiVersion: aigateway.envoyproxy.io/v1alpha1
// kind: AIServiceBackend
// metadata:
//   name: envoy-ai-gateway-llmaz-model-1 # backendRef
//   namespace: default
// spec:
//   schema:
//     name: OpenAI
//   backendRef:
//     name: qwen2-0--5b-lb # model name
//     kind: Service
//     port: 8080
func CreateAIServiceBackend(ctx context.Context, client client.Client, backendRefName, namespace string, port int, schemaName string) error {
	if schemaName == "" {
		schemaName = "OpenAI"
	}
	// create the AIServiceBackend
 	backend := &aigv1a1.AIServiceBackend{
		ObjectMeta: metav1.ObjectMeta{
			Name:      backendRefName,
			Namespace: namespace,
		},
		Spec: aigv1a1.AIServiceBackendSpec{
			Schema: aigv1a1.AIServiceBackendSchema{
				Name: schemaName,
			},
			BackendRef: aigv1a1.AIServiceBackendRef{
				Name: backendRefName,
				Kind: "Service",
				Port: port,
			},
		},
	}
	return client.Create(ctx, backend)
}

// update aigateway.envoyproxy.io/v1alpha1 AIGatewayRoute default `envoy-ai-gateway-basic` spec.rules list
// example like below:
// - matches:
//   - headers:
// 	  - type: Exact
// 	    name: x-ai-eg-model
// 	    value: qwen2-0.5b # model name
//   backendRefs:
//   - name: envoy-ai-gateway-llmaz-model-1 # backendRef
func UpdateAIGatewayRoute(ctx context.Context, client client.Client, backendRefName, namespace, modelName string) error {
	// get the AIGatewayRoute
	var route aigv1a1.AIGatewayRoute
	if err := client.Get(ctx, types.NamespacedName{
		Name:      "envoy-ai-gateway-basic",
		Namespace: namespace,
	}, &route); err != nil {
		return err
	}
	// update the spec.rules list if the rule does not exist
	for _, r := range route.Spec.Rules {
		if len(r.Matches) == 0 {
			continue
		}
		if len(r.BackendRefs) == 0 {
			continue
		}
		if r.Matches[0].Headers[0].Value == modelName && r.BackendRefs[0].Name == backendRefName {
			return nil
		}
	}
	// if the rule does not exist, append it to the spec.rules list
	rule := aigv1a1.AIGatewayRouteRule{
		Matches: []aigv1a1.AIGatewayRouteMatch{
			{
				Headers: []aigv1a1.AIGatewayRouteHeaderMatch{
					{
						Type:  aigv1a1.HeaderMatchTypeExact,
						Name:  "x-ai-eg-model",
						Value: modelName,
					},
				},
			},
		},
		BackendRefs: []aigv1a1.AIGatewayRouteBackendRef{
			{
				Name: backendRefName,
			},
		},
	}
	route.Spec.Rules = append(route.Spec.Rules, rule)
	// update the AIGatewayRoute
	return client.Update(ctx, &route)
}

// delete a rule by modelName/backendRefname from the AIGatewayRoute
func deleteAIGatewayRoute(ctx context.Context, client client.Client, modelName, backendRefName string) error {
	// get the AIGatewayRoute
	var route aigv1a1.AIGatewayRoute
	if err := client.Get(ctx, types.NamespacedName{
		Name:      "envoy-ai-gateway-basic",
		Namespace: "default",
	}, &route); err != nil {
		return err
	}
	// delete the rule by modelName/backendRefname
	var newRules []aigv1a1.AIGatewayRouteRule
	for _, rule := range route.Spec.Rules {
		if len(rule.Matches) == 0 {
			continue
		}
		if len(rule.BackendRefs) == 0 {
			continue
		}
		if rule.Matches[0].Headers[0].Value == modelName && rule.BackendRefs[0].Name == backendRefName {
			continue
		}
		newRules = append(newRules, rule)
	}
	route.Spec.Rules = newRules
	
	// update the AIGatewayRoute
	return client.Update(ctx, &route)
}