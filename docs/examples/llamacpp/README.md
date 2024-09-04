# llama.cpp

## How to load the model

Models can be loaded from modelHub like Huggingface or object stores:

- From Huggingface

    ```yaml
    source:
      modelHub:
        modelID: Qwen/Qwen2-0.5B-Instruct-GGUF
        filename: qwen2-0_5b-instruct-q5_k_m.gguf
    ```

- From Object Store

    ```yaml
    source:
      uri: oss://llmaz.oss-ap-southeast-1-internal.aliyuncs.com/models/qwen2-0_5b-instruct-q5_k_m.gguf
    ```

## How to test

Once deployed successfully, you can query like this:

- export the service: `kubectl port-forward pod/qwen2-0-5b-0 8080:8080`
- run command:

    ```cmd
    curl --request POST \
    --url http://localhost:8080/v1/completions \
    --header "Content-Type: application/json" \
    --data '{"prompt": "Building a website can be done in 10 simple steps:","n_predict": 128}'
    ```

Then you will see the outputs.
