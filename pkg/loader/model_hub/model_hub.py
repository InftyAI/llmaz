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

from abc import ABC, abstractmethod

from pkg.loader.modelhub.huggingface import HUGGING_FACE, Huggingface


MODEL_LOCAL_DIR = "/workspace/models/"
MODEL_CACHE_LOCAL_DIR = "/workspace/cache/models/"


class ModelHub(ABC):
    @classmethod
    @abstractmethod
    def name(cls) -> str:
        pass

    @classmethod
    @abstractmethod
    def load_model(cls, model_name: str) -> bool:
        pass


# TODO: support modelScope.
SUPPORT_MODEL_HUBS = {
    HUGGING_FACE: Huggingface,
}


class HubFactory:
    def __init__(self, hub_name: str, model_name: str) -> ModelHub:
        if model_name not in SUPPORT_MODEL_HUBS.keys():
            raise ValueError(f"Unknown model hub: {hub_name}")

        return SUPPORT_MODEL_HUBS[hub_name]
