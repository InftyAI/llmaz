# Proposal-361: P2P-based Model and Image Distribution with Dragonfly

- [Summary](#summary)
- [Motivation](#motivation)
  - [Goals](#goals)
  - [Non-Goals](#non-goals)
- [Proposal](#proposal)
  - [User Stories](#user-stories)
    - [Story 1: Fast Model Distribution in a Multi-Node Cluster](#story-1-fast-model-distribution-in-a-multi-node-cluster)
  - [Risks and Mitigations](#risks-and-mitigations)
- [Design Details](#design-details)
  - [Controller-Driven Integration (Operator Pattern)](#controller-driven-integration-operator-pattern)
  - [Test Plan](#test-plan)
- [Implementation History](#implementation-history)
- [Alternatives](#alternatives)

## Summary

This proposal aims to integrate Dragonfly, a P2P-based file and image distribution system, into llmaz to accelerate the distribution of large model files and container images. In multi-node, multi-replica scenarios, direct downloads from sources like Hugging Face cause significant network overhead and slow down pod startup. By leveraging Dragonfly's P2P mechanism, we can ensure that a model is downloaded from the source only once per cluster, and subsequent requests are fulfilled by peers within the cluster, drastically improving efficiency and stability.

This integration will follow a controller-driven pattern that aligns with the project's existing design philosophy, ensuring a robust and loosely-coupled architecture.

## Motivation

The core motivation is to solve the "last-mile" distribution problem for large language models in a Kubernetes environment. Models can be tens or hundreds of gigabytes, and container images for inference are also large.

### Goals

-   **Reduce Pod Startup Time**: Significantly decrease the time it takes for model-serving pods to start by eliminating redundant downloads.
-   **Lower Network Egress Costs**: Minimize data transfer from external sources (e.g., Hugging Face, Docker Hub) to the cluster.
-   **Improve Scalability and Reliability**: Prevent the source repository from becoming a bottleneck or single point of failure during large-scale rollouts.
-   **Establish a Unified P2P Distribution Layer**: Provide a consistent mechanism for distributing both models and images.

### Non-Goals

-   This proposal will not implement a new P2P distribution system from scratch.
-   It will not initially replace the existing direct download mechanism but will act as an optional, accelerated backend.

## Proposal

We propose integrating Dragonfly by introducing a `P2PTask` CRD and an operator to manage the download lifecycle, fully decoupling the core `llmaz` components from the P2P system.

### User Stories

#### Story 1: Fast Model Distribution in a Multi-Node Cluster

As a platform operator, I have a 10-node GPU cluster. I want to deploy a 70B parameter model with 10 replicas. When I create the `Service` object, I expect the system to efficiently manage the model download. The model should be fetched from the source only once, and then rapidly distributed to all 10 pods via the in-cluster P2P network. This should reduce the total deployment time from hours to minutes.

### Risks and Mitigations

-   **Risk**: Dragonfly introduces additional components (Manager, DaemonSet) to the cluster, increasing complexity.
    -   **Mitigation**: Provide clear documentation and Helm charts for deploying Dragonfly as a dependency. The integration logic in `llmaz` will be loosely coupled, interacting only with the generic `P2PTask` CRD.
-   **Risk**: The P2P network itself could have issues (e.g., peer discovery failures).
    -   **Mitigation**: The Dragonfly system has built-in fallback mechanisms to download from the source if P2P fails. We will ensure this is properly configured.

## Design Details

### Controller-Driven Integration (Operator Pattern)

This approach represents the ideal state for the integration, prioritizing loose coupling and clear separation of concerns.

1.  **Define `P2PTask` CRD**: Introduce a `P2PTask` Custom Resource Definition as a generic API for requesting P2P downloads. This CRD will serve as the unified interface for model and image distribution, containing fields such as:
    - `url`: The source URL of the model or image.
    - `destinationPath`: The target path within the container for the downloaded artifact.
    - `p2pSystem`: A field to specify the desired P2P provider (e.g., "dragonfly").

2.  **Update `model-controller`**: Modify `pkg/controller/core/model_controller.go`. When an `OpenModel` is created, the controller will create a corresponding `P2PTask` CRD instance. It will set the `ownerReference` so the task is automatically garbage-collected with the model.

3.  **Implement Dragonfly Operator**: This new component (the "operator") will watch for `P2PTask` resources. When it finds a task configured for "dragonfly", it will:
    -   Use the Dragonfly Manager's API to trigger a pre-heating/caching job for the specified URL.
    -   Monitor the job's progress and update the `P2PTask`'s status field accordingly (e.g., `Pending`, `Downloading`, `Succeeded`, `Failed`).

4.  **Simplify `model-loader`**: The `model-loader`'s responsibility is significantly reduced. It no longer contains any download logic. Instead, it will simply wait for the file to appear at the `destinationPath`, relying on the Dragonfly operator to place it there.

### Test Plan

-   **Unit Tests**: Add unit tests for the new logic in `model_controller.go` that creates `P2PTask` CRDs.
-   **Integration Tests**: Create an integration test that:
    1.  Creates an `OpenModel` CRD.
    2.  Verifies that a `P2PTask` CRD is created correctly.
    3.  Uses a mock Dragonfly Operator to update the task's status to "Succeeded".
    4.  Verifies that the `OpenModel`'s status is updated to "Ready".
-   **E2E Tests**: Set up an E2E test environment with Kind, llmaz, and Dragonfly installed. The test will deploy a real model and verify that it is downloaded via the P2P mechanism and that the serving pod starts successfully.

## Implementation History

-   [2025-07-11] Proposal submitted.

## Alternatives

-   **Manta**: The project already has stubs for Manta integration. We could complete that implementation. However, Dragonfly is a CNCF project with a broader feature set, including image distribution, making it a more strategic choice.
-   **Nydus**: Another CNCF sandbox project focused on container image acceleration. It could be used for the image distribution part but is less focused on arbitrary file distribution for models.