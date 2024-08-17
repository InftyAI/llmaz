import pytest

from llmaz.model_loader.model_hub.hub_factory import HubFactory
from llmaz.model_loader.model_hub.huggingface import HUGGING_FACE
from llmaz.model_loader.model_hub.modelscope import MODEL_SCOPE


class TestHubFactory:
    def test_new(self):
        test_cases = [
            {
                "hub_name": "Huggingface",
                "expected_hub_name": HUGGING_FACE,
                "should_fail": False,
            },
            {
                "hub_name": "ModelScope",
                "expected_hub_name": MODEL_SCOPE,
                "should_fail": False,
            },
            {
                "hub_name": "unknown",
                "should_fail": True,
            },
        ]

        for tc in test_cases:
            if tc["should_fail"]:
                reason = "Unknown model hub: " + tc["hub_name"]
                with pytest.raises(ValueError, match=reason):
                    hub = HubFactory.new(tc["hub_name"])
            else:
                hub = HubFactory.new(tc["hub_name"])
                assert tc["expected_hub_name"], hub.name()
