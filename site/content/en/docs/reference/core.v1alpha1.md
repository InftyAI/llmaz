---
title: llmaz core API
content_type: tool-reference
package: llmaz.io/v1alpha1
auto_generated: true
description: Generated API reference documentation for llmaz.io/v1alpha1.
---


## Resource Types


- [OpenModel](#llmaz-io-v1alpha1-OpenModel)
  

## `OpenModel`     {#llmaz-io-v1alpha1-OpenModel}


**Appears in:**



<p>OpenModel is the Schema for the open models API</p>


<table class="table">
<thead><tr><th width="30%">Field</th><th>Description</th></tr></thead>
<tbody>
    
<tr><td><code>apiVersion</code><br/>string</td><td><code>llmaz.io/v1alpha1</code></td></tr>
<tr><td><code>kind</code><br/>string</td><td><code>OpenModel</code></td></tr>
    
  
<tr><td><code>spec</code> <B>[Required]</B><br/>
<a href="#llmaz-io-v1alpha1-ModelSpec"><code>ModelSpec</code></a>
</td>
<td>
   <span class="text-muted">No description provided.</span></td>
</tr>
<tr><td><code>status</code> <B>[Required]</B><br/>
<a href="#llmaz-io-v1alpha1-ModelStatus"><code>ModelStatus</code></a>
</td>
<td>
   <span class="text-muted">No description provided.</span></td>
</tr>
</tbody>
</table>

## `Flavor`     {#llmaz-io-v1alpha1-Flavor}


**Appears in:**

- [InferenceConfig](#llmaz-io-v1alpha1-InferenceConfig)


<p>Flavor defines the accelerator requirements for a model and the necessary parameters
in autoscaling. Right now, it will be used in two places:</p>
<ul>
<li>Pod scheduling with node selectors specified.</li>
<li>Cluster autoscaling with essential parameters provided.</li>
</ul>


<table class="table">
<thead><tr><th width="30%">Field</th><th>Description</th></tr></thead>
<tbody>
    
  
<tr><td><code>name</code> <B>[Required]</B><br/>
<a href="#llmaz-io-v1alpha1-FlavorName"><code>FlavorName</code></a>
</td>
<td>
   <p>Name represents the flavor name, which will be used in model claim.</p>
</td>
</tr>
<tr><td><code>limits</code><br/>
<a href="https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.28/#resourcelist-v1-core"><code>k8s.io/api/core/v1.ResourceList</code></a>
</td>
<td>
   <p>Limits defines the required accelerators to serve the model for each replica,
like &lt;nvidia.com/gpu: 8&gt;. For multi-hosts cases, the limits here indicates
the resource requirements for each replica, usually equals to the TP size.
Not recommended to set the cpu and memory usage here:</p>
<ul>
<li>if using playground, you can define the cpu/mem usage at backendConfig.</li>
<li>if using inference service, you can define the cpu/mem at the container resources.
However, if you define the same accelerator resources at playground/service as well,
the resources will be overwritten by the flavor limit here.</li>
</ul>
</td>
</tr>
<tr><td><code>nodeSelector</code><br/>
<code>map[string]string</code>
</td>
<td>
   <p>NodeSelector represents the node candidates for Pod placements, if a node doesn't
meet the nodeSelector, it will be filtered out in the resourceFungibility scheduler plugin.
If nodeSelector is empty, it means every node is a candidate.</p>
</td>
</tr>
<tr><td><code>params</code><br/>
<code>map[string]string</code>
</td>
<td>
   <p>Params stores other useful parameters and will be consumed by cluster-autoscaler / Karpenter
for autoscaling or be defined as model parallelism parameters like TP or PP size.
E.g. with autoscaling, when scaling up nodes with 8x Nvidia A00, the parameter can be injected
with &lt;INSTANCE-TYPE: p4d.24xlarge&gt; for AWS.
Preset parameters: TP, PP, INSTANCE-TYPE.</p>
</td>
</tr>
</tbody>
</table>

## `FlavorName`     {#llmaz-io-v1alpha1-FlavorName}

(Alias of `string`)

**Appears in:**

- [Flavor](#llmaz-io-v1alpha1-Flavor)







## `InferenceConfig`     {#llmaz-io-v1alpha1-InferenceConfig}


**Appears in:**

- [ModelSpec](#llmaz-io-v1alpha1-ModelSpec)


<p>InferenceConfig represents the inference configurations for the model.</p>


<table class="table">
<thead><tr><th width="30%">Field</th><th>Description</th></tr></thead>
<tbody>
    
  
<tr><td><code>flavors</code><br/>
<a href="#llmaz-io-v1alpha1-Flavor"><code>[]Flavor</code></a>
</td>
<td>
   <p>Flavors represents the accelerator requirements to serve the model.
Flavors are fungible following the priority represented by the slice order.</p>
</td>
</tr>
</tbody>
</table>

## `ModelHub`     {#llmaz-io-v1alpha1-ModelHub}


**Appears in:**

- [ModelSource](#llmaz-io-v1alpha1-ModelSource)


<p>ModelHub represents the model registry for model downloads.</p>


<table class="table">
<thead><tr><th width="30%">Field</th><th>Description</th></tr></thead>
<tbody>
    
  
<tr><td><code>name</code><br/>
<code>string</code>
</td>
<td>
   <p>Name refers to the model registry, such as huggingface.</p>
</td>
</tr>
<tr><td><code>modelID</code> <B>[Required]</B><br/>
<code>string</code>
</td>
<td>
   <p>ModelID refers to the model identifier on model hub,
such as meta-llama/Meta-Llama-3-8B.</p>
</td>
</tr>
<tr><td><code>filename</code> <B>[Required]</B><br/>
<code>string</code>
</td>
<td>
   <p>Filename refers to a specified model file rather than the whole repo.
This is helpful to download a specified GGUF model rather than downloading
the whole repo which includes all kinds of quantized models.
TODO: this is only supported with Huggingface, add support for ModelScope
in the near future.
Note: once filename is set, allowPatterns and ignorePatterns should be left unset.</p>
</td>
</tr>
<tr><td><code>revision</code><br/>
<code>string</code>
</td>
<td>
   <p>Revision refers to a Git revision id which can be a branch name, a tag, or a commit hash.</p>
</td>
</tr>
<tr><td><code>allowPatterns</code><br/>
<code>[]string</code>
</td>
<td>
   <p>AllowPatterns refers to files matched with at least one pattern will be downloaded.</p>
</td>
</tr>
<tr><td><code>ignorePatterns</code><br/>
<code>[]string</code>
</td>
<td>
   <p>IgnorePatterns refers to files matched with any of the patterns will not be downloaded.</p>
</td>
</tr>
</tbody>
</table>

## `ModelName`     {#llmaz-io-v1alpha1-ModelName}

(Alias of `string`)

**Appears in:**


- [ModelRef](#llmaz-io-v1alpha1-ModelRef)

- [ModelSpec](#llmaz-io-v1alpha1-ModelSpec)





## `ModelRef`     {#llmaz-io-v1alpha1-ModelRef}


**Appears in:**



<p>ModelRef refers to a created Model with it's role.</p>


<table class="table">
<thead><tr><th width="30%">Field</th><th>Description</th></tr></thead>
<tbody>
    
  
<tr><td><code>name</code> <B>[Required]</B><br/>
<a href="#llmaz-io-v1alpha1-ModelName"><code>ModelName</code></a>
</td>
<td>
   <p>Name represents the model name.</p>
</td>
</tr>
<tr><td><code>role</code><br/>
<a href="#llmaz-io-v1alpha1-ModelRole"><code>ModelRole</code></a>
</td>
<td>
   <p>Role represents the model role once more than one model is required.
Such as a draft role, which means running with SpeculativeDecoding,
and default arguments for backend will be searched in backendRuntime
with the name of speculative-decoding.</p>
</td>
</tr>
</tbody>
</table>

## `ModelRole`     {#llmaz-io-v1alpha1-ModelRole}

(Alias of `string`)

**Appears in:**

- [ModelRef](#llmaz-io-v1alpha1-ModelRef)





## `ModelSource`     {#llmaz-io-v1alpha1-ModelSource}


**Appears in:**

- [ModelSpec](#llmaz-io-v1alpha1-ModelSpec)


<p>ModelSource represents the source of the model.
Only one model source will be used.</p>


<table class="table">
<thead><tr><th width="30%">Field</th><th>Description</th></tr></thead>
<tbody>
    
  
<tr><td><code>modelHub</code><br/>
<a href="#llmaz-io-v1alpha1-ModelHub"><code>ModelHub</code></a>
</td>
<td>
   <p>ModelHub represents the model registry for model downloads.</p>
</td>
</tr>
<tr><td><code>uri</code><br/>
<a href="#llmaz-io-v1alpha1-URIProtocol"><code>URIProtocol</code></a>
</td>
<td>
   <p>URI represents a various kinds of model sources following the uri protocol, protocol://<!-- raw HTML omitted -->, e.g.</p>
<ul>
<li>oss://<!-- raw HTML omitted -->.<!-- raw HTML omitted -->/<!-- raw HTML omitted --></li>
<li>ollama://llama3.3</li>
<li>host://<!-- raw HTML omitted --></li>
</ul>
</td>
</tr>
</tbody>
</table>

## `ModelSpec`     {#llmaz-io-v1alpha1-ModelSpec}


**Appears in:**

- [OpenModel](#llmaz-io-v1alpha1-OpenModel)


<p>ModelSpec defines the desired state of Model</p>


<table class="table">
<thead><tr><th width="30%">Field</th><th>Description</th></tr></thead>
<tbody>
    
  
<tr><td><code>familyName</code> <B>[Required]</B><br/>
<a href="#llmaz-io-v1alpha1-ModelName"><code>ModelName</code></a>
</td>
<td>
   <p>FamilyName represents the model type, like llama2, which will be auto injected
to the labels with the key of <code>llmaz.io/model-family-name</code>.</p>
</td>
</tr>
<tr><td><code>source</code> <B>[Required]</B><br/>
<a href="#llmaz-io-v1alpha1-ModelSource"><code>ModelSource</code></a>
</td>
<td>
   <p>Source represents the source of the model, there're several ways to load
the model such as loading from huggingface, OCI registry, s3, host path and so on.</p>
</td>
</tr>
<tr><td><code>inferenceConfig</code> <B>[Required]</B><br/>
<a href="#llmaz-io-v1alpha1-InferenceConfig"><code>InferenceConfig</code></a>
</td>
<td>
   <p>InferenceConfig represents the inference configurations for the model.</p>
</td>
</tr>
<tr><td><code>ownedBy</code><br/>
<code>string</code>
</td>
<td>
   <p>OwnedBy represents the owner of the running models serving by the backends,
which will be exported as the field of &quot;OwnedBy&quot; in openai-compatible API &quot;/models&quot;.
Default to &quot;llmaz&quot; if not set.</p>
</td>
</tr>
<tr><td><code>createdAt</code><br/>
<a href="https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.28/#time-v1-meta"><code>k8s.io/apimachinery/pkg/apis/meta/v1.Time</code></a>
</td>
<td>
   <p>CreatedAt represents the creation timestamp of the running models serving by the backends,
which will be exported as the field of &quot;Created&quot; in openai-compatible API &quot;/models&quot;.
It follows the format of RFC 3339, for example &quot;2024-05-21T10:00:00Z&quot;.</p>
</td>
</tr>
</tbody>
</table>

## `ModelStatus`     {#llmaz-io-v1alpha1-ModelStatus}


**Appears in:**

- [OpenModel](#llmaz-io-v1alpha1-OpenModel)


<p>ModelStatus defines the observed state of Model</p>


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

## `URIProtocol`     {#llmaz-io-v1alpha1-URIProtocol}

(Alias of `string`)

**Appears in:**

- [ModelSource](#llmaz-io-v1alpha1-ModelSource)


<p>URIProtocol represents the protocol of the URI.</p>



  