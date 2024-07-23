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

from huggingface_hub import hf_hub_download, list_repo_files

from loader.model_hub.model_hub import ModelHub, MODEL_LOCAL_DIR


HUGGING_FACE = "Huggingface"


class Huggingface(ModelHub):
    @classmethod
    def name(cls) -> str:
        return HUGGING_FACE

    @classmethod
    def load_model(cls, model_id: str) -> None:
        # TODO: Should we verify the download is finished?
        for file in list_repo_files(repo_id=model_id):
            path = hf_hub_download(
                repo_id=model_id,
                filename=file,
                local_dir=MODEL_LOCAL_DIR + model_id.replace("/", "--"),
            )
            print(f"file path is {path}")
