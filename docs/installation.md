# Installation Guide

## Prerequisites

* Kubernetes version >= 1.27
* Helm 3

## Install a released version

### Install

```cmd
helm repo add inftyai https://inftyai.github.io/llmaz
helm repo update
helm install llmaz inftyai/llmaz --namespace llmaz-system --create-namespace --version 0.0.2
```

### Uninstall

```cmd
helm uninstall llmaz
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

### Install

```cmd
git clone https://github.com/inftyai/llmaz.git && cd llmaz
make helm-install
```

### Uninstall

```cmd
helm uninstall llmaz
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
