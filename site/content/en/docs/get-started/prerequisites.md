---
title: Prerequisites
weight: 1
description: >
    This section contains the prerequisites for llmaz.
---

**Requirements**:

- Kubernetes version >= 1.27.

    LWS requires Kubernetes version **v1.27 or higher**. If you are using a lower Kubernetes version and most of your workloads rely on single-node inference, we may consider replacing LWS with a Deployment-based approach. This fallback plan would involve using Kubernetes Deployments to manage single-node inference workloads efficiently. See [#32](https://github.com/InftyAI/llmaz/issues/32) for more details and updates.
- Helm 3, see [installation](https://helm.sh/docs/intro/install/).

Note that llmaz helm chart will by default install:

- [LWS](https://github.com/kubernetes-sigs/lws) as the default inference workload in the llmaz-system, if you *already installed it * or *want to deploy it in other namespaces* , append `--set leaderWorkerSet.enabled=false` to the command below.
- [Envoy Gateway](https://github.com/envoyproxy/gateway) and [Envoy AI Gateway](https://github.com/envoyproxy/ai-gateway) as the frontier in the llmaz-system, if you *already installed these two components* or *want to deploy in other namespaces* , append `--set envoy-gateway.enabled=false --set envoy-ai-gateway.enabled=false` to the command below.
- [Open WebUI](https://github.com/open-webui/open-webui) as the default chatbot, if you want to disable it, append `--set open-webui.enabled=false` to the command below.
