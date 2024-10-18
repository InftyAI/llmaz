"""
Copyright 2024.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
"""

import os
from datetime import datetime

from llmaz.model_loader.constant import *

from llmaz.model_loader.objstore.objstore import model_download
from llmaz.model_loader.model_hub.hub_factory import HubFactory
from llmaz.model_loader.model_hub.huggingface import HUB_HUGGING_FACE
from llmaz.util.logger import Logger

if __name__ == "__main__":
    model_source_type = os.getenv(ENV_HUB_MODEL_SOURCE_TYPE)
    start_time = datetime.now()

    if model_source_type == "modelhub":
        hub_name = os.getenv(ENV_HUB_MODEL_HUB_NAME, HUB_HUGGING_FACE)
        revision = os.getenv(ENV_HUB_REVISION)
        model_id = os.getenv(ENV_HUB_MODEL_ID)
        model_file_name = os.getenv(ENV_HUB_MODEL_FILENAME)
        model_allow_patterns = os.getenv(ENV_HUB_MODEL_ALLOW_PATTERNS)
        model_ignore_patterns = os.getenv(ENV_HUB_MODEL_IGNORE_PATTERNS)

        if not model_id:
            raise EnvironmentError(f"Environment variable '{model_id}' not found.")
        hub = HubFactory.new(hub_name)
        model_allow_patterns_list, model_ignore_patterns_list = [], []
        if model_allow_patterns:
            model_allow_patterns_list = model_allow_patterns.split(',')
        if model_ignore_patterns:
            model_ignore_patterns_list = model_ignore_patterns.split(',')
        hub.load_model(model_id, model_file_name, revision, model_allow_patterns_list, model_ignore_patterns_list)
    elif model_source_type == "objstore":
        provider = os.getenv(ENV_OBJ_PROVIDER)
        endpoint = os.getenv(ENV_OBJ_ENDPOINT)
        bucket = os.getenv(ENV_OBJ_BUCKET)
        src = os.getenv(ENV_OBJ_MODEL_PATH)

        model_download(provider=provider, endpoint=endpoint, bucket=bucket, src=src)
    else:
        raise EnvironmentError(f"unknown model source type {model_source_type}")

    Logger.info(
        f"loading models from {model_source_type} takes {(datetime.now() - start_time).total_seconds()}s"
    )
