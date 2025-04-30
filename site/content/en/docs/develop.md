---
title: Develop Guidance
weight: 3
description: >
  This section contains a develop guidance for people who want to learn more about this project.
---

## Project Structure

```structure
llmaz # root
├── bin # where the binaries locates, like the kustomize, ginkgo, etc.
├── chart # where the helm chart locates
├── cmd # where the main entry locates
├── docs # where all the documents locate, like examples, installation guidance, etc.
├── llmaz # where the model loader logic locates
├── pkg # where the main logic for Kubernetes controllers locates
```

## API design

### Core APIs

See the [API Reference](./reference/core.v1alpha1.md) for more details.

### Inference APIs

See the [API Reference](./reference/inference.v1alpha1.md) for more details.
