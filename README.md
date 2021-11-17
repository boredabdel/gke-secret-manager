# GKE Secret Manager. Environment setup

This repo contains examples of how to consume secrets from [Google Secret Manager (GSM)](https://cloud.google.com/secret-manager) from [Google Kubernetes Engine (GKE)](https://cloud.google.com/kubernetes-engine)

This main README file contains the steps needed to prepare the environment for the various example. Each sub-folder contains an example, each example will send you to this main page to prepare the GKE cluster, secrets and IAM before you can proceed. Start by check the example you want to follow and follow the instructions from there

## Prepare environment

```
export PROJECT_ID=db-pso-project
export GKE_ZONE=europe-west6-a
```

## Create Cluster

```
gcloud container clusters create gke-secret-manager \
    --project ${PROJECT_ID} \
    --zone ${GKE_ZONE} \
    --release-channel "rapid" \
    --workload-pool "${PROJECT_ID}.svc.id.goog" \
    --scopes=gke-default,cloud-platform
```

## Fetch Credentials for the cluster

```
gcloud container clusters get-credentials gke-secret-manager \
    --project ${PROJECT_ID} \
    --zone ${GKE_ZONE} \
```

## Create a secret

```
echo -n "mypassword" | gcloud secrets create my-db-password \
    --project ${PROJECT_ID} \
    --replication-policy automatic \
    --data-file=-
```

## Verify the secret

```
gcloud secrets versions access 1 --secret my-db-password
```

## Setup Workload Identity

### Create a Google Service Account (GSA)

```
gcloud iam service-accounts create secret-gsa --project ${PROJECT_ID}
```

### Grant the GSA the secretAccessor role on the previously created Secret

```
gcloud secrets add-iam-policy-binding my-db-password \
    --project ${PROJECT_ID} \
    --member="serviceAccount:secret-gsa@${PROJECT_ID}.iam.gserviceaccount.com" \
    --role="roles/secretmanager.secretAccessor"
```

### Create a Kubernetes Service Account (KSA)

```
kubectl create sa --namespace default secret-ksa
```

### Allow the KSA to impersonate the GSA

```
gcloud iam service-accounts add-iam-policy-binding \
    secret-gsa@${PROJECT_ID}.iam.gserviceaccount.com \
    --role roles/iam.workloadIdentityUser \
    --member "serviceAccount:${PROJECT_ID}.svc.id.goog[default/secret-ksa]"
```

### Annotate the KSA

```
kubectl annotate serviceaccount \
    --namespace default secret-ksa  \
    iam.gke.io/gcp-service-account=secret-gsa@${PROJECT_ID}.iam.gserviceaccount.com
```
