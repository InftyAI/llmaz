# Envoy AI Gateway

[Envoy AI Gateway](https://aigateway.envoyproxy.io/) is an open source project for using Envoy Gateway
to handle request traffic from application clients to Generative AI services.

## How to use

### 1. Enable Envoy Gateway and Envoy AI Gateway in llmaz Helm

Enable Envoy Gateway and Envoy AI Gateway in the `values.global.yaml` file, envoy gateway and envoy ai gateway are disabled by default.

```yaml
envoy-gateway:
    enabled: true
envoy-ai-gateway:
    enabled: true
```

Note: [Envoy Gateway installation](https://gateway.envoyproxy.io/latest/install/install-helm/) and [Envoy AI Gateway installation](https://aigateway.envoyproxy.io/docs/getting-started/) can be done standalone.

### 2. Check Envoy Gateway and Envoy AI Gateway

Run `kubectl wait --timeout=5m -n envoy-gateway-system deployment/envoy-gateway --for=condition=Available` to wait for the envoy gateway to be ready.

Run `kubectl wait --timeout=2m -n envoy-ai-gateway-system deployment/ai-gateway-controller --for=condition=Available` to wait for the envoy ai gateway to be ready.

### 3. Basic AI Gateway example

To expose your model(Playground) to Envoy Gateway, you need to create a GatewayClass, Gateway, and AIGatewayRoute. The following example shows how to do this.

Example [qwen playground](docs/examples/llamacpp/playground.yaml) configuration for a basic AI Gateway.
The model name is `qwen2-0.5b`, so the backend ref name is `qwen2-0--5b`, and the model lb service: `qwen2-0--5b-lb`
- Playground in [docs/examples/llamacpp/playground.yaml](docs/examples/llamacpp/playground.yaml)
- GatewayClass in [docs/examples/envoy-ai-gateway/basic.yaml](docs/examples/envoy-ai-gateway/basic.yaml)

Check if the gateway pod to be ready:

```bash
kubectl wait pods --timeout=2m \
    -l gateway.envoyproxy.io/owning-gateway-name=envoy-ai-gateway-basic \
    -n envoy-gateway-system \
    --for=condition=Ready
```

### 4. Check Envoy AI Gateway APIs

- For local test with port forwarding, use `export GATEWAY_URL="http://localhost:8080"`. 
- Using external IP, use `export GATEWAY_URL=$(kubectl get gateway/envoy-ai-gateway-basic -o jsonpath='{.status.addresses[0].value}')`

See https://aigateway.envoyproxy.io/docs/getting-started/basic-usage for more details.

`$GATEWAY_URL/v1/models` will show the models that are available in the Envoy AI Gateway. The response will look like this:

```json
{
  "data": [
    {
      "id": "some-cool-self-hosted-model",
      "created": 1744880950,
      "object": "model",
      "owned_by": "Envoy AI Gateway"
    },
    {
      "id": "qwen2-0.5b",
      "created": 1744880950,
      "object": "model",
      "owned_by": "Envoy AI Gateway"
    }
  ],
  "object": "list"
}
```

`$GATEWAY_URL/v1/chat/completions` will show the chat completions for the model. The request will look like this:

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
            "message": {
                "content": "I'll be back."
            }
        }
    ]
}
```

