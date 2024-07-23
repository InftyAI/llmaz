# llmaz

[![stability-wip](https://img.shields.io/badge/stability-wip-lightgrey.svg)](https://github.com/mkenney/software-guides/blob/master/STABILITY-BADGES.md#work-in-progress)
[![GoReport Widget]][GoReport Status]
[![Latest Release](https://img.shields.io/github/v/release/inftyai/llmaz?include_prereleases)](https://github.com/inftyai/llmaz/releases/latest)

[GoReport Widget]: https://goreportcard.com/badge/github.com/inftyai/llmaz
[GoReport Status]: https://goreportcard.com/report/github.com/inftyai/llmaz

llmaz, pronounced as `/lima:z/`, aims to provide a production-ready inference platform for large language models on Kubernetes. It tightly integrates with state-of-the-art inference backends, such as [vLLM](https://github.com/vllm-project/vllm).

## Concept

![image](./docs/assets/overview.png)

## Feature Overview

- **Easy to use**: People can deploy a production-ready LLM service with minimal configurations.
- **High performance**: llmaz integrates with vLLM by default for high performance inference. Other backend supports are on the way.
- **Autoscaling efficiency**: llmaz works smoothly with autoscaling components like [cluster-autoscaler](https://github.com/kubernetes/autoscaler/tree/master/cluster-autoscaler) and [Karpenter](https://github.com/kubernetes-sigs/karpenter) to support elastic scenarios.
- **Accelerator fungibility**: llmaz supports serving LLMs with different accelerators for the sake of cost and performance.
- **SOTA inference technologies**: llmaz support the latest SOTA technologies like [speculative decoding](https://arxiv.org/abs/2211.17192) and [Splitwise](https://arxiv.org/abs/2311.18677).

## Quick Start

Once `Model`s (e.g. opt-125m) published, you can quick deploy a `Playground` for serving.

### Model

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
  - name: t4
    requests:
      nvidia.com/gpu: 1
```

### Inference Playground

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

Refer to more **[Examples](/docs/examples/README.md)** for references.

## Roadmap

- Metrics support
- Autoscaling support
- Gateway support
- Serverless support
- CLI tool
- Model training, fine tuning in the long-term.

## Contributions

🚀 All kinds of contributions are welcomed ! Please follow [Contributing](https://github.com/InftyAI/community/blob/main/CONTRIBUTING.md).

## Contributors

🎉 Thanks to all these contributors.

<a href="https://github.com/InftyAI/llmaz/graphs/contributors">
  <img src="https://contrib.rocks/image?repo=InftyAI/llmaz" />
</a>