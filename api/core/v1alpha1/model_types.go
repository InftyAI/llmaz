/*
Copyright 2024.

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
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

const (
	ModelFamilyNameLabelKey = "llmaz.io/model-family-name"
	ModelNameLabelKey       = "llmaz.io/model-name"
)

// ModelHub represents the model registry for model downloads.
type ModelHub struct {
	// Name refers to the model registry, such as huggingface.
	// +kubebuilder:default=Huggingface
	// +kubebuilder:validation:Enum={Huggingface,ModelScope}
	// +optional
	Name *string `json:"name,omitempty"`
	// ModelID refers to the model identifier on model hub,
	// such as meta-llama/Meta-Llama-3-8B.
	ModelID string `json:"modelID,omitempty"`
	// Revision refers to a Git revision id which can be a branch name, a tag, or a commit hash.
	// Most of the time, you don't need to specify it.
	// +optional
	Revision *string `json:"revision,omitempty"`
}

// URIProtocol represents the protocol of the URI.
type URIProtocol string

// Add roles for operating leaderWorkerSet.
//
// +kubebuilder:rbac:groups=leaderworkerset.x-k8s.io,resources=leaderworkersets,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=leaderworkerset.x-k8s.io,resources=leaderworkersets/status,verbs=get;update;patch

// ModelSource represents the source of the model.
// Only one model source will be used.
type ModelSource struct {
	// ModelHub represents the model registry for model downloads.
	// +optional
	ModelHub *ModelHub `json:"modelHub,omitempty"`
	// URI represents a various kinds of model sources following the uri protocol, e.g.:
	// - OSS: oss://<bucket>.<endpoint>/<path-to-your-model>
	//
	// +optional
	URI *URIProtocol `json:"uri,omitempty"`
}

type FlavorName string

// Flavor defines the accelerator requirements for a model and the necessary parameters
// in autoscaling. Right now, it will be used in two places:
// - Pod scheduling with node selectors specified.
// - Cluster autoscaling with essential parameters provided.
type Flavor struct {
	// Name represents the flavor name, which will be used in model claim.
	Name FlavorName `json:"name"`
	// Requests defines the required accelerators to serve the model, like nvidia.com/gpu: 8.
	// When GPU number is greater than 8, like 32, then multi-host inference is enabled and
	// 32/8=4 hosts will be grouped as an unit, each host will have a resource request as
	// nvidia.com/gpu: 8. The may change in the future if the GPU number limit is broken.
	// Not recommended to set the cpu and memory usage here.
	// If using playground, you can define the cpu/mem usage at backendConfig.
	// If using service, you can define the cpu/mem at the container resources.
	// Note: if you define the same accelerator requests at playground/service as well,
	// the requests here will be covered.
	// +optional
	Requests v1.ResourceList `json:"requests,omitempty"`
	// NodeSelector defines the labels to filter specified nodes, like
	// cloud-provider.com/accelerator: nvidia-a100.
	// NodeSelector will be auto injected to the Pods as scheduling primitives.
	// +optional
	NodeSelector map[string]string `json:"nodeSelector,omitempty"`
	// Params stores other useful parameters and will be consumed by the autoscaling components
	// like cluster-autoscaler, Karpenter.
	// E.g. when scaling up nodes with 8x Nvidia A00, the parameter can be injected with
	// instance-type: p4d.24xlarge for AWS.
	// +optional
	Params map[string]string `json:"params,omitempty"`
}

type ModelName string

// ModelClaim represents the references to one model.
// It's a simple config for most of the cases compared to multiModelsClaim.
type ModelClaim struct {
	// ModelName represents a list of models, there maybe multiple models here
	// to support state-of-the-art technologies like speculative decoding.
	ModelName ModelName `json:"modelName,omitempty"`
	// InferenceFlavors represents a list of flavors with fungibility supports
	// to serve the model. The flavor names should be a subset of the model
	// configured flavors. If not set, will use the model configured flavors.
	// +optional
	InferenceFlavors []FlavorName `json:"inferenceFlavors,omitempty"`
}

// MultiModelsClaim represents the references to multiple models.
// It's an advanced and more complicated config comparing to modelClaim.
type MultiModelsClaim struct {
	// ModelNames represents a list of models, there maybe multiple models here
	// to support state-of-the-art technologies like speculative decoding.
	// +kubebuilder:validation:MinItems=1
	ModelNames []ModelName `json:"modelNames,omitempty"`
	// InferenceFlavors represents a list of flavors with fungibility supported
	// to serve the model.
	// - If not set, always apply with the 0-index model by default.
	// - If set, will lookup the flavor names following the model orders.
	// +optional
	InferenceFlavors []FlavorName `json:"inferenceFlavors,omitempty"`
	// Rate works only when multiple claims declared, it represents the replicas rates of
	// the sub-workload, like when claim1.rate:claim2.rate = 1:2 and 3 replicas defined in
	// workload, then sub-workload1 will have 1 replica, and sub-workload2 will have 2 replicas.
	// This is mostly designed for state-of-the-art technology called splitwise, the prefill
	// and decode phase will be separated and requires different accelerators.
	// The sum of the rates should be divisible by replicas.
	Rate *int32 `json:"rate,omitempty"`
}

// ModelSpec defines the desired state of Model
type ModelSpec struct {
	// FamilyName represents the model type, like llama2, which will be auto injected
	// to the labels with the key of `llmaz.io/model-family-name`.
	FamilyName ModelName `json:"familyName"`
	// Source represents the source of the model, there're several ways to load
	// the model such as loading from huggingface, OCI registry, s3, host path and so on.
	Source ModelSource `json:"source"`
	// InferenceFlavors represents the accelerator requirements to serve the model.
	// Flavors are fungible following the priority of slice order.
	// +optional
	InferenceFlavors []Flavor `json:"inferenceFlavors,omitempty"`
}

// ModelStatus defines the observed state of Model
type ModelStatus struct {
	// Conditions represents the Inference condition.
	Conditions []metav1.Condition `json:"conditions,omitempty"`
}

//+genclient
//+kubebuilder:object:root=true
//+kubebuilder:subresource:status
//+kubebuilder:resource:scope=Cluster

// OpenModel is the Schema for the open models API
type OpenModel struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   ModelSpec   `json:"spec,omitempty"`
	Status ModelStatus `json:"status,omitempty"`
}

//+kubebuilder:object:root=true

// OpenModelList contains a list of OpenModel
type OpenModelList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []OpenModel `json:"items"`
}

func init() {
	SchemeBuilder.Register(&OpenModel{}, &OpenModelList{})
}
