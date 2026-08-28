---
title: Serverless
weight: 4
---

## Overview

This comprehensive guide provides enterprise-grade configuration patterns for serverless environments on Kubernetes, focusing on advanced integrations between Prometheus monitoring and KEDA autoscaling. The architecture delivers optimal resource efficiency through event-driven scaling while maintaining observability and resilience for AI/ML workloads and other latency-sensitive applications.

### Relationship Between Activator and KEDA

The serverless architecture combines two complementary components:

- **KEDA (Kubernetes Event-driven Autoscaling)**: Handles dynamic scaling based on metrics from Prometheus or other event sources. KEDA monitors application metrics (such as request queue length, processing time, or custom metrics) and automatically adjusts the number of replicas between `minReplicaCount` and `maxReplicaCount` to meet demand.

- **Activator**: Serves as a request interceptor and buffer for scale-from-zero scenarios. When KEDA scales a workload down to zero replicas (to save resources during idle periods), the activator intercepts incoming requests, triggers KEDA to scale up the workload, buffers the requests during the cold start period, and forwards them once the workload is ready.

Together, these components enable true serverless behavior: workloads can scale to zero when idle (minimizing costs) and automatically scale up on-demand when requests arrive (maintaining responsiveness). KEDA provides the scaling mechanism, while the activator ensures no requests are lost during the scale-from-zero process.

## Concepts

### Prometheus Configuration

Prometheus is utilized for monitoring and alerting purposes in the serverless architecture. It collects and stores metrics from your workloads, which KEDA can then query to make scaling decisions.

#### ServiceMonitor Explained

A ServiceMonitor is a Kubernetes custom resource that tells Prometheus which services to monitor and how to scrape metrics from them. The key configuration aspects are:

- **Cross-namespace Discovery**: The `namespaceSelector` field allows Prometheus to discover and monitor services across different namespaces. Setting `any: true` enables monitoring of services in any namespace, not just the namespace where the ServiceMonitor is deployed.

- **Label Matching**: The `selector.matchLabels` field defines which services to monitor based on their labels. In this example, it monitors all services with the label `llmaz.io/model-name: qwen2-0--5b`.

- **Metrics Endpoint**: The `endpoints` section specifies how to scrape metrics - which port to use, the HTTP path where metrics are exposed, and the protocol (HTTP/HTTPS).

Here's an example ServiceMonitor configuration:

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
      # Service label, you can change this to match your service label
      llmaz.io/model-name: qwen2-0--5b
  endpoints:
    - port: http
      path: /metrics
      scheme: http
```

**Important Configuration Notes:**

- Ensure the `namespaceSelector` is configured to allow cross-namespace monitoring if your services are in different namespaces than your ServiceMonitor.
- Your Kubernetes Services must have matching labels (as specified in `selector.matchLabels`) for Prometheus to discover them.
- The service must expose a `/metrics` endpoint (or the path you specify) that returns metrics in Prometheus format.
- Verify that your Prometheus instance is configured with a `serviceMonitorSelector` that matches the labels on this ServiceMonitor (in this example, `control-plane: controller-manager` and `app.kubernetes.io/name: servicemonitor`).

### KEDA Configuration

KEDA (Kubernetes Event-driven Autoscaling) is employed for scaling applications based on custom metrics. It extends Kubernetes' native autoscaling capabilities by allowing you to scale based on external event sources like Prometheus metrics, message queues, or cloud events.

#### ScaledObject Explained

A ScaledObject is the core KEDA resource that defines how to scale your workload. The key configuration aspects are:

- **Scale Target**: The `scaleTargetRef` field specifies which Kubernetes resource to scale. In this example, it targets a custom resource of type `Playground` from the `inference.llmaz.io` API group. KEDA can scale Deployments, StatefulSets, or any custom resource that implements the scale subresource.

- **Replica Boundaries**: The `minReplicaCount` and `maxReplicaCount` fields define the scaling limits. Setting `minReplicaCount: 0` enables scale-to-zero functionality, which requires the activator component to handle cold starts.

- **Timing Parameters**:
  - `pollingInterval` (30s in this example): How frequently KEDA checks the metrics from the trigger source
  - `cooldownPeriod` (50s in this example): The waiting period after the last trigger activation before scaling back down to prevent flapping

- **Prometheus Trigger**: The `triggers` section defines what metrics to monitor and when to scale. The Prometheus trigger queries a specific metric and compares it against a threshold. When the metric value exceeds the threshold, KEDA scales up; when it falls below, KEDA scales down (after the cooldown period).

Here's an example ScaledObject configuration:

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

**Important Configuration Notes:**

- **Prometheus Connection**: Ensure the `serverAddress` correctly points to your Prometheus service. The format is typically `http://<service-name>.<namespace>.svc.cluster.local:<port>`.

- **Metric Query**: The `query` field should be a valid PromQL (Prometheus Query Language) expression. In this example, `sum(llamacpp:requests_processing)` aggregates all processing requests across instances.

- **Threshold Tuning**: The `threshold` value determines when scaling occurs. For example, `"0.2"` means KEDA will scale up when the metric value exceeds 0.2. Adjust this based on your workload characteristics and desired responsiveness.

- **Timing Optimization**:
  - Set `pollingInterval` based on how quickly you need to respond to load changes (shorter intervals = faster response but more API calls)
  - Set `cooldownPeriod` long enough to avoid scaling down too quickly during temporary traffic drops, but short enough to save resources during idle periods
  - Ensure these values don't conflict with HPA (Horizontal Pod Autoscaler) if you're using both

- **Scale-to-Zero Requirements**: When using `minReplicaCount: 0`, you must deploy the activator component to handle requests during cold starts and trigger the scale-up process.

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
