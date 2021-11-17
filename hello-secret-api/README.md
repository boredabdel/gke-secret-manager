# GKE Secret Manager. Consuming secrets using the GSM API

This folder contains a code sample for consuming GSM Secrets from GKE using the GSM API in Golang.

GSM has client libraries for various programming languages, you can check [them](https://cloud.google.com/secret-manager/docs/reference/libraries#client-libraries-install-go) out

## Prepare the environment

If you already have a GKE cluster with [Workload Identity](https://cloud.google.com/kubernetes-engine/docs/how-to/workload-identity) enabled and configured to pull a secret from GSM, you can skip this step.

Otherwise, start by preparing the environment using these [instructions](./README.md) 


## Build the container image

```
gcloud builds submit -t gcr.io/${PROJECT_ID}/hello-secret-api
```

NB: Replace PROJECT_ID with your own if you haven't followed the environment setup instructions in this repo

## Deploy the application

Replace the ${PROJECT_ID}, ${SECRET_NAME} and ${SECRET_VERSION} in the ```k8s-manifest.yaml``` file with your values.

```
kubectl apply -f k8s-manifest.yaml
```

## Check the output

WARNING: The steps below will print the content of your secret to the console. Use at your own risk.

Use the port-forward command of kubectl to check the pod output

```
kubectl port-forward --address 0.0.0.0 pod/hello-secret-api 8080:8080
```

Curl the output of the server

```
curl localhost:8080
```

Your secret content should be printed in plaintext

## (Optional) Check the Data read Logs

If you have enabled the DATA Read logs on the GSM Service in the environment setup [steps](add link). You can check the logs using the logging page from the console. 

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