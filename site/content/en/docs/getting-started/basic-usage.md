---
title: Basic Usage
weight: 3
description: >
    This section introduces the basic usage of llmaz.
---

Let's assume that you have installed the llmaz with the default settings, which means both the [AI Gateway](../integrations/envoy-ai-gateway.md) and [Open WebUI](../integrations/open-webui.md) are installed. Now let's following the steps to chat with your models.

### Deploy the Services

Run the following command to deploy two models (cpu only).

```bash
kubectl apply -f https://raw.githubusercontent.com/InftyAI/llmaz/refs/heads/main/docs/examples/envoy-ai-gateway/basic.yaml
```

### Chat with Models

Waiting for your services ready, generally looks like:

```bash
NAME                                                            READY   STATUS            RESTARTS   AGE
ai-eg-route-extproc-default-envoy-ai-gateway-6ddcd49b64-ldwcd   1/1     Running           0          6m37s
qwen2--5-coder-0                                                1/1     Running           0          6m37s
qwen2-0--5b-0                                                   1/1     Running           0          6m37s
```

Once ready, you can access the Open WebUI by port-forwarding the service:

```bash
kubectl port-forward svc/open-webui 8080:80 -n llmaz-system
```

Let's chat on `http://localhost:8080` now, two models are available to you! 🎉