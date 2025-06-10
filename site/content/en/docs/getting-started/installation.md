---
title: Installation
weight: 2
description: >
    This section introduces the installation guidance for llmaz.
---

## Install a released version (recommended)

### Install

```cmd
helm install llmaz oci://registry-1.docker.io/inftyai/llmaz --namespace llmaz-system --create-namespace --version 0.0.10
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

If you want to change the default configurations, please change the values in [values.global.yaml](https://github.com/InftyAI/llmaz/blob/main/chart/values.global.yaml).

**Do not change** the values in _values.yaml_ because it's auto-generated and will be overwritten.

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