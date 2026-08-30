# Load a model from an OCI registry

A model published as a [CNCF ModelPack](https://github.com/modelpack/model-spec)
artifact can be used as a model source with the `oci://` protocol:

```yaml
source:
  uri: oci://ghcr.io/inftyai/qwen2-0.5b:latest
```

The model-loader initContainer pulls the artifact through a running
[`llmman serve`](https://github.com/llmmanorg/llmman) daemon and places the
files in the model directory, so any inference backend loads it the same way it
would a model fetched from a model hub or an object store.

This reuses the registry, credentials and mirroring a cluster already has for
container images, which is often easier to run air-gapped than a model hub.

## Requirements

The loader talks to an `llmman serve` daemon. By default it uses llmman's own
default address, `127.0.0.1:17434`. To point every loader at one shared daemon
instead, set `LLMAZ_LLMMAN_HOST` on the controller:

```yaml
env:
  - name: LLMAZ_LLMMAN_HOST
    value: llmman.llmaz-system.svc:17434
```

## Private registries

Registry credentials are configured on the llmman daemon rather than on the
model, so one place covers every model pulled through it.

## How to use

```bash
kubectl apply -f playground.yaml
```
