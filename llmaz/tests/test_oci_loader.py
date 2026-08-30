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
from unittest import mock

import pytest

from llmaz.model_loader.oci.oci import link_or_copy, materialize, model_download


class TestModelDownload:
    def test_rejects_an_empty_reference(self):
        for ref in ("", "   "):
            with pytest.raises(ValueError, match="cannot be empty"):
                model_download(ref, out_dir="/tmp/unused")

    def test_pulls_through_the_daemon_then_materializes(self, tmp_path):
        store = tmp_path / "store"
        store.mkdir()
        (store / "model.safetensors").write_text("w")
        out = str(tmp_path / "out")

        with mock.patch(
            "llmaz.model_loader.oci.oci.llmman.pull_and_resolve",
            return_value=str(store),
        ) as acquire:
            model_download("ghcr.io/org/model:tag", out_dir=out)

        # The daemon receives the bare reference; progress is forwarded.
        assert acquire.call_args[0][0] == "ghcr.io/org/model:tag"
        assert acquire.call_args[1]["progress"] is not None
        assert open(os.path.join(out, "model.safetensors")).read() == "w"


class TestMaterialize:
    def test_hard_links_a_directory(self, tmp_path):
        src = tmp_path / "store"
        (src / "sub").mkdir(parents=True)
        (src / "config.json").write_text("{}")
        (src / "sub" / "model.safetensors").write_text("w")
        out = str(tmp_path / "out")

        materialize(str(src), out)

        assert open(os.path.join(out, "sub", "model.safetensors")).read() == "w"
        # A model shared with llmman's store should cost its bytes once.
        assert (
            os.stat(os.path.join(out, "config.json")).st_ino
            == os.stat(str(src / "config.json")).st_ino
        )

    def test_handles_a_single_file_payload(self, tmp_path):
        # A GGUF payload resolves to the file itself, not a directory.
        src = tmp_path / "model.gguf"
        src.write_text("gguf")
        out = str(tmp_path / "out")

        materialize(str(src), out)

        assert open(os.path.join(out, "model.gguf")).read() == "gguf"

    def test_overwrites_a_stale_destination(self, tmp_path):
        src = tmp_path / "store"
        src.mkdir()
        (src / "config.json").write_text("new")
        out = tmp_path / "out"
        out.mkdir()
        (out / "config.json").write_text("stale")

        materialize(str(src), str(out))

        assert (out / "config.json").read_text() == "new"

    def test_falls_back_to_copy_across_filesystems(self, tmp_path):
        src = tmp_path / "store"
        src.mkdir()
        (src / "config.json").write_text("{}")
        out = str(tmp_path / "out")

        with mock.patch(
            "llmaz.model_loader.oci.oci.os.link", side_effect=OSError("EXDEV")
        ):
            materialize(str(src), out)

        assert open(os.path.join(out, "config.json")).read() == "{}"
        assert (
            os.stat(os.path.join(out, "config.json")).st_ino
            != os.stat(str(src / "config.json")).st_ino
        )


def test_link_or_copy_replaces_an_existing_target(tmp_path):
    src = tmp_path / "a"
    src.write_text("new")
    dest = tmp_path / "b"
    dest.write_text("stale")

    link_or_copy(str(src), str(dest))

    assert dest.read_text() == "new"
