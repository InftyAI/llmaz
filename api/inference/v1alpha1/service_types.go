/*
Copyright 2024 The InftyAI Team.

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

package v1alpha1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	lws "sigs.k8s.io/lws/api/leaderworkerset/v1"

	coreapi "github.com/inftyai/llmaz/api/core/v1alpha1"
)

const (
	// InferenceServiceFlavorsAnnoKey is the annotation key for the flavors specified
	// in the inference service, the value is a comma-separated list of flavor names.
	InferenceServiceFlavorsAnnoKey = "llmaz.io/inference-service-flavors"
)

// ServiceSpec defines the desired state of Service.
// Service controller will maintain multi-flavor of workloads with
// different accelerators for cost or performance considerations.
type ServiceSpec struct {
	// ModelClaims represents multiple claims for different models.
	ModelClaims coreapi.ModelClaims `json:"modelClaims,omitempty"`
	// Replicas represents the replica number of inference workloads.
	// +kubebuilder:default=1
	// +optional
	Replicas *int32 `json:"replicas,omitempty"`
	// WorkloadTemplate defines the template for leader/worker pods
	WorkloadTemplate lws.LeaderWorkerTemplate `json:"workloadTemplate"`
	// RolloutStrategy defines the strategy that will be applied to update replicas
	// when a revision is made to the leaderWorkerTemplate.
	// +kubebuilder:default:={type: "RollingUpdate", rollingUpdateConfiguration: {"maxUnavailable": 1, "maxSurge": 0}}
	// +optional
	RolloutStrategy *lws.RolloutStrategy `json:"rolloutStrategy,omitempty"`
}

const (
	// ServiceAvailable means the inferenceService is available and all the
	// workloads are running as expected.
	ServiceAvailable = "Available"
	// ServiceProgressing means the inferenceService is progressing now, such as
	// in creation, rolling update or scaling up and down.
	ServiceProgressing = "Progressing"
)

// ServiceStatus defines the observed state of Service
type ServiceStatus struct {
	// Conditions represents the Inference condition.
	Conditions []metav1.Condition `json:"conditions,omitempty"`
	// Replicas track the replicas that have been created, whether ready or not.
	Replicas int32 `json:"replicas"`
	// Selector points to the string form of a label selector, the HPA will be
	// able to autoscale your resource.
	Selector string `json:"selector,omitempty"`
}

//+genclient
//+kubebuilder:object:root=true
//+kubebuilder:subresource:status
//+kubebuilder:resource:shortName={isvc}
//+kubebuilder:subresource:scale:specpath=.spec.replicas,statuspath=.status.replicas,selectorpath=.status.selector
//+kubebuilder:printcolumn:name="NAME",type=string,JSONPath=`.metadata.name`,description="Name of the Inference Service"
//+kubebuilder:printcolumn:name="REPLICAS",type=integer,JSONPath=`.status.replicas`,description="Current number of replicas"
//+kubebuilder:printcolumn:name="STATUS",type=string,JSONPath=`.status.conditions[?(@.type=='Available')].reason`,description="Current status (Available/Progressing)"
//+kubebuilder:printcolumn:name="AGE",type=date,JSONPath=`.metadata.creationTimestamp`,description="Time since creation"

// Service is the Schema for the services API
type Service struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   ServiceSpec   `json:"spec,omitempty"`
	Status ServiceStatus `json:"status,omitempty"`
}

//+kubebuilder:object:root=true

// ServiceList contains a list of Service
type ServiceList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []Service `json:"items"`
}

func init() {
	SchemeBuilder.Register(&Service{}, &ServiceList{})
}
