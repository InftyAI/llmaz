#!/usr/bin/env bash

set -o errexit
set -o nounset
set -o pipefail

REPO_ROOT=$(dirname "${BASH_SOURCE[0]}")/..

export CWD=$(pwd)
function cleanup {
    if [ $USE_EXISTING_CLUSTER == 'false' ]
    then
        $KIND delete cluster --name $KIND_CLUSTER_NAME
    fi
}
function startup {
    if [ $USE_EXISTING_CLUSTER == 'false' ]
    then
        $KIND create cluster --name $KIND_CLUSTER_NAME --image $E2E_KIND_NODE_VERSION --config ./hack/kind-config.yaml
    fi
}
function kind_load {
    $KIND load docker-image $IMAGE_TAG --name $KIND_CLUSTER_NAME
}
function deploy {
    cd $CWD
    HELM_EXT_OPTS='--set controllerManager.manager.image.tag=${TAG}' make helm-install
    $KUBECTL wait --timeout=30m --for=condition=ready pods --namespace=llmaz-system -l app.kubernetes.io/component!=open-webui,app!=certgen
    echo "all pods of llmaz-system is ready..."
    $KUBECTL get pod -n llmaz-system
}
trap cleanup EXIT
startup
kind_load
deploy
