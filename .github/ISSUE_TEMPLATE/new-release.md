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
- [ ] For major or minor releases (v$MAJ.$MIN.0), create a new release branch.
  - [ ] an OWNER creates a vanilla release branch with
        `git branch release-$MAJ.$MIN main`
  - [ ] An OWNER pushes the new release branch with
        `git push release-$MAJ.$MIN`
- [ ] Update things like README, deployment templates, docs, configuration, test/e2e flags.
      Submit a PR against the release branch:
- [ ] An OWNER [prepares a draft release](https://github.com/inftyai/llmaz/releases)
  - [ ] Write the change log into the draft release.
  - [ ] Run
      `make artifacts GIT_TAG=$VERSION`
      to generate the artifacts and upload the files in the `artifacts` folder
      to the draft release.
- [ ] An OWNER creates a signed tag running
     `git tag -s $VERSION`
      and inserts the changelog into the tag description.
      To perform this step, you need [a PGP key registered on github](https://docs.github.com/en/authentication/managing-commit-signature-verification/checking-for-existing-gpg-keys).
- [ ] An OWNER pushes the tag with
      `git push $VERSION`
  - Publish a staging container image
      make image-push GIT_TAG=$VERSION
- [ ] Publish the draft release prepared at the [Github releases page](https://github.com/inftyai/llmaz/releases).
- [ ] Add a link to the tagged release in this issue: <!-- example https://github.com/inftyai/llmaz/releases/tag/v0.1.0 -->
- [ ] For a major or minor release, update `README.md` and `docs/setup/install.md`
      in `main` branch:
- [ ] Close this issue

## Changelog
<!--
Describe changes since the last release here.
-->
