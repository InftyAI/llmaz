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
import shutil

from llmaz.model_loader.constant import MODEL_LOCAL_DIR
from llmaz.model_loader.oci import llmman
from llmaz.util.logger import Logger


def model_download(reference: str, out_dir: str = MODEL_LOCAL_DIR):
    """Acquire a CNCF ModelPack artifact through a running `llmman serve`.

    The daemon does the pull (POST /api/pull, streamed so a multi-gigabyte
    fetch is not silent) but deliberately exposes no local path, so
    `llmman resolve --no-pull` reports where the bytes landed.
    """
    if not reference or not reference.strip():
        raise ValueError("OCI reference cannot be empty")
    reference = reference.strip()

    def _progress(status, completed, total):
        if total:
            Logger.info(f"llmman: {status} ({completed}/{total} bytes)")
        else:
            Logger.info(f"llmman: {status}")

    resolved = llmman.pull_and_resolve(reference, progress=_progress)
    materialize(resolved, out_dir)
    Logger.info(f"placed {reference} at {out_dir}")


def materialize(src: str, out_dir: str):
    """Place llmman's extracted model at out_dir.

    Files are hard-linked where possible so a model shared with llmman's store
    costs its bytes once, falling back to a copy across filesystems.
    """
    os.makedirs(out_dir, exist_ok=True)

    if not os.path.isdir(src):
        link_or_copy(src, os.path.join(out_dir, os.path.basename(src)))
        return

    for root, _, files in os.walk(src):
        rel = os.path.relpath(root, src)
        dest_root = out_dir if rel == "." else os.path.join(out_dir, rel)
        os.makedirs(dest_root, exist_ok=True)
        for name in files:
            link_or_copy(os.path.join(root, name), os.path.join(dest_root, name))


def link_or_copy(src: str, dest: str):
    if os.path.lexists(dest):
        os.remove(dest)
    try:
        os.link(src, dest)
    except OSError:
        # Different filesystem, or one without hard links.
        shutil.copy2(src, dest)
