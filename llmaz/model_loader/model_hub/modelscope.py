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
import concurrent.futures
from typing import Optional

from modelscope import snapshot_download

from llmaz.model_loader.defaults import MODEL_LOCAL_DIR
from llmaz.model_loader.model_hub.model_hub import (
    MAX_WORKERS,
    MODEL_SCOPE,
    ModelHub,
)
from llmaz.util.logger import Logger


class ModelScope(ModelHub):
    @classmethod
    def name(cls) -> str:
        return MODEL_SCOPE

    # TODO: support filename
    @classmethod
    def load_model(
        cls, model_id: str, filename: Optional[str], revision: Optional[str]
    ) -> None:
        Logger.info(
            f"Start to download, model_id: {model_id}, filename: {filename}, revision: {revision}"
        )

        with concurrent.futures.ThreadPoolExecutor(max_workers=MAX_WORKERS) as executor:
            futures = []
            local_dir = os.path.join(
                MODEL_LOCAL_DIR, f"models--{model_id.replace('/','--')}"
            )
            futures.append(
                executor.submit(
                    snapshot_download,
                    model_id=model_id,
                    local_dir=local_dir,
                    revision=revision,
                ).add_done_callback(handle_completion)
            )


def handle_completion(future):
    filename = future.result()
    Logger.info(f"Download completed for {filename}")
