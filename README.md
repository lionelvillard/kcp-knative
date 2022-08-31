# Knative in KCP

This guide describes the steps to install Knative in KCP.

## Installing KCP from source

This guide requires KCP with DNS support enabled. 

In a directory:

```shell
git clone https://github.com/kcp-dev/kcp.git
cd kcp
gh pr checkout 1708 # DNS support
make install
```

Then create a Kind cluster (or use your own cluster):

```shell
kind create cluster
```

Build and load the syncer and dns images into Kind:

```shell
KO_DOCKER_REPO=kind.local ko build -B ./cmd/syncer/
KO_DOCKER_REPO=kind.local ko build -B ./cmd/coredns/
```

## Clone this repository

In a directory:

```shell
git clone https://github.com/lionelvillard/kcp-knative.git
cd kcp-knative
```

You will need three terminals:
- one running KCP
- one for running `kubectl` commands against the KCP workspace
- one for running `kubectl` commands against the workload (kind) cluster (previously created)

The current directory in all terminals must be this repository root directory.

## Start KCP and create a new workspace

1. In the KCP terminal, ensure that your ${PATH} contains the output directory of go install, and start kcp on your machine with:
    ```shell
    rm -rf .kcp/ # cleanup 
    kcp start
    ```

2. Switch to the KCP workspace terminal
3. Export KUBECONFIG to point to your KCP instance:

    ```shell
    export KUBECONFIG=$(pwd)/.kcp/admin.kubeconfig
    ```
 
4. By default, KCP starts with only one workspace (no batteries included), the root workspace.  
   Create an organization workspace called `knative` and immediately enter it:
    
    ```shell
    kubectl kcp workspace create knative --enter
    ```

    ```shell
    Workspace "knative" (type root:organization) created. Waiting for it to be ready...
    Workspace "knative" (type root:organization) is ready to use.
    Current workspace is "root:knative".
    ``` 
 
## Registering Kind as a SyncTarget

1. Enable the syncer for the previously created Kind cluster:

    ```shell
    kubectl kcp workload sync kindcluster --resources=poddisruptionbudgets.policy,horizontalpodautoscalers.autoscaling,services,endpoints,pods --syncer-image kind.local/syncer --dns-image kind.local/coredns -o syncer.yaml
    ```
   
2. Switch to the physical cluster (PC) terminal and register the k8s cluster:

    ```shell
    kubectl apply -f syncer.yaml
    ```

3. Verify the syncer is ready (if you don't have [`yq`](https://github.com/mikefarah/yq) just looks for a namespace starting with `kcp-syncer`):

    ```shell
    kubectl get deployments.apps -n $(yq 'select(di == 1).metadata.namespace' syncer.yaml) 
    NAME                              READY   UP-TO-DATE   AVAILABLE   AGE
    kcp-dns-kindcluster-l8dhp7f7      1/1     1            1           28s
    kcp-syncer-kindcluster-l8dhp7f7   1/1     1            1           28s
    ```

## Installing Knative in KCP

1. Switch to the KCP cluster terminal
2. Install the Knative CRDs in the KCP knative workspace
    
    ```shell
    kubectl apply -f https://github.com/knative/serving/releases/download/knative-v1.6.0/serving-crds.yaml
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
   
   You should see only `knative.dev` CRDs.


4. Install Knative Serving Core:
    
    ```shell
    kubectl apply -f https://github.com/knative/serving/releases/download/knative-v1.6.0/serving-core.yaml
    ```

   > Note: ignore the last two errors `no matches for kind "HorizontalPodAutoscaler" in version "autoscaling/v2beta2`

5. Wait a bit (20s-40s or more) and verify all Knative Serving deployments are ready:

   ```shell
   kubectl -n knative-serving get deployments.apps 
   ```
   
   ```shell
   kubectl -n knative-serving get deployments.apps 
   NAME                    READY   UP-TO-DATE   AVAILABLE   AGE
   activator               1/0     1            1           83s
   autoscaler              1/1     1            1           83s
   controller              1/0     1            1           83s
   domain-mapping          1/0     1            1           83s
   domainmapping-webhook   1/0     1            1           83s
   webhook                 1/0     1            1           82s
   ```

6. Install the networking layer. This guide uses [net-kourier](https://github.com/knative-sandbox/net-kourier).

   ```shell
   kubectl apply -f https://github.com/knative/net-kourier/releases/download/knative-v1.6.0/kourier.yaml
   ```
   
7. Since KCP does not support admission controllers yet the config map validating
   webhook needs to be deleted:

   ```shell
   kubectl delete validatingwebhookconfigurations.admissionregistration.k8s.io --all
   kubectl delete mutatingwebhookconfigurations.admissionregistration.k8s.io  --all
   ```

8. Patch the network configmap:
 
   ```shell
   kubectl patch configmap/config-network \
           --namespace knative-serving \
           --type merge \
           --patch '{"data":{"ingress-class":"kourier.ingress.networking.knative.dev"}}'
   ```

## Deploying your first Knative service

In the KCP terminal, deploy the hello world app:

```shell
kn service create hello \
--image gcr.io/knative-samples/helloworld-go \
--port 8080 \
--env TARGET=World
```

This is not working yet (issue with probing)

Deleting the service is currently not possible due to KCP not embedding a garbage collector.

## TODOs

TODOs:
- Eventing
- HPA 
