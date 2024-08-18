#!/usr/bin/env bash

# Copyright 2024 The Kubernetes Authors.
#
# Licensed under the Apache License, Version 2.0 (the "License");
# you may not use this file except in compliance with the License.
# You may obtain a copy of the License at
#
#     http://www.apache.org/licenses/LICENSE-2.0
#
# Unless required by applicable law or agreed to in writing, software
# distributed under the License is distributed on an "AS IS" BASIS,
# WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
# See the License for the specific language governing permissions and
# limitations under the License.

set -o errexit
set -o nounset
set -o pipefail

export CWD=$(pwd)
function cleanup {
    if [ $USE_EXISTING_CLUSTER == 'false' ]
    then
        $KIND delete cluster --name $KIND_CLUSTER_NAME
    fi
    (cd $CWD/config/manager && $KUSTOMIZE edit set image controller=inftyai/llmaz:main)
}
function startup {
    if [ $USE_EXISTING_CLUSTER == 'false' ]
    then
        $KIND create cluster --name $KIND_CLUSTER_NAME --image $E2E_KIND_VERSION --config ./hack/kind-config.yaml
    fi
}
function kind_load {
    $KIND load docker-image $IMAGE_TAG --name $KIND_CLUSTER_NAME
}
function deploy {
    cd $CWD/config/manager && $KUSTOMIZE edit set image controller=$IMAGE_TAG
    $KUSTOMIZE build $CWD/test/e2e/config | $KUBECTL apply --server-side -f -
}
function deploy_lws {
    $KUBECTL apply --server-side -f https://github.com/kubernetes-sigs/lws/releases/download/$ENVTEST_LWS_VERSION/manifests.yaml
}
trap cleanup EXIT
startup
kind_load
deploy_lws
deploy
$GINKGO -v $CWD/test/e2e/...
