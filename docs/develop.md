# Develop Guidance

A develop guidance for people who want to learn more about this project.

## Project Structure

```structure
llmaz # root
├── llmaz # where the model loader logic locates
├── pkg # where the main logic for Kubernetes controllers locates
```

## API design

### Core APIs

**OpenModel**: `OpenModel` is mostly like to store the open sourced models as a cluster-scope object. We may need namespaced models in the future for tenant isolation. Usually, the cloud provider or model provider should set this object because they know models well, like the accelerators or the scaling primitives.

### Inference APIs

**Playground**: `Playground` is for easy usage, people who has little knowledge about cloud can quick deploy a large language model with minimal configurations. `Playground` is integrated with the SOTA inference engines already, like vLLM.

**Service**: `Service` is the real inference workload, people has advanced configuration requirements can deploy with `Service` directly if `Playground` can not meet their demands like they have a customized inference engine, which hasn't been integrated with llmaz yet. Or they have different topology requirements to align with the Pods.
