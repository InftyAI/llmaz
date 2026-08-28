# Serverless Configuration and Documentation

## Overview

This document provides a detailed guide on configuring serverless environments using Kubernetes, with a focus on integrating Prometheus for monitoring and KEDA for scaling. The configuration aims to ensure efficient resource utilization and seamless scaling of applications.

## Concepts

### Prometheus Configuration

Prometheus is used for monitoring and alerting. To enable cross-namespace ServiceMonitor discovery, use `namespaceSelector`. In Prometheus, define `serviceMonitorSelector` to associate with ServiceMonitors.

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

- Ensure that the `namespaceSelector` is set to allow cross-namespace monitoring.
- Label your services appropriately to be discovered by Prometheus.

### KEDA Configuration

KEDA (Kubernetes Event-driven Autoscaling) is used for scaling applications based on custom metrics. It can be integrated with Prometheus to trigger scaling actions.


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

- Ensure that the `serverAddress` points to the correct Prometheus service.
- Adjust `pollingInterval` and `cooldownPeriod` to optimize scaling behavior and avoid conflicts with other scaling mechanisms.

### Integration with Activator

Consider integrating the serverless configuration with an activator for scale-from-zero scenarios. The activator can be implemented using a controller pattern or as a standalone goroutine.

### Controller Runtime Framework

Using the Controller Runtime framework can simplify the development of Kubernetes controllers. It provides abstractions for managing resources and handling events.

#### Key Components

1. **Controller**: Monitors resource states and triggers actions to align actual and desired states.
2. **Reconcile Function**: Core logic for transitioning resource states.
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

2.  Create a ServiceMonitor for Prometheus to discover your services.
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

5. Check with Prometheus and KEDA dashboards to monitor metrics and scaling activities in web page.
```bash
kubectl port-forward services/prometheus-operated 9090:9090 --address 0.0.0.0 -n llmaz-system
```

## Conclusion

This configuration guide provides a comprehensive approach to setting up a serverless environment with Kubernetes, Prometheus, and KEDA. By following these guidelines, you can ensure efficient scaling and monitoring of your applications.