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

	HUGGING_FACE = "Huggingface"
	MODEL_SCOPE  = "ModelScope"
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
	// Filename refers to a specified model file rather than the whole repo.
	// This is helpful to download a specified GGUF model rather than downloading
	// the whole repo which includes all kinds of quantized models.
	// TODO: this is only supported with Huggingface, add support for ModelScope
	// in the near future.
	Filename *string `json:"filename,omitempty"`
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
	// NodeSelector represents the node candidates for Pod placements, if a node doesn't
	// meet the nodeSelector, it will be filtered out in the resourceFungibility scheduler plugin.
	// If nodeSelector is empty, it means every node is a candidate.
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

// ModelClaim represents claiming for one model, it's the standard claimMode
// of multiModelsClaim compared to other modes like SpeculativeDecoding.
type ModelClaim struct {
	// ModelName represents the name of the Model.
	ModelName ModelName `json:"modelName,omitempty"`
	// InferenceFlavors represents a list of flavors with fungibility support
	// to serve the model.
	// If set, The flavor names should be a subset of the model configured flavors.
	// If not set, Model configured flavors will be used by default.
	// +optional
	InferenceFlavors []FlavorName `json:"inferenceFlavors,omitempty"`
}

type ModelRole string

const (
	// Main represents the main model, if only one model is required,
	// it must be the main model. Only one main model is allowed.
	MainRole ModelRole = "main"
	// Draft represents the draft model in speculative decoding,
	// the main model is the target model then.
	DraftRole ModelRole = "draft"
)

type ModelRepresentative struct {
	// Name represents the model name.
	Name ModelName `json:"name"`
	// Role represents the model role once more than one model is required.
	// +kubebuilder:validation:Enum={main,draft}
	// +kubebuilder:default=main
	// +optional
	Role *ModelRole `json:"role,omitempty"`
}

// ModelClaims represents multiple claims for different models.
type ModelClaims struct {
	// Models represents a list of models with roles specified, there maybe
	// multiple models here to support state-of-the-art technologies like
	// speculative decoding, then one model is main(target) model, another one
	// is draft model.
	// +kubebuilder:validation:MinItems=1
	Models []ModelRepresentative `json:"models,omitempty"`
	// InferenceFlavors represents a list of flavors with fungibility supported
	// to serve the model.
	// - If not set, always apply with the 0-index model by default.
	// - If set, will lookup the flavor names following the model orders.
	// +optional
	InferenceFlavors []FlavorName `json:"inferenceFlavors,omitempty"`
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
	// Flavors are fungible following the priority represented by the slice order.
	// +kubebuilder:validation:MaxItems=8
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
