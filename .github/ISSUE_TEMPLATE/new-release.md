---
name: New Release
about: Propose a new release
title: Release v0.x.0
labels: ''
assignees: ''

---

## Release Checklist
<!--
Please do not remove items from the checklist
-->
- [ ] All [OWNERS](https://github.com/inftyai/llmaz/blob/main/OWNERS) must LGTM the release proposal
- [ ] Verify that the changelog in this issue is up-to-date
- [ ] Prepare the image and files
  - [ ] Run `PLATFORMS=linux/amd64 make image-push GIT_TAG=$VERSION`  to build and push an image.
  - [ ] Run `make artifacts GIT_TAG=$VERSION` to generate the artifact.
  - [ ] Run `make helm-package` to package the helm chart and update the index.yaml.
- [ ] Update `docs/installation.md`, `chart/README.md`
  - [ ] Submit a PR and merge it.
- [ ] An OWNER [prepares a draft release](https://github.com/inftyai/llmaz/releases)
  - [ ] Create a new tag
  - [ ] Write the change log into the draft release
  - [ ] Upload the files to the draft release.
      - `manifests.yaml` under artifacts
      - new generated helm chart `*.zip` file
- [ ] Publish the draft release prepared at the [Github releases page](https://github.com/inftyai/llmaz/releases)
- [ ] Publish the helm chart
  - [ ] Run `git checkout gh-pages`
  - [ ] Copy the `index.yaml` from main branch
  - [ ] Submit a PR and merge it.
- [ ] Close this issue

## Changelog
<!--
Describe changes since the last release here.
-->

🚀 **Major Features**:

✨ **Features**:

🐛 **Bugs**:

♻️ **Cleanups**: