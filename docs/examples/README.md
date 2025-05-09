# Examples

We provide a set of examples to help you serve large language models, by default, we use vLLM as the backend.

## Table of Contents

- [Deploy models from Huggingface](#deploy-models-from-huggingface)
- [Deploy models from ModelScope](#deploy-models-from-modelscope)
- [Deploy models from ObjectStore](#deploy-models-from-objectstore)
- [Deploy models via SGLang](#deploy-models-via-sglang)
- [Deploy models via llama.cpp](#deploy-models-via-llamacpp)
- [Deploy models via TensorRT-LLM](#deploy-models-via-tensorrt-llm)
- [Deploy models via text-generation-inference](#deploy-models-via-text-generation-inference)
- [Deploy models via ollama](#deploy-models-via-ollama)
- [Speculative Decoding with llama.cpp](#speculative-decoding-with-llamacpp)
- [Speculative Decoding with vLLM](#speculative-decoding-with-vllm)
- [Multi-Host Inference](#multi-host-inference)
- [Deploy Host Models](#deploy-host-models)
- [Envoy AI Gateway](#envoy-ai-gateway)

### Deploy models from Huggingface

Deploy models hosted in Huggingface, see [example](./huggingface/) here.

> Note: if your model needs Huggingface token for weight downloads, please run `kubectl create secret generic modelhub-secret --from-literal=HF_TOKEN=<your token>` ahead.

In theory, we support any size of model. However, the bandwidth is limited. For example, we want to load the `llama2-7B` model, which takes about 15GB memory size, if we have a 200Mbps bandwidth, it will take about 10mins to download the model, so the bandwidth plays a vital role here.

### Deploy models from ModelScope

Deploy models hosted in ModelScope, see [example](./modelscope/) here, similar to other backends.

### Deploy models from ObjectStore

Deploy models stored in object stores, we support various providers, see the full list below.

In theory, if we want to load the `Qwen2-7B` model, which occupies about 14.2 GB memory size, and the intranet bandwidth is about 800Mbps, it will take about 2 ~ 3 minutes to download the model. However, the intranet bandwidth can be improved.

- Alibaba Cloud OSS, see [example](./objstore-oss/) here

  > Note: you should set OSS_ACCESS_KEY_ID and OSS_ACCESS_kEY_SECRET first by running `kubectl create secret generic oss-access-secret --from-literal=OSS_ACCESS_KEY_ID=<your ID> --from-literal=OSS_ACCESS_kEY_SECRET=<your secret>`

### Deploy models via SGLang

By default, we use [vLLM](https://github.com/vllm-project/vllm) as the inference backend, however, if you want to use other backends like [SGLang](https://github.com/sgl-project/sglang), see [example](./sglang/) here.

### Deploy models via llama.cpp

[llama.cpp](https://github.com/ggerganov/llama.cpp) can serve models on a wide variety of hardwares, such as CPU, see [example](./llamacpp/) here.

### Deploy models via TensorRT-LLM

[TensorRT-LLM](https://github.com/NVIDIA/TensorRT-LLM) provides users with an easy-to-use Python API to define Large Language Models (LLMs) and support state-of-the-art optimizations to perform inference efficiently on NVIDIA GPUs, see [example](./tensorrt-llm/) here.

### Deploy models via text-generation-inference

[text-generation-inference](https://github.com/huggingface/text-generation-inference) is used in production at Hugging Face to power Hugging Chat, the Inference API and Inference Endpoint. see [example](./tgi/) here.

### Deploy models via ollama

[ollama](https://github.com/ollama/ollama) based on llama.cpp, aims for local deploy. see [example](./ollama/) here.

### Speculative Decoding with llama.cpp

llama.cpp supports speculative decoding to significantly improve inference performance, see [example](./speculative-decoding/llamacpp/) here.

### Speculative Decoding with vLLM

[Speculative Decoding](https://arxiv.org/abs/2211.17192) can improve inference performance efficiently, see [example](./speculative-decoding/vllm/) here.

### Multi-Host Inference

Model size is growing bigger and bigger, Llama 3.1 405B FP16 LLM requires more than 750 GB GPU for weights only, leaving kv cache unconsidered, even with 8 x H100 Nvidia GPUs, 80 GB size of HBM each, can not fit in a single host, requires a multi-host deployment, see [example](./multi-nodes/) here.

### Deploy Host Models

Models could be loaded in prior to the hosts, especially those extremely big models, see [example](./hostpath/) to serve local models.

### Envoy AI Gateway

llmaz leverages envoy AI gateway as default API gateway, see how it works [here](../envoy-ai-gateway.md).
