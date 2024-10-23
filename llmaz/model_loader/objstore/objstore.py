from omnistore.objstore import StoreFactory

from llmaz.model_loader.constant import MODEL_LOCAL_DIR


def model_download(provider: str, endpoint: str, bucket: str, src: str):
    client = StoreFactory.new_client(
        provider=provider, endpoint=endpoint, bucket=bucket
    )

    model_name = src.split("/")[-1]
    # Such as GGUF model
    if "." in model_name:
        client.download(src, MODEL_LOCAL_DIR + model_name)
    else:
        client.download_dir(src, MODEL_LOCAL_DIR + "models--" + model_name)
