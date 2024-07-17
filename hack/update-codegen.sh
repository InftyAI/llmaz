#!/usr/bin/env bash

set -o errexit
set -o nounset
set -o pipefail

cd "$(dirname "${0}")/.."
GO_CMD=${1:-go}
CODEGEN_PKG=${2:-bin}
REPO_ROOT="$(git rev-parse --show-toplevel)"


source "${CODEGEN_PKG}/kube_codegen.sh"

# TODO: remove the workaround when the issue is solved in the code-generator
# (https://github.com/kubernetes/code-generator/issues/165).
# Here, we create the soft link named "x-k8s.io" to the parent directory of
# LeaderWorkerSet to ensure the layout required by the kube_codegen.sh script.
ln -s .. inftyai.com
trap "rm inftyai.com" EXIT

kube::codegen::gen_helpers \
    --input-pkg-root inftyai.com/llmaz/api \
    --output-base "${REPO_ROOT}" \
    --boilerplate "${REPO_ROOT}/hack/boilerplate.go.txt"

kube::codegen::gen_client \
    --with-watch \
    --with-applyconfig \
    --input-pkg-root inftyai.com/llmaz/api \
    --output-base "$REPO_ROOT" \
    --output-pkg-root inftyai.com/llmaz/client-go \
    --boilerplate "${REPO_ROOT}/hack/boilerplate.go.txt"
