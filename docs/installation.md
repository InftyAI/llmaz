# Installation Guide

## Prerequisites

* Kubernetes version >= 1.27
* Helm

## Install a released version

### Install

```cmd
helm repo add inftyai https://inftyai.github.io/llmaz
helm install llmaz inftyai/llmaz --version 0.0.2
```

### Uninstall

```cmd
helm uninstall llmaz
```

## Install from source

### Install

```cmd
make helm-install
```

### Uninstall

```cmd
helm uninstall llmaz
```
