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

from loader.model_hub.model_hub import ModelHub, MODEL_LOCAL_DIR

MODEL_SCOPE = "ModelScope"
MAX_WORKERS = 4


class ModelScope(ModelHub):
    @classmethod
    def name(cls) -> str:
        return MODEL_SCOPE

    @classmethod
    def load_model(cls, model_id: str, revision: Optional[str]) -> None:
        print(f"Start to download model {model_id}")

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
    print(f"Download completed for {filename}")
