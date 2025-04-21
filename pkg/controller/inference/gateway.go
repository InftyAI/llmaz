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
	runtimeClient "sigs.k8s.io/controller-runtime/pkg/client"
	gwapiv1 "sigs.k8s.io/gateway-api/apis/v1"
)

// gateway related utils: currently only support envoy ai gateway

var (
	// TODO make it configurable
	defaultGatewayName = "envoy-ai-gateway-basic"
	// TODO make it configurable
	DefaultGatewayNamespace = "default"
)

// GetDefaultAIGatewayRoute check if the AIGatewayRoute exist and return
// the default one or the first one if multiple AIGatewayRoute exist
func GetDefaultAIGatewayRoute (ctx context.Context, client runtimeClient.Client, namespace string) (bool, string, error) {
	routeList := &aigv1a1.AIGatewayRouteList{}
	opts := []runtimeClient.ListOption{
		runtimeClient.InNamespace(namespace),
	}

	// list the AIGatewayRoute and get `envoy-ai-gateway-basic` object if exists
	err := client.List(ctx, routeList, opts...)
	if err != nil || len(routeList.Items) == 0 {
		return false, "", err
	}
	
	for _, route := range routeList.Items {
		if route.Name == defaultGatewayName {
			return true, defaultGatewayName, nil
		}
	}
	// return the first one
	return true, routeList.Items[0].Name, nil
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
func CreateAIServiceBackend(ctx context.Context, client client.Client, backendRefName, namespace string, port int) error {
	kind := gwapiv1.Kind("Service")
	portName := gwapiv1.PortNumber(port)
	// create the AIServiceBackend
 	backend := &aigv1a1.AIServiceBackend{
		ObjectMeta: metav1.ObjectMeta{
			Name:      backendRefName,
			Namespace: namespace,
		},
		Spec: aigv1a1.AIServiceBackendSpec{
			APISchema: aigv1a1.VersionedAPISchema{
				Name: aigv1a1.APISchemaOpenAI,
			},
			BackendRef: gwapiv1.BackendObjectReference{
				Name: gwapiv1.ObjectName(backendRefName),
				Kind: &kind,
				Port: &portName,
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
func UpdateAIGatewayRoute(ctx context.Context, client client.Client, backendRefName, namespace, modelName, routeName string) error {
	// get the AIGatewayRoute
	var route aigv1a1.AIGatewayRoute
	if err := client.Get(ctx, types.NamespacedName{
		Name:      routeName,
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
	exact := gwapiv1.HeaderMatchExact
	// if the rule does not exist, append it to the spec.rules list
	rule := aigv1a1.AIGatewayRouteRule{
		Matches: []aigv1a1.AIGatewayRouteRuleMatch{
			{
				Headers: []gwapiv1.HTTPHeaderMatch{
					{
						Type:  &exact,
						Name:  "x-ai-eg-model",
						Value: modelName,
					},
				},
			},
		},
		BackendRefs: []aigv1a1.AIGatewayRouteRuleBackendRef{
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
func deleteAIGatewayRoute(ctx context.Context, client client.Client, modelName, backendRefName, namespace, routeName string) error {
	// get the AIGatewayRoute
	var route aigv1a1.AIGatewayRoute
	if err := client.Get(ctx, types.NamespacedName{
		Name:      routeName,
		Namespace: namespace,
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