# Horizontal Scaling With Playgrounds

We only support HPA right now, but will try to integrate with KEDA and Knative in the future.

## Install the Metric Server

HPA depends on the metric-server for scaling decisions, so we need to install it in prior, see install command below:

```cmd
kubectl apply -f https://github.com/kubernetes-sigs/metrics-server/releases/latest/download/components.yaml
```

## How to Use

If your backendRuntime has already configured the `ScaleTriggers`, set the `playground.elasticConfig` like this:

```yaml
spec:
  elasticConfig:
    minReplicas: 1
    maxReplicas: 3
```

The 0-index trigger will be applied, or you can specify the trigger like this:

```yaml
spec:
  elasticConfig:
    minReplicas: 1
    maxReplicas: 3
    scaleTriggerRef:
      name: <trigger-name>
```

If you want to change the target values, you can set the scaleTrigger directly in Playground like this:

```yaml
spec:
  elasticConfig:
    minReplicas: 1
    maxReplicas: 3
    scaleTrigger:
      hpa:
        metrics:
          - type: Resource
            resource:
              name: cpu
              target:
                type: Utilization
                averageUtilization: 50
```
