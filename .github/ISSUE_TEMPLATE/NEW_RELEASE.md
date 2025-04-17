---
name: New Release
about: Propose a new release
title: Release v0.x.0
labels: ""
assignees: ""
---

## Release Checklist

<!--
Please do not remove items from the checklist
-->

- [ ] All [OWNERS](https://github.com/inftyai/llmaz/blob/main/OWNERS) must LGTM the release proposal
- [ ] Prepare the image and files
  - [ ] Run `PLATFORMS=linux/amd64 make image-push GIT_TAG=$VERSION` to build and push an image.
  - [ ] Run `make artifacts GIT_TAG=$VERSION` to generate the artifact.
- [ ] Update helm chats and documents
  - [ ] Update versions
    - [ ] update `chart/Chart.yaml`, the helm version is different with the app version.
    - [ ] update `docs/installation.md`
  - [ ] Run `make helm-package` to package the helm chart and update the index.yaml.
  - [ ] Submit a PR and merge it.
- [ ] An OWNER [prepares a draft release](https://github.com/inftyai/llmaz/releases)
  - [ ] Create a new tag
  - [ ] Write the change log into the draft release which should include below items if any:
    ```
    🚀 **Major Features**:
    ✨ **Features**:
    🐛 **Bugs**:
    ♻️ **Cleanups**:
    ```
  - [ ] Upload the files to the draft release.
    - [ ] `manifests.yaml` under artifacts
    - [ ] new generated helm chart `*.zip` file
- [ ] Publish the draft release prepared at the [Github releases page](https://github.com/inftyai/llmaz/releases)
- [ ] Publish the helm chart
  - [ ] Run `git checkout gh-pages`
  - [ ] Copy the `index.yaml` from main branch
  - [ ] Submit a PR and merge it.
- [ ] Close this issue
