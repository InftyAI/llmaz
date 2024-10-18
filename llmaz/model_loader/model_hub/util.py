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
from llmaz.util.logger import Logger


def get_folder_total_size(folder_path: str) -> float:
    total_size = 0

    for dirPath, _, filenames in os.walk(folder_path):
        for filename in filenames:
            file_path = os.path.join(dirPath, filename)
            try:
                if os.path.exists(file_path):
                    total_size += os.path.getsize(file_path)
            except OSError as e:
                Logger.error(f"Failed to get file {file_path} size, err is {e}")

    total_size_gb = total_size / (1024 ** 3)
    return total_size_gb
