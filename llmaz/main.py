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

logging.basicConfig(
    level=logging.INFO,
    format="%(asctime)s - %(name)s - %(levelname)s - %(message)s",
)
logger = logging.getLogger(__name__)


def get_env_variable(var_name):
    """Retrieve the value of the environment variable or raise an error
    if it does not exist.
    """
    try:
        return os.environ[var_name]
    except KeyError:
        raise EnvironmentError(f"Environment variable '{var_name}' not found.")


if __name__ == "__main__":
    start_time = datetime.now()
    hub_name = get_env_variable("MODEL_HUB_NAME")
    model_id = get_env_variable("MODEL_ID")
    hub = HubFactory.new(hub_name)
    hub.load_model(model_id)

    logger.info(f"loading models takes {datetime.now() - start_time}s")
