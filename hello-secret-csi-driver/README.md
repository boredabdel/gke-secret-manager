# GKE Secret Manager. Consuming secrets using the CSI driver

NB: This is not an official Google Cloud Product. Use at your own risk

This folder contains a code sample for consuming GSM Secrets from GKE using the [Secrets Store CSI Driver](https://github.com/kubernetes-sigs/secrets-store-csi-driver).

The Secrets Store CSI Driver is an OSS driver maintained by the Kubernetes community. It allows Kubernetes to mount multiple secrets, keys, and certs stored in enterprise-grade external secrets stores into their pods as a volume. Once the Volume is attached, the data in it is mounted into the container's file system.

The CSI driver itself is Generic and works like any CRI compatible driver for Kubernetes. It's mean to make mounting Secrets into workloads a native step in Kubernetes. Various Cloud Providers build their own plugings against this driver. Google Cloud Has it's own [plugin](https://secrets-store-csi-driver.sigs.k8s.io/getting-started/installation.html).

In this example we will see how to install the driver, the plugin and how to use them to mount a secret from GSM into a pod using Workload Identity

## Prepare the environment

If you already have a GKE cluster with [Workload Identity](https://cloud.google.com/kubernetes-engine/docs/how-to/workload-identity) enabled and configured to pull a secret from GSM, you can skip this step.

Otherwise, start by preparing the environment using these [instructions](./README.md) 


## Installing the driver

```
kubectl apply -f https://raw.githubusercontent.com/kubernetes-sigs/secrets-store-csi-driver/main/deploy/rbac-secretproviderclass.yaml
kubectl apply -f https://raw.githubusercontent.com/kubernetes-sigs/secrets-store-csi-driver/main/deploy/csidriver.yaml
kubectl apply -f https://raw.githubusercontent.com/kubernetes-sigs/secrets-store-csi-driver/main/deploy/secrets-store.csi.x-k8s.io_secretproviderclasses.yaml
kubectl apply -f https://raw.githubusercontent.com/kubernetes-sigs/secrets-store-csi-driver/main/deploy/secrets-store.csi.x-k8s.io_secretproviderclasspodstatuses.yaml
kubectl apply -f https://raw.githubusercontent.com/kubernetes-sigs/secrets-store-csi-driver/main/deploy/secrets-store-csi-driver.yaml
```

Verify that the driver have been installed by querying the pod in the kube-system namespace

```
kubectl get po --namespace=kube-system -l app=csi-secrets-store
```

You should see the Secrets Store CSI Driver pods running on each agent node:

```
NAME                      READY   STATUS    RESTARTS   AGE
csi-secrets-store-4tkb4   3/3     Running   0          40s
csi-secrets-store-ntfd6   3/3     Running   0          40s
csi-secrets-store-rfssl   3/3     Running   0          40s
```


Check that the CRDs have been installed in your cluster

```
kubectl get crd | grep secretprovider
```

You should have 2

```
secretproviderclasses.secrets-store.csi.x-k8s.io            2021-11-17T11:55:49Z
secretproviderclasspodstatuses.secrets-store.csi.x-k8s.io   2021-11-17T11:55:50Z
```

## Installing the GCP plugin

```
kubectl apply -f https://raw.githubusercontent.com/GoogleCloudPlatform/secrets-store-csi-driver-provider-gcp/main/deploy/provider-gcp-plugin.yaml
```

Verify that the plugin have been installed by querying the pod in the kube-system namespace

```
kubectl get po -n kube-system -l app=csi-secrets-store-provider-gcp
```

You should see the Secrets Store CSI provider pods running on each agent node:

```
NAME                                   READY   STATUS    RESTARTS   AGE
csi-secrets-store-provider-gcp-p7j49   1/1     Running   0          38s
csi-secrets-store-provider-gcp-x2852   1/1     Running   0          38s
csi-secrets-store-provider-gcp-xrvs9   1/1     Running   0          38s
```

## Deploy the sample app

The sample app uses the google cloud-sdk image and tries to mount the secret from GSM to ```/var/secrets/good1.txt```

Replace the ${PROJECT_ID}, in the ```k8s-manifest.yaml``` file with your project-id.

```
kubectl apply -f k8s-manifest.yaml
```

## Check the secret mount

WARNING: The steps below will print the content of your secret to the console. Use at your own risk.

```
kubectl exec -it mypod -- cat /var/secrets/good1.txt
```

## (Optional) Check the Data read Logs

If you have enabled the DATA Read logs on the GSM Service in the environment setup [steps](https://github.com/boredabdel/gke-secret-manager#optional-enable-data-access-logs-on-gsm). You can check the logs using the logging page from the console. 

Use the following filters

```
logName="projects/db-pso-project/logs/cloudaudit.googleapis.com%2Fdata_access"
protoPayload.serviceName="secretmanager.googleapis.com"
```

The logs should have a ```principalEmail``` field. This should be the Google Service Account configured with Workload Identity
  
## Cleanup

```
kubectl delete -f k8s-manifest.yaml
```