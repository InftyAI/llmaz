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

import sys

from pkg.loader.model_hub.model_hub import HubFactory

if __name__ == "__main__":
    if len(sys.argv) != 3:
        raise Exception("args number not right")

    hub_name = sys.argv[1]
    model_name = sys.argv[2]

    hub = HubFactory(hub_name=hub_name, model_name=model_name)
    hub.load_model()
