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
from json import load
import os
from fileinput import filename
from threading import local

from huggingface_hub import hf_hub_download, list_repo_files

from loader.model_hub.model_hub import ModelHub, MODEL_LOCAL_DIR


HUGGING_FACE = "Huggingface"
MAX_WORKERS = 4
from typing import Optional


class Huggingface(ModelHub):
    @classmethod
    def name(cls) -> str:
        return HUGGING_FACE

    @classmethod
    def load_model(cls, model_id: str, revision: Optional[str]) -> None:
        print(f"Start to download model {model_id}")
        # # TODO: Should we verify the download is finished?
        with concurrent.futures.ThreadPoolExecutor(max_workers=MAX_WORKERS) as executor:
            futures = []
            for file in list_repo_files(repo_id=model_id):
                # TODO: support version management, right now we didn't distinguish with them.
                local_dir = os.path.join(
                    MODEL_LOCAL_DIR, f"models--{model_id.replace('/','--')}"
                )
                futures.append(
                    executor.submit(
                        hf_hub_download,
                        repo_id=model_id,
                        filename=file,
                        local_dir=local_dir,
                        revision=revision,
                    ).add_done_callback(handle_completion)
                )


def handle_completion(future):
    filename = future.result()
    print(f"Download completed for {filename}")
