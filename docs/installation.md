# Installation Guide

## Prerequisites

- Kubernetes version >= 1.27
- Helm 3

## Install a released version

### Install

```cmd
helm repo add inftyai https://inftyai.github.io/llmaz
helm repo update
helm install llmaz inftyai/llmaz --namespace llmaz-system --create-namespace --version 0.0.7
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
kubectl create ns llmaz-system && kubens llmaz-system
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

## Change configurations

If you want to change the default configurations, please change the values in [values.global.yaml](../chart/values.global.yaml), then run

```cmd
make helm-install
```

**Do you change** the values in _values.yaml_ because it's auto-generated and will be overwritten.

## Upgrade

Once you changed your code, run the command to upgrade the controller:

```cmd
IMG=<image-registry>:<tag> make helm-upgrade
```
