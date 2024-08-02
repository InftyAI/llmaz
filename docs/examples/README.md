# Examples

We provide a bunch of examples to serve large language models.

## Table of Contents

- [Deploy Models via Huggingface](#deploy-models-via-huggingface)

### Deploy models via Huggingface

Deploy a small [model](./huggingface/model.yaml) hosted in Huggingface, see the [example](./huggingface/playground.yaml) here. Because the model size is small, the loading time is acceptable.

In theory, we support any size of model, however, the bandwidth is limited, for example, we want to load the `llama2-7B` model, which takes about 15GB memory size, if we have a 200 Mbps bandwidth, it will take about 10mins to download the model, so the bandwidth plays a vital role here.
