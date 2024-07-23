# llmaz

[![stability-wip](https://img.shields.io/badge/stability-wip-lightgrey.svg)](https://github.com/mkenney/software-guides/blob/master/STABILITY-BADGES.md#work-in-progress)
[![GoReport Widget]][GoReport Status]
[![Latest Release](https://img.shields.io/github/v/release/inftyai/llmaz?include_prereleases)](https://github.com/inftyai/llmaz/releases/latest)

[GoReport Widget]: https://goreportcard.com/badge/github.com/inftyai/llmaz
[GoReport Status]: https://goreportcard.com/report/github.com/inftyai/llmaz

**llmaz** (pronounced `/lima:z/`), aims to provide a **Production-Ready** inference platform for large language models on **Kubernetes**. It closely integrates with state-of-the-art inference backends like [vLLM](https://github.com/vllm-project/vllm) to bring the cutting-edge researches to cloud.

## Concept

![image](./docs/assets/overview.png)

## Feature Overview

- **User Friendly**: People can quick deploy a LLM service with minimal configurations.
- **High Performance**: llmaz integrates with vLLM by default for high performance inference. Other backends support are on the way.
- **Scaling Efficiency**: llmaz works smoothly with autoscaling components like [cluster-autoscaler](https://github.com/kubernetes/autoscaler/tree/master/cluster-autoscaler) or [Karpenter](https://github.com/kubernetes-sigs/karpenter) to support elastic cases.
- **Accelerator Fungibility**: llmaz supports serving the same LLMs with various accelerators to optimize cost and performance.
- **SOTA Inference**: llmaz support the latest cutting-edge researches like [Speculative Decoding](https://arxiv.org/abs/2211.17192) and [Splitwise](https://arxiv.org/abs/2311.18677).

## Quick Start

### Installation

Read the [Installation](./docs/installation.md) for guidance.

### Deploy

Once `Model`s (e.g. facebook/opt-125m) are published, you can quick deploy a `Playground` to serve the model.

#### Model

```yaml
apiVersion: llmaz.io/v1alpha1
kind: Model
metadata:
  name: opt-125m
spec:
  familyName: opt
  dataSource:
    modelID: facebook/opt-125m
  inferenceFlavors:
  - name: t4 # GPU type
    requests:
      nvidia.com/gpu: 1
```

#### Inference Playground

```yaml
apiVersion: inference.llmaz.io/v1alpha1
kind: Playground
metadata:
  name: opt-125m
spec:
  replicas: 1
  modelClaim:
    modelName: opt-125m
```

### Test

#### Expose the service

```cmd
kubectl port-forward pod/opt-125m-0 8080:8080
```

#### See registered models

```cmd
curl http://localhost:8080/v1/models
```

#### Request a query

```cmd
curl http://localhost:8080/v1/completions \
-H "Content-Type: application/json" \
-d '{
    "model": "facebook/opt-125m",
    "prompt": "San Francisco is a",
    "max_tokens": 10,
    "temperature": 0
}'
```

Refer to **[examples](/docs/examples/README.md)** to learn more.

## Roadmap

- Gateway support for traffic routing
- Serverless support for cloud-agnostic users
- CLI tool support
- Model training, fine tuning in the long-term

## Contributions

🚀 All kinds of contributions are welcomed ! Please follow [Contributing](https://github.com/InftyAI/community/blob/main/CONTRIBUTING.md).

## Contributors

🎉 Thanks to all these contributors.

<a href="https://github.com/InftyAI/llmaz/graphs/contributors">
  <img src="https://contrib.rocks/image?repo=InftyAI/llmaz" />
</a>
