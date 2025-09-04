# Proposal-361: P2P-based Model and Image Distribution with Dragonfly

<!-- toc -->
- [Summary](#summary)
- [Motivation](#motivation)
  - [Goals](#goals)
  - [Non-Goals](#non-goals)
- [Proposal](#proposal)
  - [User Stories](#user-stories)
    - [Story 1: Multi-Replica Deployment Acceleration](#story-1-multi-replica-deployment-acceleration)
    - [Story 2: Elastic Scaling Efficiency](#story-2-elastic-scaling-efficiency)
  - [Risks and Mitigations](#risks-and-mitigations)
- [Design Details](#design-details)
  - [Component Responsibilities](#component-responsibilities)
  - [End-to-End Workflow](#end-to-end-workflow)
  - [Test Plan](#test-plan)
- [Drawbacks](#drawbacks)
- [Alternatives](#alternatives)
  - [Alternative 1: Policy-Driven Preheating](#alternative-1-policy-driven-preheating)
  - [Alternative 2: Decoupled Preheating via New Controller](#alternative-2-decoupled-preheating-via-new-controller)
<!-- /toc -->

## Summary

This proposal outlines a plan to integrate Dragonfly as a P2P distribution layer within llmaz to accelerate the distribution of model files and container images. The core of this proposal is to leverage Dragonfly as a transparent, on-demand caching and acceleration layer. This approach minimizes modifications to the existing architecture and CRDs, avoids introducing new user-facing concepts, and ensures that P2P distribution is activated precisely when resources are needed. The primary goal is to significantly reduce the time and bandwidth required for multi-replica deployments and elastic scaling scenarios by eliminating redundant downloads from source repositories.

## Motivation

Currently, when deploying a model with multiple replicas, each Pod independently downloads the model files and container images from their source (e.g., Hugging Face, Docker Hub). This leads to two major problems:

1.  **Network Bottleneck**: A large-scale deployment can saturate the cluster's egress bandwidth as numerous nodes attempt to download the same large files simultaneously.
2.  **Slow Startup Times**: The end-to-end startup time for each replica is directly tied to the download speed from the public internet, making elastic scaling inefficient and slow.

By integrating Dragonfly, we can have each file or image layer downloaded from the source exactly once, and then rapidly distributed across nodes via a high-speed internal P2P network.

### Goals

-   Reduce the total public network bandwidth consumed when deploying multiple replicas of a service.
-   Significantly accelerate the startup time for the 2nd to Nth replicas of a service.
-   Implement this acceleration layer with minimal changes to the existing CRDs and user workflow.
-   Provide a unified distribution mechanism for both model files and container images.

### Non-Goals

-   Implement a full-fledged, policy-driven, proactive preheating system in the first version. The focus is on on-demand, reactive P2P acceleration.
-   Modify the core CRDs (`OpenModel`, `BackendRuntime`, `Playground`).
-   Require users to manually create `P2PTask` or other new resources for standard deployments.

## Proposal

We will integrate Dragonfly as a transparent P2P cache. The distribution of model files and container images will be triggered on-demand by the creation of `llmaz` resources.

1.  **Model Distribution**: The `model-controller` will be updated to create a `P2PTask` when an `OpenModel` is created. This task will use Dragonfly to prepare the model files for distribution.
2.  **Image Distribution**: The `playground-controller` will be updated. When a `Playground` is created, it will trigger the creation of a `P2PTask` for the container image specified in the corresponding `BackendRuntime`.
3.  **Execution**: A new `p2ptask-controller` will manage the lifecycle of `P2PTask` resources, interacting with the Dragonfly API to perform the actual distribution. The `playground-controller` will wait for both model and image tasks to complete before creating the final `Deployment`.
4.  **Consumption**: The `model-loader` init container will be enhanced to use `dfget` to download models via the P2P network. The cluster's container runtime (`containerd`) will be configured to use the local `dfdaemon` as a registry mirror, transparently accelerating all image pulls.

This design ensures that the P2P distribution is an enhancement to the existing workflow rather than a replacement, providing acceleration without adding complexity for the end-user.

### User Stories

#### Story 1: Multi-Replica Deployment Acceleration

As a platform operator, I want to deploy a 50GB Llama3 model with 10 replicas. When I create my `Playground` resource, I expect the first replica to start downloading the model and image. I expect the subsequent 9 replicas to acquire the model and image files primarily from the first replica via the internal network, resulting in a much faster overall deployment time and significantly less traffic to Hugging Face and Docker Hub.

#### Story 2: Elastic Scaling Efficiency

As an ML engineer, my service is configured with an HPA. When a traffic spike occurs, the HPA scales my service from 2 to 8 replicas. The 6 new replicas are scheduled on new nodes. I expect these new pods to start in seconds, not minutes, because the required container image and model files are already available on the cluster nodes via Dragonfly's P2P network, are being distributed with high efficiency.

### Risks and Mitigations

-   **Risk**: The P2P distribution process introduces a new point of failure. If the Dragonfly system is down, all model and image distribution could fail.
    -   **Mitigation**: The `p2ptask-controller` and `dfget` client should be implemented with a fallback mechanism. If the Dragonfly manager or daemon is unreachable, it should fall back to downloading directly from the source URL. This ensures system availability, albeit without acceleration.
-   **Risk**: The `playground-controller` logic becomes more complex, as it now needs to manage the state of `P2PTask` resources.
    -   **Mitigation**: The logic must be implemented with robust error handling, clear status conditions on the `Playground` resource, and comprehensive unit and integration tests to ensure its stability.

## Design Details

### Component Responsibilities

1.  **`OpenModel` / `BackendRuntime` / `Playground` CRDs**: **Unchanged**.
2.  **`P2PTask` CRD**: **Retained** as the standard task interface for interacting with Dragonfly.
3.  **`model-controller`**:
    -   Watches for `OpenModel` creation.
    -   **Only responsible** for creating a `P2PTask` for the model files.
    -   Monitors the task and updates the `OpenModel` status to `Ready` upon completion.
4.  **`playground-controller` (Extended)**:
    -   Watches for `Playground` creation.
    -   Waits for the referenced `OpenModel` to become `Ready`.
    -   **Checks for and creates** a `P2PTask` for the `BackendRuntime` image.
    -   Waits for the image `P2PTask` to complete.
    -   **Finally**, creates the `Deployment` with a P2P-enhanced `model-loader`.
5.  **`p2ptask-controller`**:
    -   Watches all `P2PTask` resources and calls the Dragonfly API to execute them.
6.  **`model-loader` (P2P-Enhanced)**:
    -   Uses `dfget` internally to download model files via the P2P network.

### End-to-End Workflow

The detailed workflow and comparison with the original project can be found in the main body of this document. The core improvement is the parallel, P2P-accelerated preparation of both model and image files, triggered on-demand, before the final Pods are created.

### Test Plan

-   **Unit Tests**:
    -   `model-controller`: Verify that a `P2PTask` with the correct file task is created when an `OpenModel` is reconciled.
    -   `playground-controller`:
        -   Verify it waits for the `OpenModel` to be ready.
        -   Verify it creates a `P2PTask` for the correct image.
        -   Verify it waits for the image `P2PTask` to be ready.
        -   Verify it only creates the `Deployment` after all tasks are complete.
    -   `p2ptask-controller`: Verify it correctly calls the (mocked) Dragonfly client based on the `P2PTask` spec.
-   **Integration Tests**:
    -   Create a `Playground` and verify that the `Deployment` is only created after the mock `p2ptask-controller` updates the status of both the model and image `P2PTask` resources.
-   **E2E Tests**:
    -   A full end-to-end test in a Kind cluster with Dragonfly installed. The test will deploy a `Playground` with multiple replicas and assert that the overall deployment time is significantly faster than the baseline without Dragonfly. It will also verify that the traffic to the source registries is minimized.

## Drawbacks

1.  **Introduces Pre-Deployment Latency**: The primary trade-off. The user will experience a delay between applying a `Playground` manifest and the Pods being created, as the controller is waiting for P2P tasks to complete. This makes the process more observable.
2.  **Increased System Complexity**: The introduction of `P2PTask`, `p2ptask-controller`, and the modifications to existing controllers add to the overall complexity and maintenance overhead of the system.

## Alternatives

### Alternative 1: Policy-Driven Preheating

-   **Description**: This alternative involved creating a new `PreheatPolicy` CRD. Administrators would define policies that link `OpenModel` resources (via label selectors) to a list of container images that should be preheated. The `model-controller` would then be responsible for creating a single `P2PTask` for both the model and all matching images.
-   **Reason for Rejection**: While this approach offers the best possible performance (zero-latency deployments if preheated), it was rejected for the initial implementation due to its higher complexity. It introduces a new API for users to learn and manage, and the "policy matching" logic can be non-trivial. The chosen on-demand approach is simpler and a more incremental improvement over the baseline.