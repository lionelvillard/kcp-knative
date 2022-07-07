# Knative in KCP

This guide describes the steps to install Knative in KCP.

## Installing KCP from source

In a directory:

```shell
git clone https://github.com/kcp-dev/kcp.git
cd kcp
git checkout tags/v0.5.0-alpha.1
make build
sudo cp bin/* /usr/local/bin
```

## Clone this repository

In a directory:

```shell
git clone https://github.com/lionelvillard/kcp-knative.git
cd kcp-knative
```

You will need three terminals:
- one running KCP
- one for running `kubectl` commands against the KCP cluster
- one for running `kubectl` commands against the physical cluster.

The current directory in all terminals must be this repository root directory.

## Start KCP and create a new workspace

1. In the KCP terminal, start KCP:
    ```shell
    kcp start
    ```

2. Switch to the KCP cluster terminal
3. Export KUBECONFIG to point to your `kcp` instance:

    ```shell
    export KUBECONFIG=$(pwd)/.kcp/admin.kubeconfig
    ```
 
3. Create a KCP workspace and immediately enter it:
    
    ```shell
    kubectl kcp workspace create my-workspace --enter
    ```

    ```shell
    Workspace "my-workspace" (type root:Universal) created. Waiting for it to be ready...
    Workspace "my-workspace" (type root:Universal) is ready to use.
    Current workspace is "root:my-workspace".
    ``` 
 
## Registering a physical cluster using `syncer`

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
   
You should see only `knative.dev` CRDs.

4. Install Knative Serving Core

It is currently not possible to install vanilla Knative Serving in KCP
due to these 3 KCP bugs/missing features:
- https://github.com/kcp-dev/kcp/issues/498
- Namespace DNS resolution (issue to be opened, related to: https://github.com/kcp-dev/kcp/issues/505)
- No support for admission webhooks

This repository contains a Knative Serving configuration compatible with KCP. Apply it:

```shell
kubectl apply -f serving-core.yaml
```

5. Wait a bit (20s-40s or more) and verify all Knative Serving deployments are ready:

```shell
kubectl -n knative-serving get deployments.apps 
```

```shell
NAME                    READY   UP-TO-DATE   AVAILABLE   AGE
activator               1/0     1            1           27s
autoscaler              1/1     1            1           26s
controller              1/0     1            1           26s
domain-mapping          1/0     1            1           26s
domainmapping-webhook   1/0     1            1           26s
webhook                 1/0     1            1           26s
```

6. Install the networking layer. This guide uses net-kourier:

```shell
kubectl apply -f https://github.com/knative/net-kourier/releases/download/knative-v1.5.0/kourier.yaml
```

```shell
kubectl patch configmap/config-network \
        --namespace knative-serving \
        --type merge \
        --patch '{"data":{"ingress-class":"kourier.ingress.networking.knative.dev"}}'
```

Kourier's bootstrap configuration assumes Knative Serving is installed in the `knative-serving`, 
and consequently the envoy readiness probe is failing. You need to update the bootstrap configuration
to point to the actual Knative Serving namespace in the physical cluster.

First you need to find out the name of the namespace in the physical cluster corresponding to `knative-serving`.
In the physical cluster terminal (the one where you created the kind cluster), run this command: 

```shell
kubectl get ns -oyaml
```
Then look for the namespace with the annotation `kcp.dev/namespace-locator: '{"logical-cluster":"root:my-workspace","namespace":"knative-serving"}'`.

Back to the terminal pointing to KCP, run `kubectl edit cm -n kourier-system kourier-bootstrap`, search for
`knative-serving` and replace by the namespace name you found earlier. 

Back to the terminal pointing to the physical cluster, restart the envoy pod:

```shell
kubectl rollout restart deployment -n <the namespace where kourier is installed> 3scale-kourier-gateway 
```

## Deploying your first Knative service

In the KCP terminal, deploy the hello world app:

```shell
kn service create hello \
--image gcr.io/knative-samples/helloworld-go \
--port 8080 \
--env TARGET=World
```

Deleting the service is currently not possible due to KCP not embedding a garbage collector.



## TODOs

TODOs:
- Eventing
- HPA
- PodDisruptionBudget
