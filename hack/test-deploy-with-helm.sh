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
    HELM_EXT_OPTS='--namespace=llmaz-system --create-namespace --set controllerManager.manager.image.tag=${LOADER_IMAGE_TAG}' make helm-install
    $KUBECTL wait --timeout=3m --for=condition=ready pods --namespace=llmaz-system -l app!=certgen
    echo "all pods of llmaz-system is ready..."
    $KUBECTL get pod -n llmaz-system
}
function deploy_kube_prometheus {
    LATEST=$(curl -s https://api.github.com/repos/prometheus-operator/prometheus-operator/releases/latest | jq -cr .tag_name)
    curl -sL https://github.com/prometheus-operator/prometheus-operator/releases/download/${LATEST}/bundle.yaml | $KUBECTL  create -f -
}
trap cleanup EXIT
startup
kind_load
deploy_kube_prometheus
deploy
