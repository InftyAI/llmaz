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
import logging
from datetime import datetime


from loader.model_hub.hub_factory import HubFactory
from loader.model_hub.huggingface import HUGGING_FACE

logging.basicConfig(
    level=logging.INFO,
    format="%(asctime)s - %(name)s - %(levelname)s - %(message)s",
)
logger = logging.getLogger(__name__)


if __name__ == "__main__":
    hub_name = os.getenv("MODEL_HUB_NAME", HUGGING_FACE)
    revision = os.getenv("REVISION")
    model_id = os.getenv("MODEL_ID")

    if not model_id:
        raise EnvironmentError(f"Environment variable '{model_id}' not found.")

    hub = HubFactory.new(hub_name)

    start_time = datetime.now()
    hub.load_model(model_id, revision)
    logger.info(f"loading models takes {datetime.now() - start_time}s")
