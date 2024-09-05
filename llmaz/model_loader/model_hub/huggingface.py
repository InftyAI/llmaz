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

from huggingface_hub import hf_hub_download, list_repo_files

from llmaz.model_loader.defaults import MODEL_LOCAL_DIR
from llmaz.model_loader.model_hub.model_hub import (
    HUGGING_FACE,
    MAX_WORKERS,
    ModelHub,
)
from llmaz.util.logger import Logger
from llmaz.model_loader.model_hub.util import get_folder_total_size

from typing import Optional


class Huggingface(ModelHub):
    @classmethod
    def name(cls) -> str:
        return HUGGING_FACE

    @classmethod
    def load_model(
        cls, model_id: str, filename: Optional[str], revision: Optional[str]
    ) -> None:
        Logger.info(
            f"Start to download, model_id: {model_id}, filename: {filename}, revision: {revision}"
        )

        if filename:
            hf_hub_download(
                repo_id=model_id,
                filename=filename,
                local_dir=MODEL_LOCAL_DIR,
                revision=revision,
            )
            file_size = os.path.getsize(MODEL_LOCAL_DIR + filename) / (1024**3)
            Logger.info(
                f"The total size of {MODEL_LOCAL_DIR + filename} is {file_size: .2f} GB"
            )
            return

        local_dir = os.path.join(
            MODEL_LOCAL_DIR, f"models--{model_id.replace('/','--')}"
        )

        # # TODO: Should we verify the download is finished?
        with concurrent.futures.ThreadPoolExecutor(max_workers=MAX_WORKERS) as executor:
            futures = []
            for file in list_repo_files(repo_id=model_id):
                # TODO: support version management, right now we didn't distinguish with them.
                futures.append(
                    executor.submit(
                        hf_hub_download,
                        repo_id=model_id,
                        filename=file,
                        local_dir=local_dir,
                        revision=revision,
                    ).add_done_callback(handle_completion)
                )

        total_size = get_folder_total_size(local_dir)
        Logger.info(f"The total size of {local_dir} is {total_size: .2f} GB")


def handle_completion(future):
    filename = future.result()
    Logger.info(f"Download completed for {filename}")
