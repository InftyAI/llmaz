<p align="center">
  <picture>
    <source media="(prefers-color-scheme: dark)" srcset="https://raw.githubusercontent.com/inftyai/llmaz/main/docs/assets/logo.png">
    <img alt="llmaz" src="https://raw.githubusercontent.com/inftyai/llmaz/main/docs/assets/logo.png" width=55%>
  </picture>
</p>

<h3 align="center">
Easy, advanced inference platform for large language models on Kubernetes
</h3>

[![stability-alpha](https://img.shields.io/badge/stability-alpha-f4d03f.svg)](https://github.com/mkenney/software-guides/blob/master/STABILITY-BADGES.md#alpha)
[![GoReport Widget]][GoReport Status]
[![Latest Release](https://img.shields.io/github/v/release/inftyai/llmaz?include_prereleases)](https://github.com/inftyai/llmaz/releases/latest)

[GoReport Widget]: https://goreportcard.com/badge/github.com/inftyai/llmaz
[GoReport Status]: https://goreportcard.com/report/github.com/inftyai/llmaz

**llmaz** (pronounced `/lima:z/`), aims to provide a **Production-Ready** inference platform for large language models on Kubernetes. It closely integrates with the state-of-the-art inference backends to bring the leading-edge researches to cloud.

> 🌱 llmaz is alpha now, so API may change before graduating to Beta.

## Architecture

![image](https://github.com/InftyAI/llmaz/blob/main/docs/assets/arch.png?raw=true)

## Features Overview

- **Easy of Use**: People can quick deploy a LLM service with minimal configurations.
- **Broad Backend Support**: llmaz supports a wide range of advanced inference backends for different scenarios, like [vLLM](https://github.com/vllm-project/vllm), [SGLang](https://github.com/sgl-project/sglang), [llama.cpp](https://github.com/ggerganov/llama.cpp). Find the full list of supported backends [here](./docs/support-backends.md).
- **Scaling Efficiency (WIP)**: llmaz works smoothly with autoscaling components like [Cluster-Autoscaler](https://github.com/kubernetes/autoscaler/tree/master/cluster-autoscaler) or [Karpenter](https://github.com/kubernetes-sigs/karpenter) to support elastic scenarios.
- **Accelerator Fungibility (WIP)**: llmaz supports serving the same LLM with various accelerators to optimize cost and performance.
- **SOTA Inference**: llmaz supports the latest cutting-edge researches like [Speculative Decoding](https://arxiv.org/abs/2211.17192) or [Splitwise](https://arxiv.org/abs/2311.18677)(WIP) to run on Kubernetes.
- **Various Model Providers**: llmaz supports a wide range of model providers, such as [HuggingFace](https://huggingface.co/), [ModelScope](https://www.modelscope.cn), ObjectStores(aliyun OSS, more on the way). llmaz automatically handles the model loading requiring no effort from users.
- **Multi-hosts Support**: llmaz supports both single-host and multi-hosts scenarios with [LWS](https://github.com/kubernetes-sigs/lws) from day 1.

## How to install

### Prerequisites

* Kubernetes version >= 1.27
* Helm 3

### Install a released version

#### Install

```cmd
helm repo add inftyai https://inftyai.github.io/llmaz
helm repo update
helm install llmaz inftyai/llmaz --namespace llmaz-system --create-namespace --version 0.0.3
```

#### Uninstall

```cmd
helm uninstall llmaz
kubectl delete ns llmaz-system
```

If you want to delete the CRDs as well, run (ignore the error)
```cmd
kubectl delete crd \
    openmodels.llmaz.io \
    backendruntimes.inference.llmaz.io \
    playgrounds.inference.llmaz.io \
    services.inference.llmaz.io
```
