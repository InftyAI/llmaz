# Examples

We provide a bunch of examples to serve large language models.

## Table of Contents

- [Deploy models from Huggingface](#deploy-models-from-huggingface)
- [Deploy models from ModelScope](#deploy-models-from-modelscope)
- [Deploy models serving by SGLang](#deploy-models-serving-by-sglang)

### Deploy models from Huggingface

Deploy models hosted in Huggingface, see [example](./vllm-huggingface/) here.

In theory, we support any size of model. However, the bandwidth is limited. For example, we want to load the `llama2-7B` model, which takes about 15GB memory size, if we have a 200 Mbps bandwidth, it will take about 10mins to download the model, so the bandwidth plays a vital role here.

### Deploy models from ModelScope

Deploy models hosted in ModelScope, see [example](./vllm-modelscope/) here.

### Deploy models serving by SGLang

By default, we use [vLLM](https://github.com/vllm-project/vllm) as the inference backend, however, if you want to use other backends like [SGLang](https://github.com/sgl-project/sglang), see [examples](./sglang/) here.
