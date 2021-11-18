# GKE Secret Manager. Consuming secrets using an init-container

This folder contains a code sample for consuming GSM Secrets from GKE using an [init-container](https://kubernetes.io/docs/concepts/workloads/pods/init-containers/).

Kubernetes Init Containers are a specialised containers which run before the app container itself runs. It's typically used to prepare the environment for the app (fetch configurations and make it available to the app). Think about it as a startup-script for a vm.

In the sample below we run pod with an Init Container which uses the ```gcloud``` image to fetch the secret from GSM and writes it to the filesystem of the pod. Kubernetes pods can container multiple containers which share the filesystem. 

## Prepare the environment

If you already have a GKE cluster with [Workload Identity](https://cloud.google.com/kubernetes-engine/docs/how-to/workload-identity) enabled and configured to pull a secret from GSM, you can skip this step.

Otherwise, start by preparing the environment using these [instructions](./README.md) 

## Deploy the sample app

The sample app uses the google cloud-sdk image to fetch the secret from GSM and writes it to ```/var/secrets/good1.txt```.

The Secret Project ($PROJECT_ID), name ($SECRET_NAME) and version ($SECRET_VERSION) are defined using environment variables on the pod manifest. Make sure to change these values before you deploy the app.

You should also edit the path where you want the init-container to write the secret at by editing the $SECRET_PATH variable.

The $SECRET_PATH is defined as a volume of type emptyDIR, this creates an empty volume that all containers in the pod can read/write from/to. In this sample the volume is mounted to both containers. So that the init-container can write the secret to it and the pod could read it later when it's up. 

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