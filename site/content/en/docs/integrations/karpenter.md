---
title: Karpenter
weight: 2
---

[Karpenter](https://github.com/kubernetes-sigs/karpenter) automatically launches just the right compute resources to handle your cluster's applications, but it is built to adhere to the scheduling decisions of kube-scheduler, so it's certainly possible we would run across some cases where Karpenter makes incorrect decisions when the InftyAI scheduler is in the mix. 

We forked the Karpenter project and re-complie the karpenter image for cloud providers like AWS, and you can find the details in [this proposal](https://github.com/InftyAI/llmaz/blob/main/docs/proposals/106-spot-instance-karpenter/README.md). This document provides deployment steps to install and configure Customized Karpenter in an EKS cluster.

## How to use

### Set environment variables

```shell
export KARPENTER_NAMESPACE="kube-system"
export KARPENTER_VERSION="1.5.0"
export K8S_VERSION="1.32"

export AWS_PARTITION="aws" # if you are not using standard partitions, you may need to configure to aws-cn / aws-us-gov
export CLUSTER_NAME="${USER}-karpenter-demo"
export AWS_DEFAULT_REGION="us-west-2"
export AWS_ACCOUNT_ID="$(aws sts get-caller-identity --query Account --output text)"
export TEMPOUT="$(mktemp)"
export ALIAS_VERSION="$(aws ssm get-parameter --name "/aws/service/eks/optimized-ami/${K8S_VERSION}/amazon-linux-2023/x86_64/standard/recommended/image_id" --query Parameter.Value | xargs aws ec2 describe-images --query 'Images[0].Name' --image-ids | sed -r 's/^.*(v[[:digit:]]+).*$/\1/')"
```

If you open a new shell to run steps in this procedure, you need to set some or all of the environment variables again. To remind yourself of these values, type:

```shell
echo "${KARPENTER_NAMESPACE}" "${KARPENTER_VERSION}" "${K8S_VERSION}" "${CLUSTER_NAME}" "${AWS_DEFAULT_REGION}" "${AWS_ACCOUNT_ID}" "${TEMPOUT}" "${ALIAS_VERSION}"
```

### Create a cluster and add Karpenter

Please refer to the [Getting Started with Karpenter](https://docs.aws.amazon.com/eks/latest/userguide/getting-started-eksctl.html) to create a cluster and add Karpenter.

### Install the gpu operator

```shell
helm repo add nvidia https://helm.ngc.nvidia.com/nvidia \
    && helm repo update
helm install --wait --generate-name \
    -n gpu-operator --create-namespace \
    nvidia/gpu-operator \
    --version=v25.3.0
```

### Install llmaz with InftyAI scheduler enabled

Please refer to [heterogeneous cluster support](../features/heterogeneous-cluster-support.md).

### Configure Karpenter with customized image

We need to assign the `karpenter-core-llmaz` cluster role to the `karpenter` service account and update the karpenter image to the customized one.

```shell
cat <<EOF | envsubst | kubectl apply -f -
apiVersion: rbac.authorization.k8s.io/v1
kind: ClusterRoleBinding
metadata:
  name: karpenter-core-llmaz
roleRef:
  apiGroup: rbac.authorization.k8s.io
  kind: ClusterRole
  name: karpenter-core-llmaz
subjects:
- kind: ServiceAccount
  name: karpenter
  namespace: ${KARPENTER_NAMESPACE}
---
apiVersion: rbac.authorization.k8s.io/v1
kind: ClusterRole
metadata:
  name: karpenter-core-llmaz
rules:
- apiGroups: ["llmaz.io"]
  resources: ["openmodels"]
  verbs: ["get", "list", "watch"]
EOF

helm upgrade --install karpenter oci://public.ecr.aws/karpenter/karpenter --version "${KARPENTER_VERSION}" --namespace "${KARPENTER_NAMESPACE}" --create-namespace \
  --set "settings.clusterName=${CLUSTER_NAME}" \
  --set "settings.interruptionQueue=${CLUSTER_NAME}" \
  --set controller.resources.requests.cpu=1 \
  --set controller.resources.requests.memory=1Gi \
  --set controller.resources.limits.cpu=1 \
  --set controller.resources.limits.memory=1Gi \
  --wait \
  --set controller.image.repository=inftyai/karpenter-provider-aws \
  --set "controller.image.tag=${KARPENTER_VERSION}" \
  --set controller.image.digest=""
```

## Basic Example

1. Create a gpu node pool

```shell
cat <<EOF | envsubst | kubectl apply -f -
apiVersion: karpenter.k8s.aws/v1
kind: EC2NodeClass
metadata:
  name: llmaz-demo            # you can change the name to a more meaningful one, please align with the node pool's nodeClassRef.
spec:
  amiSelectorTerms:
  - alias: al2023@${ALIAS_VERSION}
  blockDeviceMappings:
  # the default volume size of the selected AMI is 20Gi, it is not enough for kubelet to pull
  # the images and run the workloads. So we need to map a larger volume to the root device. 
  # You can change the volume size to a larger value according to your actual needs.
  - deviceName: /dev/xvda
    ebs:
      deleteOnTermination: true
      volumeSize: 50Gi     
      volumeType: gp3
  role: KarpenterNodeRole-${CLUSTER_NAME}          # replace with your cluster name
  securityGroupSelectorTerms:
  - tags:
      karpenter.sh/discovery: ${CLUSTER_NAME}      # replace with your cluster name
  subnetSelectorTerms:
  - tags:
      karpenter.sh/discovery: ${CLUSTER_NAME}      # replace with your cluster name
---
apiVersion: karpenter.sh/v1
kind: NodePool
metadata:
  name: llmaz-demo-gpu-nodepool  # you can change the name to a more meaningful one. 
spec:
  disruption:
    budgets:
    - nodes: 10%
    consolidateAfter: 5m        
    consolidationPolicy: WhenEmptyOrUnderutilized
  limits:  # You can change the limits to match your actual needs.
    cpu: 1000
  template:
    spec:
      expireAfter: 720h
      nodeClassRef:
        group: karpenter.k8s.aws
        kind: EC2NodeClass
        name: llmaz-demo
      requirements:
      - key: kubernetes.io/arch
        operator: In
        values:
        - amd64
      - key: kubernetes.io/os
        operator: In
        values:
        - linux
      - key: karpenter.sh/capacity-type
        operator: In
        values:
        - spot
      - key: karpenter.k8s.aws/instance-family
        operator: In
        values:                                # replace with your instance-family with gpu supported
        - g4dn
        - g5g
      taints:
      - effect: NoSchedule
        key: nvidia.com/gpu
        value: "true"
```

2. Deploy a model with flavors

```shell
cat <<EOF | kubectl apply -f -
apiVersion: llmaz.io/v1alpha1
kind: OpenModel
metadata:
  name: qwen2-0--5b
spec:
  familyName: qwen2
  source:
    modelHub:
      modelID: Qwen/Qwen2-0.5B-Instruct
  inferenceConfig:
    flavors:
      # The g5g instance family in the aws cloud can provide the t4g GPU type.
      # we define the instance family in the node pool like llmaz-demo-gpu-nodepool.
      - name: t4g
        limits:
          nvidia.com/gpu: 1
        # The flavorName is not recongnized by the Karpenter, so we need to specify the
        # instance-gpu-name via nodeSelector to match the t4g GPU type when node is provisioned
        # by Karpenter from multiple node pools.
        #
        # When you only have a single node pool to provision the GPU instance and the node pool
        # only has one GPU type, it is okay to not specify the nodeSelector. But in practice,
        # it is better to specify the nodeSelector to make the provisioned node more predictable.
        #
        # The available node labels for selecting the target GPU device is listed below:
        # karpenter.k8s.aws/instance-gpu-count
        # karpenter.k8s.aws/instance-gpu-manufacturer
        # karpenter.k8s.aws/instance-gpu-memory
        # karpenter.k8s.aws/instance-gpu-name
        nodeSelector:
          karpenter.k8s.aws/instance-gpu-name: t4g
      # The g4dn instance family in the aws cloud can provide the t4 GPU type.
      # we define the instance family in the node pool like llmaz-demo-gpu-nodepool.
      - name: t4
        limits:
          nvidia.com/gpu: 1
        # The flavorName is not recongnized by the Karpenter, so we need to specify the
        # instance-gpu-name via nodeSelector to match the t4 GPU type when node is provisioned
        # by Karpenter from multiple node pools.
        #
        # When you only have a single node pool to provision the GPU instance and the node pool
        # only has one GPU type, it is okay to not specify the nodeSelector. But in practice,
        # it is better to specify the nodeSelector to make the provisioned node more predictable.
        #
        # The available node labels for selecting the target GPU device is listed below:
        # karpenter.k8s.aws/instance-gpu-count
        # karpenter.k8s.aws/instance-gpu-manufacturer
        # karpenter.k8s.aws/instance-gpu-memory
        # karpenter.k8s.aws/instance-gpu-name
        nodeSelector:
          karpenter.k8s.aws/instance-gpu-name: t4
---
# Currently, the Playground resource type does not support to configure tolerations
# for the generated pods. But luckily, when a pod with the `nvidia.com/gpu` resource  
# is created on the eks cluster, the generated pod will be tweaked with the following
# tolerations:
#   - effect: NoExecute
#      key: node.kubernetes.io/not-ready
#      operator: Exists
#      tolerationSeconds: 300
#   - effect: NoExecute
#     key: node.kubernetes.io/unreachable
#     operator: Exists
#     tolerationSeconds: 300
#   - effect: NoSchedule
#     key: nvidia.com/gpu
#     operator: Exists
apiVersion: inference.llmaz.io/v1alpha1
kind: Playground
metadata:
  labels:
    llmaz.io/model-name: qwen2-0--5b
  name: qwen2-0--5b
spec:
  backendRuntimeConfig:
    backendName: tgi
    # Due to the limitation of our aws account, we have to decrease the resources to match
    # the avaliable instance type which is g4dn.xlarge. If your account has no such limitation,
    # you can remove the custom resources settings below.
    resources:
      limits:
        cpu: "2"
        memory: 4Gi
      requests:
        cpu: "2"
        memory: 4Gi
  modelClaim:
    modelName: qwen2-0--5b
  replicas: 1
EOF
```