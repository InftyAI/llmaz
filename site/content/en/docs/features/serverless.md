---
title: Serverless
weight: 4
---

## Overview

This comprehensive guide provides enterprise-grade configuration patterns for serverless environments on Kubernetes, focusing on advanced integrations between Prometheus monitoring and KEDA autoscaling. The architecture delivers optimal resource efficiency through event-driven scaling while maintaining observability and resilience for AI/ML workloads and other latency-sensitive applications.

## Concepts

### Prometheus Configuration

Prometheus is utilized for monitoring and alerting purposes. To enable cross-namespace ServiceMonitor discovery, configure the `namespaceSelector`. In Prometheus, define the `serviceMonitorSelector` to associate with ServiceMonitors.

```yaml
apiVersion: monitoring.coreos.com/v1
kind: ServiceMonitor
metadata:
  name: qwen2-0--5b-lb-monitor
  namespace: llmaz-system
  labels:
    control-plane: controller-manager
    app.kubernetes.io/name: servicemonitor
spec:
  namespaceSelector:
    any: true
  selector:
    matchLabels:
      llmaz.io/model-name: qwen2-0--5b
  endpoints:
    - port: http
      path: /metrics
      scheme: http
```

- Ensure the `namespaceSelector` is configured to allow cross-namespace monitoring.
- Appropriately label your services to be discovered by Prometheus.

### KEDA Configuration

KEDA (Kubernetes Event-driven Autoscaling) is employed for scaling applications based on custom metrics. It can be integrated with Prometheus to trigger scaling actions.

```yaml
apiVersion: keda.sh/v1alpha1
kind: ScaledObject
metadata:
  name: qwen2-0--5b-scaler
  namespace: default
spec:
  scaleTargetRef:
    apiVersion: inference.llmaz.io/v1alpha1
    kind: Playground
    name: qwen2-0--5b
  pollingInterval: 30
  cooldownPeriod: 50
  minReplicaCount: 0
  maxReplicaCount: 3
  triggers:
  - type: prometheus
    metadata:
      serverAddress: http://prometheus-operated.llmaz-system.svc.cluster.local:9090
      metricName: llamacpp:requests_processing
      query: sum(llamacpp:requests_processing)
      threshold: "0.2"
```

- Ensure the `serverAddress` correctly points to the Prometheus service.
- Adjust `pollingInterval` and `cooldownPeriod` to optimize scaling behavior and prevent conflicts with other scaling mechanisms.

### Integration with Activator

Consider integrating the serverless configuration with an activator for scale-from-zero scenarios. The activator can be implemented using a controller pattern or as a standalone goroutine.

Key Architecture Components:
- Request Interception: Capture incoming requests to scaled-to-zero services
- Pre-Scale Trigger: Initiate scale-up before forwarding requests
- Request Buffering: Queue requests during cold start period
- Event-Driven Scaling: Integrate with KEDA using CloudEvents:

### Controller Runtime Framework

The Controller Runtime framework simplifies the development of Kubernetes controllers by providing abstractions for managing resources and handling events.

#### Key Components

1. **Controller**: Monitors resource states and triggers actions to align actual and desired states.
2. **Reconcile Function**: Contains the core logic for transitioning resource states.
3. **Manager**: Manages the lifecycle of controllers and shared resources.
4. **Client**: Interface for interacting with the Kubernetes API.
5. **Scheme**: Registry for resource types.
6. **Event Source and Handler**: Define event sources and handling logic.

## Quick Start Guide

1. Install Prometheus and KEDA using Helm charts, following the official documentation [Install Guide](https://llmaz.inftyai.com/docs/getting-started/installation/).

```bash
helm install llmaz oci://registry-1.docker.io/inftyai/llmaz --namespace llmaz-system --create-namespace --version 0.0.10
make install-keda
make install-prometheus
```

2. Create a ServiceMonitor for Prometheus to discover your services.

```bash
kubectl apply -f service-monitor.yaml
```

3. Create a ScaledObject for KEDA to manage scaling.

```bash
kubectl apply -f scaled-object.yaml
```

4. Test with a cold start application.

```bash
kubectl exec -it -n kube-system deploy/activator -- wget -O- qwen2-0--5b-lb.default.svc:8080
```

5. Use Prometheus and KEDA dashboards to monitor metrics and scaling activities via web pages.

```bash
kubectl port-forward services/prometheus-operated 9090:9090 --address 0.0.0.0 -n llmaz-system
```

## Benchmark

Cold start latency is a critical metric for evaluating user experience in llmaz Serverless environments. To assess performance stability and efficiency, we conducted rigorous testing under different instance scaling scenarios. The testing methodology included:

| Scaling Pattern | Avg. Latency (s) | P90 Latency (s) | Resource Initialization | Optimization Potential |
|-----------------|------------------|-----------------|-------------------------|-------------------------|
| **0 -> 1**       | 29               | 31              | Full pod creation<br>Image pull<br>Engine initialization | Pre-fetching<br>Snapshot restore |
| **1 -> 2**       | 15               | 16              | Partial image reuse<br>Network reuse<br>Pod creation | Warm pool<br>Priority scheduling |
| **2 -> 3**       | 11               | 12              | Cached dependencies<br>Parallel scheduling<br>Shared resources | Predictive scaling<br>Node affinity |

## Conclusion

This configuration guide offers a detailed approach to setting up a serverless environment with Kubernetes, Prometheus, and KEDA. By adhering to these guidelines, you can ensure efficient scaling and monitoring of your applications.
