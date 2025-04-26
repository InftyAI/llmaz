# Installation Guide

## Prerequisites

**Requirements**:

- Kubernetes version >= 1.26
- Helm 3, see [installation](https://helm.sh/docs/intro/install/).
- Prometheus, see [installation](https://github.com/InftyAI/llmaz/tree/main/docs/prometheus-operator#install-the-prometheus-operator).

Note: llmaz helm chart will by default install
- [Envoy Gateway](https://github.com/envoyproxy/gateway) and [Envoy AI Gateway](https://github.com/envoyproxy/ai-gateway) as the frontier in the llmaz-system, if you *already installed these two components* or *want to deploy in other namespaces* , append `--set envoy-gateway.enabled=false --set envoy-ai-gateway.enabled=false` to the command below.
- [Open WebUI](https://github.com/open-webui/open-webui) as the default chatbot, if you want to disable it, append `--set open-webui.enabled=false` to the command below.

## Install a released version

### Install

```cmd
helm repo add inftyai https://inftyai.github.io/llmaz
helm repo update
helm install llmaz inftyai/llmaz --namespace llmaz-system --create-namespace --version 0.0.9
```

### Uninstall

```cmd
helm uninstall llmaz --namespace llmaz-system
kubectl delete ns llmaz-system
```

If you want to delete the CRDs as well, run

```cmd
kubectl delete crd \
    openmodels.llmaz.io \
    backendruntimes.inference.llmaz.io \
    playgrounds.inference.llmaz.io \
    services.inference.llmaz.io
```

## Install from source

### Change configurations

If you want to change the default configurations, please change the values in [values.global.yaml](../chart/values.global.yaml).

**Do you change** the values in _values.yaml_ because it's auto-generated and will be overwritten.


### Install

```cmd
git clone https://github.com/inftyai/llmaz.git && cd llmaz
kubectl create ns llmaz-system && kubens llmaz-system
make helm-install
```

### Uninstall

```cmd
helm uninstall llmaz --namespace llmaz-system
kubectl delete ns llmaz-system
```

If you want to delete the CRDs as well, run

```cmd
kubectl delete crd \
    openmodels.llmaz.io \
    backendruntimes.inference.llmaz.io \
    playgrounds.inference.llmaz.io \
    services.inference.llmaz.io
```

## Upgrade

Once you changed your code, run the command to upgrade the controller:

```cmd
IMG=<image-registry>:<tag> make helm-upgrade
```
