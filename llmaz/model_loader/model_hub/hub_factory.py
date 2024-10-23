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
from llmaz.model_loader.constant import HUB_HUGGING_FACE, HUB_MODEL_SCOPE
from llmaz.model_loader.model_hub.model_hub import ModelHub
from llmaz.model_loader.model_hub.huggingface import Huggingface
from llmaz.model_loader.model_hub.modelscope import ModelScope

SUPPORT_MODEL_HUBS = {
    HUB_HUGGING_FACE: Huggingface,
    HUB_MODEL_SCOPE: ModelScope,
}


class HubFactory:

    @classmethod
    def new(cls, hub_name: str) -> ModelHub:
        if hub_name not in SUPPORT_MODEL_HUBS.keys():
            raise ValueError(f"Unknown model hub: {hub_name}")

        return SUPPORT_MODEL_HUBS[hub_name]
