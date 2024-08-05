# Examples

We provide a bunch of examples to serve large language models.

## Table of Contents

- [Deploy models via Huggingface](#deploy-models-via-huggingface)
- [Deploy models via ModelScope](#deploy-models-via-modelscope)

### Deploy models via Huggingface

Deploy a small [model](./huggingface/model.yaml) hosted in Huggingface, see the [vllm example](./huggingface/vllm-playground.yaml) or [sglang example](./huggingface/sglang-playground.yaml) here.

### Deploy models via ModelScope 

Deploy a small [model](./modelscope/model.yaml) hosted in ModelScope, see the [vllm example](./modelscope/vllm-playground.yaml) or [sglang example](./modelscope/sglang-playground.yaml) here.


We choose `opt-125m` as example model because the model size is small, the bootstrap loading time is acceptable.

In theory, we support any size of model. However, the bandwidth is limited. For example, we want to load the `llama2-7B` model, which takes about 15GB memory size, if we have a 200 Mbps bandwidth, it will take about 10mins to download the model, so the bandwidth plays a vital role here.
