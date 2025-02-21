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
# Here, we create the soft link named "github.com" to the parent directory of
# llmaz to ensure the layout required by the kube_codegen.sh script.
mkdir -p github.com && ln -s ../.. github.com/inftyai
trap "rm -r github.com" EXIT

kube::codegen::gen_helpers github.com/inftyai/llmaz/api \
    --boilerplate "${REPO_ROOT}/hack/boilerplate.go.txt"

kube::codegen::gen_client github.com/inftyai/llmaz/api \
    --with-watch \
    --with-applyconfig \
    --output-dir "$REPO_ROOT"/client-go \
    --output-pkg github.com/inftyai/llmaz/client-go \
    --boilerplate "${REPO_ROOT}/hack/boilerplate.go.txt"
