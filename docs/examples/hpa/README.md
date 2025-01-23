# How to autoscaling Playgrounds

## Install the Metric Server

HPA depends on the metric-server for scaling decisions, so we need to install it in prior, see install command below:

```cmd
kubectl apply -f https://github.com/kubernetes-sigs/metrics-server/releases/latest/download/components.yaml
```

## How to Use

Set the Playground ElasticConfig like this:

```yaml
spec:
  elasticConfig:
    minReplicas: 1
    maxReplicas: 3
```

If your backendRuntime has already configured the `ScalePolicy`, then it's working now. If not, you can set the scalingPolicy directly in Playground like this:

```yaml
spec:
  elasticConfig:
    minReplicas: 1
    maxReplicas: 3
    scalePolicy:
      hpa:
        metrics:
          - type: Resource
            resource:
              name: cpu
              target:
                type: Utilization
                averageUtilization: 50
```
