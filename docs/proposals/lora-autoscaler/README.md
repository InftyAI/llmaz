# Proposal-27: LoRA Autoscaler

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
- [Summary](#summary)
- [Motivation](#motivation)
  - [Goals](#goals)
  - [Non-Goals](#non-goals)
- [Proposal](#proposal)
  - [User Stories (Optional)](#user-stories-optional)
    - [Story 1](#story-1)
    - [Story 2](#story-2)
  - [Notes/Constraints/Caveats (Optional)](#notesconstraintscaveats-optional)
  - [Risks and Mitigations](#risks-and-mitigations)
- [Design Details](#design-details)
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

[documentation style guide]: https://github.com/kubernetes/community/blob/master/contributors/guide/style-guide.md
-->

## Motivation

<!--
This section is for explicitly listing the motivation, goals, and non-goals of
this Proposal.  Describe why the change is important and the benefits to users. The
motivation section can optionally provide links to [experience reports] to
demonstrate the interest in a Proposal within the wider Kubernetes community.

[experience reports]: https://github.com/golang/go/wiki/ExperienceReports
-->

The foundation model size of GenAI is becoming bigger and bigger, which leads to the hight latency of autoscaling
of new model servers, more serious with new nodes. LoRA adapter, on the other hand, is a lightweight solution for
different scenarios will less training cost and resource requirement. The combination of **Foundation Model** + **Multi LoRA Adapter**
would be a dense solution for the sake of the cost saving and latency reduction.

### Goals

<!--
List the specific goals of the Proposal. What is it trying to achieve? How will we
know that this has succeeded?
-->

- Support to serving lora models via both the Playground and the inference Service
- Support to exchange the lora models in the runtime
- Support to autoscale LoRAs based on the load
- Autoscaling framework should be easy to extend with other metrics
- Integrate with vLLM as the first step which supports load/unload LoRAs in the runtime
- Route the lora requests to the specific lora server

### Non-Goals

<!--
What is out of scope for this Proposal? Listing non-goals helps to focus discussion
and make progress.
-->

- Efficient loading lora models, this should be designed with another proposal
- Different scaling policies to implement, this will be designed in another proposal
- More fine-gained lora requests routing policies should be designed in another proposal, like:
  - spread scheduling
  - binpack scheduling
  - latency-aware scheduling
  - throughput-aware scheduling
- Support other inference engines like SGLang
- More fine-gained LoRA replica dispatching policies, right now we just dispatch the LoRAs to the replicas
  as much equally as we can

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

#### Story 1

I want to serve several LoRAs with the same foundation model and I want to route the lora requests to the specific lora server,
rather than routing the requests by random or round-robin.

#### Story 2

I want to dynamic autoscaling the LoRAs based on the request and load, for example, if model server A is under high load with lora-1,
model server B has no traffic with lora-2, so model server B should unload the lora-2 and load the lora-1 for better traffic loadbalancing.
This just looks like how HPA works for Pods.

### Notes/Constraints/Caveats (Optional)

<!--
What are the caveats to the proposal?
What are some important details that didn't come across above?
Go in to as much detail as necessary here.
This might be a good place to talk about core concepts and how they relate.
-->

- LoRA replicas should not be 0 in the cluster side to avoid the cold start.
- We may need the integrate with inference engines for providing more precious metrics.

### Risks and Mitigations

<!--
What are the risks of this proposal, and how do we mitigate? Think broadly.
For example, consider both security and how this will impact the larger
Kubernetes ecosystem.

How will security be reviewed, and by whom?

How will UX be reviewed, and by whom?

Consider including folks who also work outside the SIG or subproject.
-->

The metric is a reactive indicator, which means the latency is unavoidable, the same as HPA. But we'll try to mitigate the
latency by offering different policies for configuration. The scaling down has the same problem in response to the cost.

## Design Details

<!--
This section should contain enough information that the specifics of your
change are understandable. This may include API specs (though not always
required) or even code snippets. If there's any ambiguity about HOW your
proposal will be implemented, this is the place to discuss them.
-->

### The LoRA Autoscaler

Right now, vLLM has one lora metric `vllm:lora_requests_info` containing three labels:

- running_lora_adapters: a per-adapter count of the number requests running using that adapter, formatted as a comma-separated string
- waiting_lora_adapters: similar, except counting requests that are waiting to be scheduled
- max_lora: the static "max number of LoRAs in a single batch." configuration

We will leverage the `waiting_lora_adapters` as the dominant metric for the autoscaling decision.

At a high level, the workflow looks like this:

- Create Playground or Inference Service with the lora configured
- Dispatch the LoRAs to the instances as much equally as we can, for instance:
  - if we have 2 replicas with 3 loras, we may dispatch the LoRAs to the replicas as follows:
    - Replica 1: lora-1, lora-3
    - Replica 2: lora-2
  - if we have 7 replicas with 3 loras, we may dispatch the LoRAs to the replicas as follows:
    - Replica 1: lora-1
    - Replica 2: lora-2
    - Replica 3: lora-3
    - Replica 4: lora-1
    - Replica 5: lora-2
    - Replica 6: lora-3
    - Replica 7: lora-1

  Make sure **at least one lora exists** in replicas, to avoid lora loading overhead in runtime.
- Once the lora model loaded successfully, the gateway will update the route table for the lora requests
- The LoRA autoscaler will monitor the `waiting_lora_adapters` metrics:

  - once exceed the target threshold, the **lora autoscaler**, another controller, will jump in. It will trigger the lora loading for the hot loras but not beyond the max_lora configuration and same loras can't be in the same replica which is meaningless.
  - also once a lora is **under low load**, lora autoscaler will first cut the corresponding traffic to the lora server and then offload the lora model. Note that the offload threshold should be bigger than the loading threshold to avoid the frequent loading/unloading overhead, both of them should be configurable.

Several concerns here about the lora autoscaling:

  - the metric algorithm: right now, we'll use the `waiting_lora_adapters` as the dominant metric for the autoscaling decision
  - How lora server knows when to load/offload the lora: we'll have a new CRD for tracking the lora loading status
  - load dispatching policy: we'll make decision based on the lora number and the waiting requests. The policy should be configurable for extension in the future.
  - the boundary with pod autoscaling: basically we'll autoscaling the loras first, once the lora autoscaler can't handle the load, for example, met the max_loras for all instances, the pod autoscaler will jump in. But considering the Pod autoscaling may also depend on the waiting requests, we may need to tune the metrics.

### Test Plan

<!--
**Note:** *Not required until targeted at a release.*
The goal is to ensure that we don't accept enhancements with inadequate testing.

All code is expected to have adequate tests (eventually with coverage
expectations). Please adhere to the [Kubernetes testing guidelines][testing-guidelines]
when drafting this test plan.

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
The data can be easily read from:
https://testgrid.k8s.io/sig-testing-canaries#ci-kubernetes-coverage-unit

This can inform certain test coverage improvements that we want to do before
extending the production code to implement this enhancement.
-->

- function tests in gateway
- function tests in lora dispatching

##### Integration tests

<!--
Integration tests are contained in k8s.io/kubernetes/test/integration.
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

- webhook tests for lora validation
- controller test to make sure the Playground or Service will run successfully

##### e2e tests

<!--
This question should be filled when targeting a release.
For Alpha, describe what tests will be added to ensure proper quality of the enhancement.

For Beta and GA, add links to added tests together with links to k8s-triage for those tests:
https://storage.googleapis.com/k8s-triage/index.html

We expect no non-infra related flakes in the last month as a GA graduation criteria.
-->

- e2e tests to make sure the lora service will run successfully
- e2e tests to make sure the lora autoscaling works as expected, both scaling up and down

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
- the first Kubernetes release where an initial version of the Proposal was available
- the version of Kubernetes where the Proposal graduated to general availability
- when the Proposal was retired or superseded
-->

- 2025-03-13: Proposal submitted

## Drawbacks

<!--
Why should this Proposal _not_ be implemented?
-->

TODO.

## Alternatives

<!--
What other approaches did you consider, and why did you rule them out? These do
not need to be as detailed as the proposal, but should include enough
information to express the idea and why it was not acceptable.
-->

None.