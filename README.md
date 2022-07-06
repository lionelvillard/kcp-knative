# Knative in KCP

This guide describes the steps to install Knative in KCP.

## Installing KCP from source

```shell
git clone https://github.com/kcp-dev/kcp.git
cd kcp
git checkout tags/v0.5.0-alpha.1
make build
sudo cp bin/* /usr/local/bin
```

## Start KCP and create a new workspace

You will need three terminals, one for running KCP, one for running `kubectl` 
commands against the KCP cluster and one for running `kubectl`
commands against the physical cluster. 

1. In the KCP terminal, start KCP:
    ```shell
    kcp start
    ```

2. Switch to the KCP cluster terminal
3. Export KUBECONFIG to point to your `kcp` instance:

    ```shell
    export KUBECONFIG=$(pwd)/.kcp/admin.kubeconfig
    ```

Make sure to run this command in the same directory as the one you
used when starting KCP.

3. Create a KCP workspace and immediately enter it:
    
    ```shell
    kubectl kcp workspace create my-workspace --enter
    ```

    ```shell
    Workspace "my-workspace" (type root:Universal) created. Waiting for it to be ready...
    Workspace "my-workspace" (type root:Universal) is ready to use.
    Current workspace is "root:my-workspace".
    ``` 
 
## Registering a Physical Cluster using `syncer`

1. Enable the syncer for a new cluster

    ```shell
    kubectl kcp workload sync localcluster --resources=services,endpoints,pods --syncer-image ghcr.io/kcp-dev/kcp/syncer:fbc1f1a  > syncer.yaml
    ```
2. Switch to the physical cluster (PC) terminal
3. Create a Kubernetes cluster. You can use any Kubernetes cluster. This guide uses `kind`:

    ```shell
    kind create cluster
    ```

4. Register the k8s cluster:

    ```shell
    kubectl apply -f syncer.yaml
    ```

5. Verify the syncer is ready

    ```shell
    kubectl get deployments.apps -n kcpsyncd3465d38e74d3834d4f3694c5bd77d0f015860900047e4ef1a30caf1 
    NAME         READY   UP-TO-DATE   AVAILABLE   AGE
    kcp-syncer   1/1     1            1           97s
    ```

## Installing Knative in KCP

1. Switch to the KCP cluster terminal
2. Install the Knative CRDs in KCP
    
    ```shell
    kubectl apply -f https://github.com/knative/serving/releases/download/knative-v1.5.0/serving-crds.yaml
    ```

3. Verify 
    
    ```shell
    kubectl get crds
    ```
    
    ```
    NAME                                                  CREATED AT
    certificates.networking.internal.knative.dev          2022-07-05T21:26:59Z
    clusterdomainclaims.networking.internal.knative.dev   2022-07-05T21:26:59Z
    configurations.serving.knative.dev                    2022-07-05T21:26:59Z
    domainmappings.serving.knative.dev                    2022-07-05T21:26:59Z
    images.caching.internal.knative.dev                   2022-07-05T21:27:00Z
    ingresses.networking.internal.knative.dev             2022-07-05T21:26:59Z
    metrics.autoscaling.internal.knative.dev              2022-07-05T21:26:59Z
    podautoscalers.autoscaling.internal.knative.dev       2022-07-05T21:27:00Z
    revisions.serving.knative.dev                         2022-07-05T21:27:00Z
    routes.serving.knative.dev                            2022-07-05T21:27:00Z
    serverlessservices.networking.internal.knative.dev    2022-07-05T21:27:00Z
    services.serving.knative.dev                          2022-07-05T21:27:00Z
    ```

4. Install Knative Serving

It is currently not possible to install vanilla Knative Serving in KCP
due to these 3 KCP bugs/missing features:
- https://github.com/kcp-dev/kcp/issues/498
- Namespace DNS resolution (issue to be opened, related to: https://github.com/kcp-dev/kcp/issues/505)
- No support for admission webhooks

This repository contains a Knative Serving configuration compatible with KCP. Apply it:

```shell
kubectl apply -f serving-core.yaml
```

5. Verify all Knative Serving deployments are ready:

```shell
kubectl -n knative-serving get deployments.apps 
```

```shell
NAME                    READY   UP-TO-DATE   AVAILABLE   AGE
activator               1/0     1            1           4m23s
autoscaler              1/1     1            1           11m
controller              1/0     1            1           32m
domain-mapping          1/0     1            1           32m
domainmapping-webhook   1/0     1            1           32m
webhook                 1/0     1            1           32m
```

## TODOs

TODOs:
- networking layer
- Eventing
- HPA
- PodDisruptionBudget
