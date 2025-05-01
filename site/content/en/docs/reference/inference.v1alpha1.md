---
title: llmaz inference API
content_type: tool-reference
package: inference.llmaz.io/v1alpha1
auto_generated: true
description: Generated API reference documentation for inference.llmaz.io/v1alpha1.
---


## Resource Types


- [Playground](#inference-llmaz-io-v1alpha1-Playground)
- [Service](#inference-llmaz-io-v1alpha1-Service)
  

## `Playground`     {#inference-llmaz-io-v1alpha1-Playground}


**Appears in:**



<p>Playground is the Schema for the playgrounds API</p>


<table class="table">
<thead><tr><th width="30%">Field</th><th>Description</th></tr></thead>
<tbody>
    
<tr><td><code>apiVersion</code><br/>string</td><td><code>inference.llmaz.io/v1alpha1</code></td></tr>
<tr><td><code>kind</code><br/>string</td><td><code>Playground</code></td></tr>
    
  
<tr><td><code>spec</code> <B>[Required]</B><br/>
<a href="#inference-llmaz-io-v1alpha1-PlaygroundSpec"><code>PlaygroundSpec</code></a>
</td>
<td>
   <span class="text-muted">No description provided.</span></td>
</tr>
<tr><td><code>status</code> <B>[Required]</B><br/>
<a href="#inference-llmaz-io-v1alpha1-PlaygroundStatus"><code>PlaygroundStatus</code></a>
</td>
<td>
   <span class="text-muted">No description provided.</span></td>
</tr>
</tbody>
</table>

## `Service`     {#inference-llmaz-io-v1alpha1-Service}


**Appears in:**



<p>Service is the Schema for the services API</p>


<table class="table">
<thead><tr><th width="30%">Field</th><th>Description</th></tr></thead>
<tbody>
    
<tr><td><code>apiVersion</code><br/>string</td><td><code>inference.llmaz.io/v1alpha1</code></td></tr>
<tr><td><code>kind</code><br/>string</td><td><code>Service</code></td></tr>
    
  
<tr><td><code>spec</code> <B>[Required]</B><br/>
<a href="#inference-llmaz-io-v1alpha1-ServiceSpec"><code>ServiceSpec</code></a>
</td>
<td>
   <span class="text-muted">No description provided.</span></td>
</tr>
<tr><td><code>status</code> <B>[Required]</B><br/>
<a href="#inference-llmaz-io-v1alpha1-ServiceStatus"><code>ServiceStatus</code></a>
</td>
<td>
   <span class="text-muted">No description provided.</span></td>
</tr>
</tbody>
</table>

## `BackendName`     {#inference-llmaz-io-v1alpha1-BackendName}

(Alias of `string`)

**Appears in:**

- [BackendRuntimeConfig](#inference-llmaz-io-v1alpha1-BackendRuntimeConfig)





## `BackendRuntime`     {#inference-llmaz-io-v1alpha1-BackendRuntime}


**Appears in:**



<p>BackendRuntime is the Schema for the backendRuntime API</p>


<table class="table">
<thead><tr><th width="30%">Field</th><th>Description</th></tr></thead>
<tbody>
    
  
<tr><td><code>spec</code> <B>[Required]</B><br/>
<a href="#inference-llmaz-io-v1alpha1-BackendRuntimeSpec"><code>BackendRuntimeSpec</code></a>
</td>
<td>
   <span class="text-muted">No description provided.</span></td>
</tr>
<tr><td><code>status</code> <B>[Required]</B><br/>
<a href="#inference-llmaz-io-v1alpha1-BackendRuntimeStatus"><code>BackendRuntimeStatus</code></a>
</td>
<td>
   <span class="text-muted">No description provided.</span></td>
</tr>
</tbody>
</table>

## `BackendRuntimeConfig`     {#inference-llmaz-io-v1alpha1-BackendRuntimeConfig}


**Appears in:**

- [PlaygroundSpec](#inference-llmaz-io-v1alpha1-PlaygroundSpec)



<table class="table">
<thead><tr><th width="30%">Field</th><th>Description</th></tr></thead>
<tbody>
    
  
<tr><td><code>backendName</code><br/>
<a href="#inference-llmaz-io-v1alpha1-BackendName"><code>BackendName</code></a>
</td>
<td>
   <p>BackendName represents the inference backend under the hood, e.g. vLLM.</p>
</td>
</tr>
<tr><td><code>version</code><br/>
<code>string</code>
</td>
<td>
   <p>Version represents the backend version if you want a different one
from the default version.</p>
</td>
</tr>
<tr><td><code>envs</code><br/>
<a href="https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.28/#envvar-v1-core"><code>[]k8s.io/api/core/v1.EnvVar</code></a>
</td>
<td>
   <p>Envs represents the environments set to the container.</p>
</td>
</tr>
<tr><td><code>configName</code> <B>[Required]</B><br/>
<code>string</code>
</td>
<td>
   <p>ConfigName represents the recommended configuration name for the backend,
It will be inferred from the models in the runtime if not specified, e.g. default,
speculative-decoding.</p>
</td>
</tr>
<tr><td><code>args</code><br/>
<code>[]string</code>
</td>
<td>
   <p>Args defined here will &quot;append&quot; the args defined in the recommendedConfig,
either explicitly configured in configName or inferred in the runtime.</p>
</td>
</tr>
<tr><td><code>resources</code><br/>
<a href="#inference-llmaz-io-v1alpha1-ResourceRequirements"><code>ResourceRequirements</code></a>
</td>
<td>
   <p>Resources represents the resource requirements for backend, like cpu/mem,
accelerators like GPU should not be defined here, but at the model flavors,
or the values here will be overwritten.
Resources defined here will &quot;overwrite&quot; the resources in the recommendedConfig.</p>
</td>
</tr>
<tr><td><code>sharedMemorySize</code><br/>
<a href="https://pkg.go.dev/k8s.io/apimachinery/pkg/api/resource#Quantity"><code>k8s.io/apimachinery/pkg/api/resource.Quantity</code></a>
</td>
<td>
   <p>SharedMemorySize represents the size of /dev/shm required in the runtime of
inference workload.
SharedMemorySize defined here will &quot;overwrite&quot; the sharedMemorySize in the recommendedConfig.</p>
</td>
</tr>
</tbody>
</table>

## `BackendRuntimeSpec`     {#inference-llmaz-io-v1alpha1-BackendRuntimeSpec}


**Appears in:**

- [BackendRuntime](#inference-llmaz-io-v1alpha1-BackendRuntime)


<p>BackendRuntimeSpec defines the desired state of BackendRuntime</p>


<table class="table">
<thead><tr><th width="30%">Field</th><th>Description</th></tr></thead>
<tbody>
    
  
<tr><td><code>command</code><br/>
<code>[]string</code>
</td>
<td>
   <p>Command represents the default command for the backendRuntime.</p>
</td>
</tr>
<tr><td><code>image</code> <B>[Required]</B><br/>
<code>string</code>
</td>
<td>
   <p>Image represents the default image registry of the backendRuntime.
It will work together with version to make up a real image.</p>
</td>
</tr>
<tr><td><code>version</code> <B>[Required]</B><br/>
<code>string</code>
</td>
<td>
   <p>Version represents the default version of the backendRuntime.
It will be appended to the image as a tag.</p>
</td>
</tr>
<tr><td><code>envs</code><br/>
<a href="https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.28/#envvar-v1-core"><code>[]k8s.io/api/core/v1.EnvVar</code></a>
</td>
<td>
   <p>Envs represents the environments set to the container.</p>
</td>
</tr>
<tr><td><code>lifecycle</code><br/>
<a href="https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.28/#lifecycle-v1-core"><code>k8s.io/api/core/v1.Lifecycle</code></a>
</td>
<td>
   <p>Lifecycle represents hooks executed during the lifecycle of the container.</p>
</td>
</tr>
<tr><td><code>livenessProbe</code><br/>
<a href="https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.28/#probe-v1-core"><code>k8s.io/api/core/v1.Probe</code></a>
</td>
<td>
   <p>Periodic probe of backend liveness.
Backend will be restarted if the probe fails.
Cannot be updated.</p>
</td>
</tr>
<tr><td><code>readinessProbe</code><br/>
<a href="https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.28/#probe-v1-core"><code>k8s.io/api/core/v1.Probe</code></a>
</td>
<td>
   <p>Periodic probe of backend readiness.
Backend will be removed from service endpoints if the probe fails.</p>
</td>
</tr>
<tr><td><code>startupProbe</code><br/>
<a href="https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.28/#probe-v1-core"><code>k8s.io/api/core/v1.Probe</code></a>
</td>
<td>
   <p>StartupProbe indicates that the Backend has successfully initialized.
If specified, no other probes are executed until this completes successfully.
If this probe fails, the backend will be restarted, just as if the livenessProbe failed.
This can be used to provide different probe parameters at the beginning of a backend's lifecycle,
when it might take a long time to load data or warm a cache, than during steady-state operation.</p>
</td>
</tr>
<tr><td><code>recommendedConfigs</code><br/>
<a href="#inference-llmaz-io-v1alpha1-RecommendedConfig"><code>[]RecommendedConfig</code></a>
</td>
<td>
   <p>RecommendedConfigs represents the recommended configurations for the backendRuntime.</p>
</td>
</tr>
</tbody>
</table>

## `BackendRuntimeStatus`     {#inference-llmaz-io-v1alpha1-BackendRuntimeStatus}


**Appears in:**

- [BackendRuntime](#inference-llmaz-io-v1alpha1-BackendRuntime)


<p>BackendRuntimeStatus defines the observed state of BackendRuntime</p>


<table class="table">
<thead><tr><th width="30%">Field</th><th>Description</th></tr></thead>
<tbody>
    
  
<tr><td><code>conditions</code> <B>[Required]</B><br/>
<a href="https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.28/#condition-v1-meta"><code>[]k8s.io/apimachinery/pkg/apis/meta/v1.Condition</code></a>
</td>
<td>
   <p>Conditions represents the Inference condition.</p>
</td>
</tr>
</tbody>
</table>

## `ElasticConfig`     {#inference-llmaz-io-v1alpha1-ElasticConfig}


**Appears in:**

- [PlaygroundSpec](#inference-llmaz-io-v1alpha1-PlaygroundSpec)



<table class="table">
<thead><tr><th width="30%">Field</th><th>Description</th></tr></thead>
<tbody>
    
  
<tr><td><code>minReplicas</code><br/>
<code>int32</code>
</td>
<td>
   <p>MinReplicas indicates the minimum number of inference workloads based on the traffic.
Default to 1.
MinReplicas couldn't be 0 now, will support serverless in the future.</p>
</td>
</tr>
<tr><td><code>maxReplicas</code><br/>
<code>int32</code>
</td>
<td>
   <p>MaxReplicas indicates the maximum number of inference workloads based on the traffic.
Default to nil means there's no limit for the instance number.</p>
</td>
</tr>
<tr><td><code>scaleTrigger</code><br/>
<a href="#inference-llmaz-io-v1alpha1-ScaleTrigger"><code>ScaleTrigger</code></a>
</td>
<td>
   <p>ScaleTrigger defines the rules to scale the workloads.
Only one trigger cloud work at a time, mostly used in Playground.
ScaleTrigger defined here will &quot;overwrite&quot; the scaleTrigger in the recommendedConfig.</p>
</td>
</tr>
</tbody>
</table>

## `HPATrigger`     {#inference-llmaz-io-v1alpha1-HPATrigger}


**Appears in:**

- [ScaleTrigger](#inference-llmaz-io-v1alpha1-ScaleTrigger)


<p>HPATrigger represents the configuration of the HorizontalPodAutoscaler.
Inspired by kubernetes.io/pkg/apis/autoscaling/types.go#HorizontalPodAutoscalerSpec.
Note: HPA component should be installed in prior.</p>


<table class="table">
<thead><tr><th width="30%">Field</th><th>Description</th></tr></thead>
<tbody>
    
  
<tr><td><code>metrics</code><br/>
<a href="https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.28/#metricspec-v2-autoscaling"><code>[]k8s.io/api/autoscaling/v2.MetricSpec</code></a>
</td>
<td>
   <p>metrics contains the specifications for which to use to calculate the
desired replica count (the maximum replica count across all metrics will
be used).  The desired replica count is calculated multiplying the
ratio between the target value and the current value by the current
number of pods.  Ergo, metrics used must decrease as the pod count is
increased, and vice-versa.  See the individual metric source types for
more information about how each type of metric must respond.</p>
</td>
</tr>
<tr><td><code>behavior</code><br/>
<a href="https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.28/#horizontalpodautoscalerbehavior-v2-autoscaling"><code>k8s.io/api/autoscaling/v2.HorizontalPodAutoscalerBehavior</code></a>
</td>
<td>
   <p>behavior configures the scaling behavior of the target
in both Up and Down directions (scaleUp and scaleDown fields respectively).
If not set, the default HPAScalingRules for scale up and scale down are used.</p>
</td>
</tr>
</tbody>
</table>

## `PlaygroundSpec`     {#inference-llmaz-io-v1alpha1-PlaygroundSpec}


**Appears in:**

- [Playground](#inference-llmaz-io-v1alpha1-Playground)


<p>PlaygroundSpec defines the desired state of Playground</p>


<table class="table">
<thead><tr><th width="30%">Field</th><th>Description</th></tr></thead>
<tbody>
    
  
<tr><td><code>replicas</code><br/>
<code>int32</code>
</td>
<td>
   <p>Replicas represents the replica number of inference workloads.</p>
</td>
</tr>
<tr><td><code>modelClaim</code><br/>
<a href="#llmaz-io-v1alpha1-ModelClaim"><code>ModelClaim</code></a>
</td>
<td>
   <p>ModelClaim represents claiming for one model, it's a simplified use case
of modelClaims. Most of the time, modelClaim is enough.
ModelClaim and modelClaims are exclusive configured.</p>
</td>
</tr>
<tr><td><code>modelClaims</code><br/>
<a href="#llmaz-io-v1alpha1-ModelClaims"><code>ModelClaims</code></a>
</td>
<td>
   <p>ModelClaims represents claiming for multiple models for more complicated
use cases like speculative-decoding.
ModelClaims and modelClaim are exclusive configured.</p>
</td>
</tr>
<tr><td><code>backendRuntimeConfig</code><br/>
<a href="#inference-llmaz-io-v1alpha1-BackendRuntimeConfig"><code>BackendRuntimeConfig</code></a>
</td>
<td>
   <p>BackendRuntimeConfig represents the inference backendRuntime configuration
under the hood, e.g. vLLM, which is the default backendRuntime.</p>
</td>
</tr>
<tr><td><code>elasticConfig</code> <B>[Required]</B><br/>
<a href="#inference-llmaz-io-v1alpha1-ElasticConfig"><code>ElasticConfig</code></a>
</td>
<td>
   <p>ElasticConfig defines the configuration for elastic usage,
e.g. the max/min replicas.</p>
</td>
</tr>
</tbody>
</table>

## `PlaygroundStatus`     {#inference-llmaz-io-v1alpha1-PlaygroundStatus}


**Appears in:**

- [Playground](#inference-llmaz-io-v1alpha1-Playground)


<p>PlaygroundStatus defines the observed state of Playground</p>


<table class="table">
<thead><tr><th width="30%">Field</th><th>Description</th></tr></thead>
<tbody>
    
  
<tr><td><code>conditions</code> <B>[Required]</B><br/>
<a href="https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.28/#condition-v1-meta"><code>[]k8s.io/apimachinery/pkg/apis/meta/v1.Condition</code></a>
</td>
<td>
   <p>Conditions represents the Inference condition.</p>
</td>
</tr>
<tr><td><code>replicas</code> <B>[Required]</B><br/>
<code>int32</code>
</td>
<td>
   <p>Replicas track the replicas that have been created, whether ready or not.</p>
</td>
</tr>
<tr><td><code>selector</code> <B>[Required]</B><br/>
<code>string</code>
</td>
<td>
   <p>Selector points to the string form of a label selector which will be used by HPA.</p>
</td>
</tr>
</tbody>
</table>

## `RecommendedConfig`     {#inference-llmaz-io-v1alpha1-RecommendedConfig}


**Appears in:**

- [BackendRuntimeSpec](#inference-llmaz-io-v1alpha1-BackendRuntimeSpec)


<p>RecommendedConfig represents the recommended configurations for the backendRuntime,
user can choose one of them to apply.</p>


<table class="table">
<thead><tr><th width="30%">Field</th><th>Description</th></tr></thead>
<tbody>
    
  
<tr><td><code>name</code> <B>[Required]</B><br/>
<code>string</code>
</td>
<td>
   <p>Name represents the identifier of the config.</p>
</td>
</tr>
<tr><td><code>args</code><br/>
<code>[]string</code>
</td>
<td>
   <p>Args represents all the arguments for the command.
Argument around with {{ .CONFIG }} is a configuration waiting for render.</p>
</td>
</tr>
<tr><td><code>resources</code><br/>
<a href="#inference-llmaz-io-v1alpha1-ResourceRequirements"><code>ResourceRequirements</code></a>
</td>
<td>
   <p>Resources represents the resource requirements for backend, like cpu/mem,
accelerators like GPU should not be defined here, but at the model flavors,
or the values here will be overwritten.</p>
</td>
</tr>
<tr><td><code>sharedMemorySize</code><br/>
<a href="https://pkg.go.dev/k8s.io/apimachinery/pkg/api/resource#Quantity"><code>k8s.io/apimachinery/pkg/api/resource.Quantity</code></a>
</td>
<td>
   <p>SharedMemorySize represents the size of /dev/shm required in the runtime of
inference workload.</p>
</td>
</tr>
<tr><td><code>scaleTrigger</code><br/>
<a href="#inference-llmaz-io-v1alpha1-ScaleTrigger"><code>ScaleTrigger</code></a>
</td>
<td>
   <p>ScaleTrigger defines the rules to scale the workloads.
Only one trigger cloud work at a time.</p>
</td>
</tr>
</tbody>
</table>

## `ResourceRequirements`     {#inference-llmaz-io-v1alpha1-ResourceRequirements}


**Appears in:**

- [BackendRuntimeConfig](#inference-llmaz-io-v1alpha1-BackendRuntimeConfig)

- [RecommendedConfig](#inference-llmaz-io-v1alpha1-RecommendedConfig)


<p>TODO: Do not support DRA yet, we can support that once needed.</p>


<table class="table">
<thead><tr><th width="30%">Field</th><th>Description</th></tr></thead>
<tbody>
    
  
<tr><td><code>limits</code><br/>
<a href="https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.28/#resourcelist-v1-core"><code>k8s.io/api/core/v1.ResourceList</code></a>
</td>
<td>
   <p>Limits describes the maximum amount of compute resources allowed.
More info: https://kubernetes.io/docs/concepts/configuration/manage-resources-containers/</p>
</td>
</tr>
<tr><td><code>requests</code><br/>
<a href="https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.28/#resourcelist-v1-core"><code>k8s.io/api/core/v1.ResourceList</code></a>
</td>
<td>
   <p>Requests describes the minimum amount of compute resources required.
If Requests is omitted for a container, it defaults to Limits if that is explicitly specified,
otherwise to an implementation-defined value. Requests cannot exceed Limits.
More info: https://kubernetes.io/docs/concepts/configuration/manage-resources-containers/</p>
</td>
</tr>
</tbody>
</table>

## `ScaleTrigger`     {#inference-llmaz-io-v1alpha1-ScaleTrigger}


**Appears in:**

- [ElasticConfig](#inference-llmaz-io-v1alpha1-ElasticConfig)

- [RecommendedConfig](#inference-llmaz-io-v1alpha1-RecommendedConfig)


<p>ScaleTrigger defines the rules to scale the workloads.
Only one trigger cloud work at a time, mostly used in Playground.</p>


<table class="table">
<thead><tr><th width="30%">Field</th><th>Description</th></tr></thead>
<tbody>
    
  
<tr><td><code>hpa</code> <B>[Required]</B><br/>
<a href="#inference-llmaz-io-v1alpha1-HPATrigger"><code>HPATrigger</code></a>
</td>
<td>
   <p>HPA represents the trigger configuration of the HorizontalPodAutoscaler.</p>
</td>
</tr>
</tbody>
</table>

## `ServiceSpec`     {#inference-llmaz-io-v1alpha1-ServiceSpec}


**Appears in:**

- [Service](#inference-llmaz-io-v1alpha1-Service)


<p>ServiceSpec defines the desired state of Service.
Service controller will maintain multi-flavor of workloads with
different accelerators for cost or performance considerations.</p>


<table class="table">
<thead><tr><th width="30%">Field</th><th>Description</th></tr></thead>
<tbody>
    
  
<tr><td><code>modelClaims</code> <B>[Required]</B><br/>
<a href="#llmaz-io-v1alpha1-ModelClaims"><code>ModelClaims</code></a>
</td>
<td>
   <p>ModelClaims represents multiple claims for different models.</p>
</td>
</tr>
<tr><td><code>replicas</code><br/>
<code>int32</code>
</td>
<td>
   <p>Replicas represents the replica number of inference workloads.</p>
</td>
</tr>
<tr><td><code>workloadTemplate</code> <B>[Required]</B><br/>
<code>sigs.k8s.io/lws/api/leaderworkerset/v1.LeaderWorkerTemplate</code>
</td>
<td>
   <p>WorkloadTemplate defines the template for leader/worker pods</p>
</td>
</tr>
<tr><td><code>rolloutStrategy</code><br/>
<code>sigs.k8s.io/lws/api/leaderworkerset/v1.RolloutStrategy</code>
</td>
<td>
   <p>RolloutStrategy defines the strategy that will be applied to update replicas
when a revision is made to the leaderWorkerTemplate.</p>
</td>
</tr>
</tbody>
</table>

## `ServiceStatus`     {#inference-llmaz-io-v1alpha1-ServiceStatus}


**Appears in:**

- [Service](#inference-llmaz-io-v1alpha1-Service)


<p>ServiceStatus defines the observed state of Service</p>


<table class="table">
<thead><tr><th width="30%">Field</th><th>Description</th></tr></thead>
<tbody>
    
  
<tr><td><code>conditions</code> <B>[Required]</B><br/>
<a href="https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.28/#condition-v1-meta"><code>[]k8s.io/apimachinery/pkg/apis/meta/v1.Condition</code></a>
</td>
<td>
   <p>Conditions represents the Inference condition.</p>
</td>
</tr>
<tr><td><code>replicas</code> <B>[Required]</B><br/>
<code>int32</code>
</td>
<td>
   <p>Replicas track the replicas that have been created, whether ready or not.</p>
</td>
</tr>
<tr><td><code>selector</code> <B>[Required]</B><br/>
<code>string</code>
</td>
<td>
   <p>Selector points to the string form of a label selector, the HPA will be
able to autoscale your resource.</p>
</td>
</tr>
</tbody>
</table>
  