# Open-WebUI

[Open WebUI](https://github.com/open-webui/open-webui) is a user-friendly AI interface with OpenAI-compatible APIs, serving as the default chatbot for llmaz.

## Prerequisites

- Make sure you're located in **llmaz-system** namespace, haven't tested with other namespaces.
- Make sure [EnvoyGateway](https://github.com/envoyproxy/gateway) and [Envoy AI Gateway](https://github.com/envoyproxy/ai-gateway) are installed, both of them are installed by default in llmaz. See [AI Gateway](docs/envoy-ai-gateway.md) for more details.

## How to use

If open-webui already installed, what you need to do is just update the OpenAI API endpoint in the admin settings. You can get the value from step2 & 3 below. Otherwise, following the steps here to install open-webui.

1. Enable Open WebUI in the `values.global.yaml` file, open-webui is enabled by default.

    ```yaml
    open-webui:
      enabled: true
    ```

    > Optional to set the `persistence=true` to persist the data, recommended for production.

2. Run `kubectl get svc -n llmaz-system` to list out the services, the output looks like:

    ```cmd
    envoy-default-default-envoy-ai-gateway-dbec795a   LoadBalancer   10.96.145.150   <pending>     80:30548/TCP                              132m
    envoy-gateway                                     ClusterIP      10.96.52.76     <none>        18000/TCP,18001/TCP,18002/TCP,19001/TCP   172m
    ```

3. Set `openaiBaseApiUrl` in the `values.global.yaml` like:

    ```yaml
    open-webui:
      enabled: true
      openaiBaseApiUrl: http://envoy-default-default-envoy-ai-gateway-dbec795a.llmaz-system.svc.cluster.local/v1
    ```

4. Run `make install-chatbot` to install the chatbot.

5. Port forwarding by:
    ```
    kubectl port-forward svc/open-webui 8080:80
    ```

6. Visit [http://localhost:8080](http://localhost:8080) to access the Open WebUI.

7. Configure the administrator for the first time.

**That's it! You can now chat with llmaz models with Open WebUI.**
