# Installation Guide

## Prerequisites

* Kubernetes version >= 1.27

## Install a released version

### Install

```cmd
# leaderworkerset runs in lws-system
LWS_VERSION=v0.3.0
kubectl apply --server-side -f https://github.com/kubernetes-sigs/lws/releases/download/$LWS_VERSION/manifests.yaml

# llmaz runs in llmaz-system
LLMAZ_VERSION=v0.0.2
kubectl apply --server-side -f https://github.com/inftyai/llmaz/releases/download/$LLMAZ_VERSION/manifests.yaml
```

### Uninstall

```cmd
LWS_VERSION=v0.3.0
kubectl delete -f https://github.com/kubernetes-sigs/lws/releases/download/$LWS_VERSION/manifests.yaml

LLMAZ_VERSION=v0.0.2
kubectl delete -f https://github.com/inftyai/llmaz/releases/download/$LLMAZ_VERSION/manifests.yaml
```

## Install from source

### Install

```cmd
LWS_VERSION=v0.3.0
kubectl apply --server-side -f https://github.com/kubernetes-sigs/lws/releases/download/$LWS_VERSION/manifests.yaml

git clone https://github.com/inftyai/llmaz.git && cd llmaz
IMG=<IMAGE_REPO>:<GIT_TAG> make image-push deploy
```

### Uninstall

```cmd
make undeploy
```
