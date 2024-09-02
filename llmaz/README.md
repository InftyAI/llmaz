# llmaz

ModelLoader maintains the codes to load model weights with various ways, such as from Huggingface or from object stores.

## Load Models From ModelHub

The full list of supported model hubs:

- [Huggingface](https://huggingface.co/welcome)
- [ModelScope](https://www.modelscope.cn/home)

## Load Models From ObjectStore

The full list of supported object store providers:

- AlibabaCloud OSS
<!-- - AWS S3
- Azure Storage Account
- Google Cloud Storage
- MinIO or other S3-compatible object storages
- Tencent COS -->

## How to use

The model loader will be build as an image to provide services, it will be part of the initContainer of the inference Service.

### How to build the image

- `make loader-image-load` will build the image locally.
- `make loader-image-push` will push the image to the registry, default to `inftyai/model-loader:<version>`.

## Test

Run `make pytest` to make sure tests passed.
