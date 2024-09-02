# Examples

We provide a set of examples to help you serve large language models, by default, we use vLLM as the backend.

## Table of Contents

- [Deploy models from Huggingface](#deploy-models-from-huggingface)
- [Deploy models from ModelScope](#deploy-models-from-modelscope)
- [Deploy models from ObjectStore](#deploy-models-from-objectstore)
- [Deploy models via SGLang](#deploy-models-via-sglang)
- [Deploy models via llama.cpp](#deploy-models-via-llamacpp)
- [Speculative Decoding with vLLM](#speculative-decoding-with-vllm)

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

### Speculative Decoding with vLLM

[Speculative Decoding](https://arxiv.org/abs/2211.17192) can improve inference performance efficiently, see [example](./speculative-decoding/vllm/) here.
