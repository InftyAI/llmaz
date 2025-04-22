# Envoy AI Gateway

[Envoy AI Gateway](https://aigateway.envoyproxy.io/) is an open source project for using Envoy Gateway
to handle request traffic from application clients to Generative AI services.

## How to use

### 1. Enable Envoy Gateway and Envoy AI Gateway in llmaz Helm

Both of them are enabled by default in `values.global.yaml` and will be deployed in llmaz-system.

```yaml
envoy-gateway:
    enabled: true
envoy-ai-gateway:
    enabled: true
```

However, [Envoy Gateway installation](https://gateway.envoyproxy.io/latest/install/install-helm/) and [Envoy AI Gateway installation](https://aigateway.envoyproxy.io/docs/getting-started/) can be deployed standalone in case you want to deploy them in other namespaces.

### 2. Basic AI Gateway Example

To expose your models via Envoy Gateway, you need to create a GatewayClass, Gateway, and AIGatewayRoute. The following example shows how to do this.

We'll deploy two models `Qwen/Qwen2-0.5B-Instruct-GGUF` and `Qwen/Qwen2.5-Coder-0.5B-Instruct-GGUF` with llama.cpp (cpu only) and expose them via Envoy AI Gateway.

The full example is [here](./examples/envoy-ai-gateway/basic.yaml), apply it.

### 3. Check Envoy AI Gateway APIs

If Open-WebUI is enabled, you can chat via the webui (recommended), see [documentation](./open-webui.md). Otherwise, following the steps below to test the Envoy AI Gateway APIs.

- For local test with port forwarding, use `export GATEWAY_URL="http://localhost:8080"`.
- Using external IP, use `export GATEWAY_URL=$(kubectl get gateway/envoy-ai-gateway-basic -o jsonpath='{.status.addresses[0].value}')`

`$GATEWAY_URL/v1/models` will show the models that are available in the Envoy AI Gateway.

Expected response will look like this:

```json
{
  "data": [
    {
      "id": "qwen2-0.5b",
      "created": 1745327294,
      "object": "model",
      "owned_by": "Envoy AI Gateway"
    },
    {
      "id": "qwen2.5-coder",
      "created": 1745327294,
      "object": "model",
      "owned_by": "Envoy AI Gateway"
    }
  ],
  "object": "list"
}
```

`$GATEWAY_URL/v1/chat/completions` will show the chat result for the model. The request will look like this:

```bash
curl -H "Content-Type: application/json"     -d '{
        "model": "qwen2-0.5b",
        "messages": [
            {
                "role": "system",
                "content": "Hi."
            }
        ]
    }'     $GATEWAY_URL/v1/chat/completions
```

Expected response will look like this:

```json
{
  "choices": [
    {
      "finish_reason": "stop",
      "index": 0,
      "message": {
        "role": "assistant",
        "content": "Hello! How can I assist you today?"
      }
    }
  ],
  "created": 1745327371,
  "model": "qwen2-0.5b",
  "system_fingerprint": "b5124-bc091a4d",
  "object": "chat.completion",
  "usage": {
    "completion_tokens": 10,
    "prompt_tokens": 10,
    "total_tokens": 20
  },
  "id": "chatcmpl-AODlT8xnf4OjJwpQH31XD4yehHLnurr0",
  "timings": {
    "prompt_n": 1,
    "prompt_ms": 319.876,
    "prompt_per_token_ms": 319.876,
    "prompt_per_second": 3.1262114069201816,
    "predicted_n": 10,
    "predicted_ms": 1309.393,
    "predicted_per_token_ms": 130.9393,
    "predicted_per_second": 7.63712651587415
  }
}
```
