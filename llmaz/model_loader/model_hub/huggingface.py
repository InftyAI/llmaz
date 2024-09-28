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

import concurrent.futures
import os

from huggingface_hub import snapshot_download

from llmaz.model_loader.constant import MODEL_LOCAL_DIR, HUB_HUGGING_FACE
from llmaz.model_loader.model_hub.model_hub import (
    ModelHub,
)
from llmaz.util.logger import Logger
from llmaz.model_loader.model_hub.util import get_folder_total_size

from typing import Optional, List


class Huggingface(ModelHub):
    @classmethod
    def name(cls) -> str:
        return HUB_HUGGING_FACE

    @classmethod
    def load_model(
            cls,
            model_id: str,
            filename: Optional[str],
            revision: Optional[str],
            allow_patterns: Optional[List[str]],
            ignore_patterns: Optional[List[str]],
    ) -> None:
        Logger.info(
            f"Start to download, model_id: {model_id}, filename: {filename}, revision: {revision}"
        )

        local_dir = os.path.join(
            MODEL_LOCAL_DIR, f"models--{model_id.replace('/', '--')}"
        )

        if filename:
            allow_patterns.append(filename)
            local_dir = MODEL_LOCAL_DIR

        snapshot_download(
            repo_id=model_id,
            revision=revision,
            local_dir=local_dir,
            allow_patterns=allow_patterns,
            ignore_patterns=ignore_patterns,
        )

        total_size = get_folder_total_size(local_dir)
        Logger.info(f"The total size of {local_dir} is {total_size: .2f} GB")