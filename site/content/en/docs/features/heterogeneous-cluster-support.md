---
title: Heterogeneous Cluster Support
weight: 2
---

A `llama2-7B` model can be running on __1xA100__ GPU, also on __1xA10__ GPU, even on __1x4090__ and a variety of other types of GPUs as well, that's what we called resource fungibility. In practical scenarios, we may have a heterogeneous cluster with different GPU types, and high-end GPUs will stock out a lot, to meet the SLOs of the service as well as the cost, we need to schedule the workloads on different GPU types. With the [ResourceFungibility](https://github.com/InftyAI/scheduler-plugins/blob/main/pkg/plugins/resource_fungibility) in the InftyAI scheduler, we can simply achieve this with at most 8 alternative GPU types.

## How to use

### Enable InftyAI scheduler

Edit the `values.global.yaml` file to modify the following values:

```yaml
kube-scheduler:
  enabled: true

globalConfig:
  configData: |-
    scheduler-name: inftyai-scheduler
```

Run `make helm-upgrade` to install or upgrade llmaz.
