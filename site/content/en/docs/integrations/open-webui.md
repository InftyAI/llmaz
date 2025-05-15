---
title: Open-WebUI
weight: 2
---

[Open WebUI](https://github.com/open-webui/open-webui) is a user-friendly AI interface with OpenAI-compatible APIs, serving as the default chatbot for llmaz.

## Prerequisites

- Make sure [EnvoyGateway](https://github.com/envoyproxy/gateway) and [Envoy AI Gateway](https://github.com/envoyproxy/ai-gateway) are installed, both of them are installed by default in llmaz. See [AI Gateway](docs/envoy-ai-gateway.md) for more details.

## How to use

### Enable Open WebUI

Open-WebUI is enabled by default in the `values.global.yaml` and will be deployed in llmaz-system.

```yaml
open-webui:
  enabled: true
```

### Set the Service Address

1. Run `kubectl get svc -n llmaz-system` to list out the services, the output looks like below, the LoadBalancer service name will be used later.

    ```cmd
    envoy-default-default-envoy-ai-gateway-dbec795a   LoadBalancer   10.96.145.150   <pending>     80:30548/TCP                              132m
    envoy-gateway                                     ClusterIP      10.96.52.76     <none>        18000/TCP,18001/TCP,18002/TCP,19001/TCP   172m
    ```

2. Port forward the Open-WebUI service, and visit `http://localhost:8080`.

    ```bash
    kubectl port-forward svc/open-webui 8080:80 -n llmaz-system
    ```

3. Click `Settings -> Admin Settings -> Connections`, set the URL to `http://envoy-default-default-envoy-ai-gateway-dbec795a.llmaz-system.svc.cluster.local/v1` and save. (You can also set the `openaiBaseApiUrl` in the `values.global.yaml`)

![img](/images/open-webui-setting.png)

4. Start to chat now.


## Persistence

Set the `persistence=true` in `values.global.yaml` to enable persistence.