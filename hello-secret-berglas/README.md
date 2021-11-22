# GKE Secret Manager. Consuming secrets using Berglas

NB: This is not an official Google Cloud Product. Use at your own risk

This folder contains a code sample for consuming GSM Secrets from GKE using the [Berglas](https://github.com/GoogleCloudPlatform/berglas).

Berglas is an Open Source CLI for managing Secrets in GSM OR Google Cloud Storage. It uses Cloud KMS to encrypt secrets before storing them. 

In this example we will focus on the GSM integration, we will see how to install the Berglas CLI and Mutating Webhook in GKE, How to configure them and Berglas to create and retrieve a secret from GSM into a pod using Workload Identity and make it available in the environment of the container.

## Prepare the environment

If you already have a GKE cluster with [Workload Identity](https://cloud.google.com/kubernetes-engine/docs/how-to/workload-identity) enabled and configured to pull a secret from GSM, you can skip this step.

Otherwise, start by preparing the environment using these [instructions](./README.md). ONLY follow the [Prepare Env](https://github.com/boredabdel/gke-secret-manager#prepare-environment), [Create Cluster](https://github.com/boredabdel/gke-secret-manager#create-cluster) and [Fetch Credentials](https://github.com/boredabdel/gke-secret-manager#fetch-credentials-for-the-cluster) steps. After you are done come back here.

## Download and setup the Berglas CLI

Export the bucketID as an environment variable, this will used for the remaining of this code sample.

```
export BUCKET_ID=${PROJECT_ID}-berglas
```

Enable needed API's

```
gcloud services enable --project ${PROJECT_ID} \
  cloudkms.googleapis.com \
  storage-api.googleapis.com \
  storage-component.googleapis.com
  secretmanager.googleapis.com \
  cloudfunctions.googleapis.com
```

Bootstrap a Berglas environment, this will create the required KMS and GCS resources.

```
berglas bootstrap --project $PROJECT_ID --bucket $BUCKET_ID
```

(Optional) Enable [Cloud Audit Logging](https://cloud.google.com/logging/docs/audit/configure-data-access#config-api) on the Bucket and KMS

Get the current policy

```
gcloud projects get-iam-policy ${PROJECT_ID} > policy.yaml
```

Add the policy below to the ```policy.yaml``` file. If the auditConfigs exists already, omit it.

```
auditConfigs:
- auditLogConfigs:
  - logType: DATA_READ
  - logType: ADMIN_READ
  - logType: DATA_WRITE
  service: cloudkms.googleapis.com
- auditLogConfigs:
  - logType: ADMIN_READ
  - logType: DATA_READ
  - logType: DATA_WRITE
  service: storage.googleapis.com
```

Apply the new policy

```
gcloud projects set-iam-policy ${PROJECT_ID} policy.yaml
```

Cleanup the policy file

```
rm policy.yaml
```

Clone the repo to deploy the [Mutating Webhook](https://kubernetes.io/docs/reference/access-authn-authz/extensible-admission-controllers/)

```
git clone git@github.com:GoogleCloudPlatform/berglas.git
```

Navigate to the Kubernetes/ folder

```
cd berglas/examples/kubernetes
```

Deploy the Cloud Function needed for the Mutating Webhook. This function is deployed without authentication, it doesn't handle secrets. It only acts as a mechanism to alter the Deployment manifest before it's push to GKE. Below you will see what it means, but briefly. You ask Berglas to retrieve a Secret using a special syntax in the Pod Manifest. This suffix is either ```sm://``` for GSM or ```berglas://``` for GCS. Before the manifest is pushed to the API server, the Cloud function below will be invoked and it will replace the reference to the secret with the proper commands needed by the pod to pull the secret value. You can read more in the Berglas [repo](https://github.com/GoogleCloudPlatform/berglas).

```
gcloud functions deploy berglas-secrets-webhook \
  --project ${PROJECT_ID} \
  --allow-unauthenticated \
  --runtime go113 \
  --entry-point F \
  --trigger-http \
  --region ${GKE_REGION}
```

Ensure the Cloud Function have been deployed

```
gcloud functions list

NAME                     STATUS  TRIGGER       REGION
berglas-secrets-webhook  ACTIVE  HTTP Trigger  europe-west6
```

Extract the Cloud Function endpoint

```
ENDPOINT=$(gcloud functions describe berglas-secrets-webhook \
  --project ${PROJECT_ID} \
  --region ${GKE_REGION} \
  --format 'value(httpsTrigger.url)')
```

Register the webhook URL with the GKE cluster

```
sed "s|REPLACE_WITH_YOUR_URL|$ENDPOINT|" deploy/webhook.yaml | kubectl apply -f -
```

Check that the webhook is running

```
kubectl get mutatingwebhookconfiguration | grep berglas-webhook
```

## Setup the Workload Identity and secrets

Create the GSA needed to configure Workload Identity

```
gcloud iam service-accounts create berglas-gsa --project ${PROJECT_ID}
```

Create a secret using the berglas CLI

```
 berglas create sm://${PROJECT_ID}/berglas-secret berglas-value
```

Grant the GSA access to the secret

```
berglas grant sm://${PROJECT_ID}/berglas-secret \
  --member serviceAccount:berglas-gsa@${PROJECT_ID}.iam.gserviceaccount.com
```

Create a KSA

```
kubectl create sa berglas-ksa
```

Grant the KSA permissions to impresonate the GSA

```
gcloud iam service-accounts add-iam-policy-binding \
  --project ${PROJECT_ID} \
  --role "roles/iam.workloadIdentityUser" \
  --member "serviceAccount:${PROJECT_ID}.svc.id.goog[default/berglas-ksa]" \
  berglas-gsa@${PROJECT_ID}.iam.gserviceaccount.com
```

Annotate the KSA

```
kubectl annotate serviceaccount berglas-ksa \
  iam.gke.io/gcp-service-account=berglas-gsa@${PROJECT_ID}.iam.gserviceaccount.com
```

## Deploy the sample app

Replace the ${PROJECT_ID}, in the ```k8s-manifest.yaml``` file with your project-id.

```
kubectl apply -f k8s-manifest.yaml
```

## Check the secret is available in the environment

WARNING: The steps below will print the content of your secret to the console. Use at your own risk.

Retrieve the pod name

```
kubectl get pods
```

Run the port-forward command to be able to load the webserver in the browser, replace POD_NAME by the value from the previous step.

```
kubectl port-forward pod/POD_NAME 8080:8080 --address 0.0.0.0
```

Open your browser and login to ```http://localhost:8080```

You should see the value of the secret in GSM defined as the env variable ```BERGLAS_SECRET``` 


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