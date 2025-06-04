# Proposal-106: Support scaling with Spot instances for cost saving with Karpenter

<!--
This is the title of your Proposal. Keep it short, simple, and descriptive. A good
title can help communicate what the Proposal is and should be considered as part of
any review.
-->

<!--
A table of contents is helpful for quickly jumping to sections of a Proposal and for
highlighting any additional information provided beyond the standard Proposal
template.

Ensure the TOC is wrapped with
  <code>&lt;!-- toc --&rt;&lt;!-- /toc --&rt;</code>
tags, and then generate with `hack/update-toc.sh`.
-->

<!-- toc -->
- [Proposal-106: Support scaling with Spot instances for cost saving with Karpenter](#proposal-106-support-scaling-with-spot-instances-for-cost-saving-with-karpenter)
  - [Summary](#summary)
  - [Motivation](#motivation)
    - [Goals](#goals)
    - [Non-Goals](#non-goals)
  - [Proposal](#proposal)
    - [User Stories (Optional)](#user-stories-optional)
      - [Story 1: ML Engineer – Cost-Efficient Deployment of LLMs](#story-1-ml-engineer--cost-efficient-deployment-of-llms)
      - [Story 2: Workload Author – Flexible and Preferred GPU Scheduling with Spot Support](#story-2-workload-author--flexible-and-preferred-gpu-scheduling-with-spot-support)
    - [Notes/Constraints/Caveats (Optional)](#notesconstraintscaveats-optional)
    - [Risks and Mitigations](#risks-and-mitigations)
  - [Design Details](#design-details)
    - [What's the Relationship between Karpenter and Cloud Providers](#whats-the-relationship-between-karpenter-and-cloud-providers)
    - [Model Inference Flavors](#model-inference-flavors)
      - [OpenModel](#openmodel)
      - [Pod Annotation](#pod-annotation)
    - [Provisioning controller in Forked Karpenter](#provisioning-controller-in-forked-karpenter)
    - [Test Plan](#test-plan)
        - [Prerequisite testing updates](#prerequisite-testing-updates)
        - [Unit tests](#unit-tests)
        - [Integration tests](#integration-tests)
        - [e2e tests](#e2e-tests)
    - [Graduation Criteria](#graduation-criteria)
  - [Implementation History](#implementation-history)
  - [Drawbacks](#drawbacks)
  - [Alternatives](#alternatives)
<!-- /toc -->

## Summary

<!--
This section is incredibly important for producing high-quality, user-focused
documentation such as release notes or a development roadmap. It should be
possible to collect this information before implementation begins, in order to
avoid requiring implementors to split their attention between writing release
notes and implementing the feature itself. Proposal editors and SIG Docs
should help to ensure that the tone and content of the `Summary` section is
useful for a wide audience.

A good summary is probably at least a paragraph in length.

Both in this section and below, follow the guidelines of the [documentation
style guide]. In particular, wrap lines to a reasonable length, to make it
easier for reviewers to cite specific portions, and to minimize diff churn on
updates.

-->

[Karpenter](https://github.com/kubernetes-sigs/karpenter) automatically launches just the right compute resources to handle your cluster's applications. It is designed to let you take full advantage of the cloud with fast and simple compute provisioning for Kubernetes clusters. This proposal enhances the coordination between Karpenter and [llmaz-scheduler](https://github.com/InftyAI/scheduler-plugins), enabling cost-efficient, GPU-aware autoscaling tailored for real-world, heterogeneous cloud infrastructure.

## Motivation

<!--
This section is for explicitly listing the motivation, goals, and non-goals of
this Proposal.  Describe why the change is important and the benefits to users. The
motivation section can optionally provide links to [experience reports] to
demonstrate the interest in a Proposal within the wider InftyAI community.

[experience reports]: https://github.com/golang/go/wiki/ExperienceReports
-->

A `llama2-7B` model can be running on __1xA100__ GPU, also on __1xA10__ GPU, even on __1x4090__ and a variety of other types of GPUs as well, that's what we called resource fungibility. In practical scenarios, we may have a heterogeneous cluster with different GPU types, and high-end GPUs will stock out a lot, to meet the SLOs of the service as well as the cost, we need to schedule the workloads on different GPU types. With the [ResourceFungibility](https://github.com/InftyAI/scheduler-plugins/blob/main/pkg/plugins/resource_fungibility) in the llmaz-scheduler, we can simply achieve this with at most 8 alternative GPU types.

To optimize cost, users commonly deploy their cluster in the cloud and rely on dynamic node scaling to match resource allocation with workload demand. This elasticity helps handle traffic spikes efficiently while avoiding overprovisioning and idle resource waste. In such setups, Karpenter is widely used as the provisioning engine, automatically launching the right instance types — including GPU-backed instances — in response to pending pods.

However, Karpenter is built to adhere to the scheduling decisions of kube-scheduler. So it's certainly possible we would run across some cases where Karpenter makes incorrect decisions when a custom scheduler is in the mix. As a result, it may launch a cheaper node with __1xT4__ GPU which is incompatible with the associated model's inference configuration, leading to pods remaining in the Pending state despite available resources in node pools. The root cause is that Karpenter is not aware of the llmaz-scheduler's scheduling constraints. 

### Goals

<!--
List the specific goals of the Proposal. What is it trying to achieve? How will we
know that this has succeeded?
-->

- Provision spot instances for inference workloads based on the model's flavor requirements.
- Support flexible and preferred GPU scheduling with spot instances.
- This proposal is only for AWS but the implementation can be extended to other cloud providers.
- Service Availability which means be aware of the spot instance disruption events. It's a beta-level requirement.
- Latency sensitivity which means scaling the inference service efficient, including efforts like accelerating the model loading process. It's a beta-level requirement.

### Non-Goals

<!--
What is out of scope for this Proposal? Listing non-goals helps to focus discussion
and make progress.
-->

- Integration with the [Kubernetes Cluster Autoscaler](https://github.com/kubernetes/autoscaler) is out of scope for this proposal.
- Add custom scheduler support for the upstream of the Karpenter project, and it is tracked in [this issue](https://github.com/kubernetes-sigs/karpenter/issues/742). Once the support is added in the upstream, we don't need to maintain the [forked version](https://github.com/InftyAI/karpenter).
- Support for multi-host inference in out-of-the-scope now because reclaiming one part of the service requires reclaiming other parts as well which brings instability to the service.

## Proposal

<!--
This is where we get down to the specifics of what the proposal actually is.
This should have enough detail that reviewers can understand exactly what
you're proposing, but should not include things like API designs or
implementation. What is the desired outcome and how do we measure success?.
The "Design Details" section below is for the real
nitty-gritty.
-->


### User Stories (Optional)

<!--
Detail the things that people will be able to do if this Proposal is implemented.
Include as much detail as possible so that people can understand the "how" of
the system. The goal here is to make this feel real for users without getting
bogged down.
-->

#### Story 1: ML Engineer – Cost-Efficient Deployment of LLMs

As a machine learning engineer deploying large language models (LLMs), I don't own any physical GPU servers, so I have to rent them from cloud providers. I want to automatically use cheaper GPU Spot instances for serving models when available, so that I can reduce infrastructure costs without sacrificing performance.

#### Story 2: Workload Author – Flexible and Preferred GPU Scheduling with Spot Support

As a workload author, I want to publish Kubernetes manifests for my model-serving workloads that are broadly compatible across different device types, without being overly prescriptive or requiring end users to modify them.

- My workloads are GPU-dependent, but there are many different GPU models available in the cloud (e.g., A10, A100, H100).
- Instead of locking my manifests to a single GPU type, I want to express a preference-ordered list of compatible GPU types (e.g., prefer A100, fall back to A10 or L4).
- This gives end users the flexibility to run the same manifest on different underlying infrastructure.
- If none of the existing nodes in the cluster meet the constraints (e.g., no compatible GPUs available), I want the system to automatically provision an appropriate Spot instance from the cloud provider, based on my declared GPU preferences and resource requirements.

This approach allows me to build and share portable workloads that are cost-aware, device-flexible, and production-safe, without the need for users to rewrite manifests or manage instance-level complexity themselves.

### Notes/Constraints/Caveats (Optional)

<!--
What are the caveats to the proposal?
What are some important details that didn't come across above?
Go in to as much detail as necessary here.
This might be a good place to talk about core concepts and how they relate.
-->

### Risks and Mitigations

<!--
What are the risks of this proposal, and how do we mitigate? Think broadly.
For example, consider both security and how this will impact the larger
InftyAI ecosystem.

How will security be reviewed, and by whom?

How will UX be reviewed, and by whom?

Consider including folks who also work outside the SIG or subproject.
-->

One risk is service availability degradation when a traffic spike occurs and the workload on the new node takes a long time to become ready. This will be addressed during the beta graduation phase.

## Design Details

<!--
This section should contain enough information that the specifics of your
change are understandable. This may include API specs (though not always
required) or even code snippets. If there's any ambiguity about HOW your
proposal will be implemented, this is the place to discuss them.
-->

### What's the Relationship between Karpenter and Cloud Providers

Karpenter is a multi-cloud Kubernetes node autoscaler with a modular architecture. Its core logic is implemented in the [kubernetes-sigs/karpenter](https://github.com/kubernetes-sigs/karpenter) project, which handles scheduling and scaling decisions independently of any cloud provider. Cloud-specific functionality—such as launching instances or configuring networks—is implemented through provider plugins like [aws/karpenter-provider-aws](https://github.com/aws/karpenter-provider-aws). This separation allows Karpenter to support multiple platforms, including [AWS](https://github.com/aws/karpenter-provider-aws), [Azure](https://github.com/Azure/karpenter-provider-azure), [GCP](https://github.com/cloudpilot-ai/karpenter-provider-gcp), [Alibaba Cloud](https://github.com/cloudpilot-ai/karpenter-provider-alibabacloud), [Cluster API](https://github.com/kubernetes-sigs/karpenter-provider-cluster-api), and [Proxmox](https://github.com/sergelogvinov/karpenter-provider-proxmox), making it flexible and extensible across different infrastructure environments.

Unfortunately, Karpenter works only with the default scheduler to schedule the pods. Add Karpenter support to work with custom schedulers (e.g., [Kueue](https://github.com/kubernetes-sigs/kueue/issues/5133), [Volcano](https://github.com/volcano-sh/volcano/issues?q=is%3Aissue%20state%3Aopen%20Karpenter)) is not yet on Karpenter's roadmap. In order to make it work well with our llmaz-scheduler, we have to fork the Karpenter project and replace cloud providers' implementation with our patches until the custom scheduler support is added in the upstream.

```go.mod
replace sigs.k8s.io/karpenter => github.com/InftyAI/karpenter v0.0.0-20250528100011-b6f483045bd6
```

> The forked Karpenter is available in the [karpenter](https://github.com/InftyAI/karpenter) repository.

### Model Inference Flavors

#### OpenModel

Here is an example of the OpenModel resource with 2 flavors.

```yaml
apiVersion: llmaz.io/v1alpha1
kind: OpenModel
metadata:
  name: qwen2-0--5b
spec:
  familyName: qwen2
  source:
    modelHub:
      modelID: Qwen/Qwen2-0.5B-Instruct
  inferenceConfig:
    flavors:
    - name: t4g
      limits:
        nvidia.com/gpu: 1
      nodeSelector:
        karpenter.k8s.aws/instance-gpu-name: t4g
    - name: t4
      limits:
        nvidia.com/gpu: 1
      nodeSelector:
        karpenter.k8s.aws/instance-gpu-name: t4
```

The following labels are well-known in Karpenter:

- `node.kubernetes.io/instance-type`
- `karpenter.k8s.aws/instance-gpu-count`
- `karpenter.k8s.aws/instance-gpu-manufacturer`
- `karpenter.k8s.aws/instance-gpu-memory`
- `karpenter.k8s.aws/instance-gpu-name`

We need to specify the instance-gpu-name via nodeSelector to match the target GPU type when node is provisioned by forked Karpenter from multiple node pools.

When you only have a single node pool to provision the GPU instance and the node pool only has one GPU type, it is okay to not specify the nodeSelector. But in practice, it is better to specify the nodeSelector to make the provisioned node more predictable.

#### Pod Annotation

```go
const (
	// InferenceServiceFlavorsAnnoKey is the annotation key for the flavors specified
	// in the inference service, the value is a comma-separated list of flavor names.
	InferenceServiceFlavorsAnnoKey = "llmaz.io/inference-service-flavors"
)
```

When users deploy a model via Playground or Service, they can reorder or select a part of flavors of the model. And it will be injected into the pod annotation `llmaz.io/inference-service-flavors`.

### Provisioning controller in Forked Karpenter

[VolumeTopology](https://github.com/kubernetes-sigs/karpenter/blob/main/pkg/controllers/provisioning/scheduling/volumetopology.go) is a good example of how to inject requirements into a pod's node affinity. We can use the same approach to inject inference flavor requirements into the pod's node affinity. Why does this work for llmaz's resource fungibility? Please refer to the [Karpenter Scheduling](https://karpenter.sh/docs/concepts/scheduling/#preferences) for more details.

```go
// filename: pkg/controllers/provisioning/scheduling/modelinference.go

func init() {
	// Add support for llmaz CRDs.
	utilruntime.Must(llmazcoreapi.AddToScheme(scheme.Scheme))
	utilruntime.Must(llmazinferenceapi.AddToScheme(scheme.Scheme))
}

func NewModelInference(kubeClient client.Client) *ModelInference {
	return &ModelInference{kubeClient: kubeClient}
}

// Inject the inference flavor requirements to the pod's node affinity.
// 
// It should be called in p.NewScheduler 
func (m *ModelInference) Inject(ctx context.Context, pod *v1.Pod) error { ... }

// ValidateInferenceFlavors checks if the inference flavors specified in the pod annotation
// are valid according to the associated OpenModel's configuration.
//
// Skips validation if the pod is not created by llmaz's inference service or if no flavors are specified.
// Returns an error if any unknown flavors are found.
// 
// It should be called in p.Validate
func (m *ModelInference) ValidateInferenceFlavors(ctx context.Context, pod *v1.Pod) (err error) { ...}
```

**Inject**

Add the inference flavor requirements to the pod's node affinity. This causes it to be OR'd with every merged requirement, so that [relaxation](https://github.com/kubernetes-sigs/karpenter/blob/main/pkg/controllers/provisioning/scheduling/preferences.go) replaces our flavor requirements according to the order of flavors in the model or pod's annotation, when no existing node, in-flight node claim, or node pool can satisfy the current flavor requirements.

```yaml
affinity: 
  nodeAffinity:
    requiredDuringSchedulingIgnoredDuringExecution:
      nodeSelectorTerms:
      - matchExpressions:
          - key: karpenter.k8s.aws/instance-gpu-name
            operator: In
            values: ["t4g"]
      - matchExpressions:
          - key: karpenter.k8s.aws/instance-gpu-name
            operator: In
            values: ["t4"]
```

We add our inference requirement to every node selector term. This causes it to be AND'd with every existing requirement so that relaxation won't remove our inference requirement.

```yaml
affinity:
  nodeAffinity:
    requiredDuringSchedulingIgnoredDuringExecution:
      nodeSelectorTerms:
      - matchExpressions:
        - key: ... # existing requirements
          ...
```

to become:

```yaml
affinity:
  nodeAffinity:
    requiredDuringSchedulingIgnoredDuringExecution:
      nodeSelectorTerms:
      - matchExpressions:
        - key: karpenter.k8s.aws/instance-gpu-name # flavor 1
          operator: In
          values: ["t4g"]
        - key: ... # existing requirements
          ...
      - matchExpressions:
        - key: karpenter.k8s.aws/instance-gpu-name # flavor 2
          operator: In
          values: ["t4"]
        - key: ... # existing requirements
          ...
```

### Test Plan

<!--
**Note:** *Not required until targeted at a release.*
The goal is to ensure that we don't accept enhancements with inadequate testing.

All code is expected to have adequate tests (eventually with coverage
expectations).

[testing-guidelines]: https://git.k8s.io/community/contributors/devel/sig-testing/testing.md
-->

[x] I/we understand the owners of the involved components may require updates to
existing tests to make this code solid enough prior to committing the changes necessary
to implement this enhancement.

##### Prerequisite testing updates

<!--
Based on reviewers feedback describe what additional tests need to be added prior
implementing this enhancement to ensure the enhancements have also solid foundations.
-->

##### Unit tests

<!--
In principle every added code should have complete unit test coverage, so providing
the exact set of tests will not bring additional value.
However, if complete unit test coverage is not possible, explain the reason of it
together with explanation why this is acceptable.
-->

<!--
Additionally, for Alpha try to enumerate the core package you will be touching
to implement this enhancement and provide the current unit coverage for those
in the form of:
- <package>: <date> - <current test coverage>

This can inform certain test coverage improvements that we want to do before
extending the production code to implement this enhancement.
-->

Forked karpenter:

- `pkg/controllers/provisioning`: `Model Inference Requirements` is used to check if the model inference requirements are met when provisioning a node. And it will be added to the existing suite tests.

##### Integration tests

<!--
Integration tests allow control of the configuration parameters used to start the binaries under test.
This is different from e2e tests which do not allow configuration of parameters.
Doing this allows testing non-default options and multiple different and potentially conflicting command line options.
-->

<!--
This question should be filled when targeting a release.
For Alpha, describe what tests will be added to ensure proper quality of the enhancement.

For Beta and GA, add links to added tests together with links to k8s-triage for those tests:
https://storage.googleapis.com/k8s-triage/index.html
-->

N/A.

##### e2e tests

<!--
This question should be filled when targeting a release.
For Alpha, describe what tests will be added to ensure proper quality of the enhancement.

For Beta and GA, add links to added tests together with links to k8s-triage for those tests:
https://storage.googleapis.com/k8s-triage/index.html

We expect no non-infra related flakes in the last month as a GA graduation criteria.
-->

- Add one e2e test to make sure the whole system can be launched via helm chart. By leveraging kwok provider from the karpenter repo, we can test the whole system with spot instances without real cloud resources.
- Manually test on EKS with real spot instances using custom image which is built from the forked karpenter.

### Graduation Criteria

<!--

Clearly define what it means for the feature to be implemented and
considered stable.

If the feature you are introducing has high complexity, consider adding graduation
milestones with these graduation criteria:
- [Maturity levels (`alpha`, `beta`, `stable`)][maturity-levels]
- [Feature gate][feature gate] lifecycle
- [Deprecation policy][deprecation-policy]

[feature gate]: https://git.k8s.io/community/contributors/devel/sig-architecture/feature-gates.md
[maturity-levels]: https://git.k8s.io/community/contributors/devel/sig-architecture/api_changes.md#alpha-beta-and-stable-versions
[deprecation-policy]: https://kubernetes.io/docs/reference/using-api/deprecation-policy/
-->

## Implementation History

<!--
Major milestones in the lifecycle of a Proposal should be tracked in this section.
Major milestones might include:
- the `Summary` and `Motivation` sections being merged, signaling SIG acceptance
- the `Proposal` section being merged, signaling agreement on a proposed design
- the date implementation started
- the first llmaz release where an initial version of the Proposal was available
- the version of llmaz where the Proposal graduated to general availability
- when the Proposal was retired or superseded
-->

- 2025-06-04: Proposal drafted.

## Drawbacks

<!--
Why should this Proposal _not_ be implemented?
-->

## Alternatives

<!--
What other approaches did you consider, and why did you rule them out? These do
not need to be as detailed as the proposal, but should include enough
information to express the idea and why it was not acceptable.
-->
