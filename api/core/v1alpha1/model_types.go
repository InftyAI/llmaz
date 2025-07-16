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
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

const (
	ModelFamilyNameLabelKey = "llmaz.io/model-family-name"
	ModelNameLabelKey       = "llmaz.io/model-name"
	// Annotation with value = "true" represents we'll preload the model,
	// by default via Manta(https://github.com/InftyAI/Manta), make sure
	// Manta is installed in prior.
	// Note: right now, we only support preloading models from Huggingface,
	// in the future, more hubs and objstores will also be supported.
	//
	// We set this as an annotation rather than a field is just because preheating
	// models is not a common sense and Manta is not a mature solution right now.
	// Once either of them qualified, we'll expose this as a field in Model.
	ModelPreheatAnnoKey = "llmaz.io/model-preheat"

	// ModelActivatorAnnotationKey is used to indicate whether the model is activated by the activator.
	ModelActivatorAnnoKey = "activator.llmaz.io/playground"
	// CachedModelActivatorAnnotationKey is used to cache the activator info of the model.
	CachedModelActivatorAnnoKey = "cached.activator.llmaz.io"

	HUGGING_FACE = "Huggingface"
	MODEL_SCOPE  = "ModelScope"

	DefaultOwnedBy = "llmaz"
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
	// Note: once filename is set, allowPatterns and ignorePatterns should be left unset.
	Filename *string `json:"filename,omitempty"`
	// Revision refers to a Git revision id which can be a branch name, a tag, or a commit hash.
	// +kubebuilder:default=main
	// +optional
	Revision *string `json:"revision,omitempty"`
	// AllowPatterns refers to files matched with at least one pattern will be downloaded.
	// +optional
	AllowPatterns []string `json:"allowPatterns,omitempty"`
	// IgnorePatterns refers to files matched with any of the patterns will not be downloaded.
	// +optional
	IgnorePatterns []string `json:"ignorePatterns,omitempty"`
}

// URIProtocol represents the protocol of the URI.
type URIProtocol string

// ModelSource represents the source of the model.
// Only one model source will be used.
type ModelSource struct {
	// ModelHub represents the model registry for model downloads.
	// +optional
	ModelHub *ModelHub `json:"modelHub,omitempty"`
	// URI represents a various kinds of model sources following the uri protocol, protocol://<address>, e.g.
	// - oss://<bucket>.<endpoint>/<path-to-your-model>
	// - ollama://llama3.3
	// - host://<path-to-your-model>
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
	// Limits defines the required accelerators to serve the model for each replica,
	// like <nvidia.com/gpu: 8>. For multi-hosts cases, the limits here indicates
	// the resource requirements for each replica, usually equals to the TP size.
	// Not recommended to set the cpu and memory usage here:
	// - if using playground, you can define the cpu/mem usage at backendConfig.
	// - if using inference service, you can define the cpu/mem at the container resources.
	// However, if you define the same accelerator resources at playground/service as well,
	// the resources will be overwritten by the flavor limit here.
	// +optional
	Limits v1.ResourceList `json:"limits,omitempty"`
	// NodeSelector represents the node candidates for Pod placements, if a node doesn't
	// meet the nodeSelector, it will be filtered out in the resourceFungibility scheduler plugin.
	// If nodeSelector is empty, it means every node is a candidate.
	// +optional
	NodeSelector map[string]string `json:"nodeSelector,omitempty"`
	// Params stores other useful parameters and will be consumed by cluster-autoscaler / Karpenter
	// for autoscaling or be defined as model parallelism parameters like TP or PP size.
	// E.g. with autoscaling, when scaling up nodes with 8x Nvidia A00, the parameter can be injected
	// with <INSTANCE-TYPE: p4d.24xlarge> for AWS.
	// Preset parameters: TP, PP, INSTANCE-TYPE.
	// +optional
	Params map[string]string `json:"params,omitempty"`
}

// InferenceConfig represents the inference configurations for the model.
type InferenceConfig struct {
	// Flavors represents the accelerator requirements to serve the model.
	// Flavors are fungible following the priority represented by the slice order.
	// +kubebuilder:validation:MaxItems=8
	// +optional
	Flavors []Flavor `json:"flavors,omitempty"`
}

type ModelName string

// ModelClaim represents claiming for one model.
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
	// MainRole represents the main model, if only one model is required,
	// it must be the main model. Only one main model is allowed.
	MainRole ModelRole = "main"
	// DraftRole represents the draft model in speculative decoding,
	// the main model is the target model then.
	DraftRole ModelRole = "draft"
	// LoraRole represents the lora model.
	LoraRole ModelRole = "lora"
)

// ModelRef refers to a created Model with it's role.
type ModelRef struct {
	// Name represents the model name.
	Name ModelName `json:"name"`
	// Role represents the model role once more than one model is required.
	// Such as a draft role, which means running with SpeculativeDecoding,
	// and default arguments for backend will be searched in backendRuntime
	// with the name of speculative-decoding.
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
	Models []ModelRef `json:"models,omitempty"`
	// InferenceFlavors represents a list of flavor names with fungibility supported
	// to serve the model.
	// - If not set, will employ the model configured flavors by default.
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
	// InferenceConfig represents the inference configurations for the model.
	InferenceConfig *InferenceConfig `json:"inferenceConfig,omitempty"`
	// OwnedBy represents the owner of the running models serving by the backends,
	// which will be exported as the field of "OwnedBy" in openai-compatible API "/models".
	// Default to "llmaz" if not set.
	// +optional
	// +kubebuilder:default="llmaz"
	OwnedBy *string `json:"ownedBy,omitempty"`
	// CreatedAt represents the creation timestamp of the running models serving by the backends,
	// which will be exported as the field of "Created" in openai-compatible API "/models".
	// It follows the format of RFC 3339, for example "2024-05-21T10:00:00Z".
	// +optional
	// +kubebuilder:validation:Format=date-time
	CreatedAt *metav1.Time `json:"createdAt,omitempty"`
}

const (
	// ModelPending means model is waiting for model downloading.
	ModelPending = "Pending"
	// ModelReady means model is already downloaded.
	ModelReady = "Ready"
)

// ModelStatus defines the observed state of Model
type ModelStatus struct {
	// Conditions represents the Inference condition.
	Conditions []metav1.Condition `json:"conditions,omitempty"`
}

//+genclient
//+kubebuilder:object:root=true
//+kubebuilder:subresource:status
//+kubebuilder:resource:shortName=om,scope=Cluster
//+kubebuilder:printcolumn:name="OWNEDBY",type=string,JSONPath=`.spec.ownedBy`,description="Owner of the model"
//+kubebuilder:printcolumn:name="AGE",type=date,JSONPath=`.metadata.creationTimestamp`,description="Time since creation"
//+kubebuilder:printcolumn:name="MODELHUB",type=string,JSONPath=`.spec.source.modelHub.name`,description="Model hub name"
//+kubebuilder:printcolumn:name="MODELID",type=string,JSONPath=`.spec.source.modelHub.modelID`,description="Model ID on the model hub"
//+kubebuilder:printcolumn:name="URI",type=string,JSONPath=`.spec.source.uri`,description="URI of the model when using a custom source (e.g., s3://, ollama://)"

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
