# GKE Secret Manager. Consuming secrets using external Secrets

NB: This is not an official Google Cloud Product. Use at your own risk

This folder contains a code sample for consuming GSM Secrets from GKE using an [External Secrets](https://github.com/external-secrets/kubernetes-external-secrets#how-to-use-it/).

Kubernetes External Secrets is an OSS controller and CRd that allow you to use external secret management systems, like Google Secret Manager or HashiCorp Vault, to securely add secrets in Kubernetes.

## Prepare the environment

If you already have a GKE cluster with [Workload Identity](https://cloud.google.com/kubernetes-engine/docs/how-to/workload-identity) enabled and configured to pull a secret from GSM, you can skip this step.

Otherwise, start by preparing the environment using these [instructions](./README.md). ONLY follow the [Prepare Env](https://github.com/boredabdel/gke-secret-manager#prepare-environment), [Create Cluster](https://github.com/boredabdel/gke-secret-manager#create-cluster) and [Fetch Credentials](https://github.com/boredabdel/gke-secret-manager#fetch-credentials-for-the-cluster) steps. After you are done come back here.

Create a Google Service Account (GSA), this SA will be used by External-Secrets to access GSM 

```
gcloud iam service-accounts create secret-gsa --project ${PROJECT_ID}
```

## Installing external secrets

You will need [helm](https://helm.sh/) to install [External Secrets](https://github.com/external-secrets/kubernetes-external-secrets#how-to-use-it). 

The command below will deploy the CRD and controller. It creates a Kubernetes Service Account (KSA) called ```secret-ksa``` and annotate it with the GSA for Workload Identity to work.
```
$ helm repo add external-secrets https://external-secrets.github.io/kubernetes-external-secrets/
$ helm install external-secrets external-secrets/kubernetes-external-secrets \
    --set serviceAccount.annotations."iam\.gke\.io/gcp-service-account"='secret-gsa@'"${PROJECT_ID}"'.iam.gserviceaccount.com' \
    --set serviceAccount.create=true \
    --set serviceAccount.name="secret-ksa"
```

Verify that controller and CRDs have been installed

```
kubectl get po -l app.kubernetes.io/name=kubernetes-external-secrets
```

You should see one pod

```
NAME                                                            READY   STATUS    RESTARTS   AGE
external-secrets-kubernetes-external-secrets-65c5bf97fd-sb442   1/1     Running   0          16m
```

```
kubectl get crd | grep external
```

You should see 1 crd

```
externalsecrets.kubernetes-client.io                        2021-11-18T10:53:05Z
```

## Configuring External Secrets

Create a Secret in GSM. NB: External Secrets requires secrets to be in JSON format, we will explain why later.

```
echo -n '{"value": "my-secret-value"}' | gcloud secrets create my-secret --replication-policy="automatic" --data-file=-
```

Grant the GSA access to Secrets. NB: We grant access to ALL Secrets in the project as External Secrets doesn't support fine grain access, the controller in the cluster would access secrets in GSM on behalf of all workloads that needs one.

```
gcloud projects add-iam-policy-binding ${PROJECT_ID} \
    --member=serviceAccount:secret-gsa@${PROJECT_ID}.iam.gserviceaccount.com \
    --role=roles/secretmanager.secretAccessor
```

Grant the External Secrets KSA the permission to impresonate the GSA

```
gcloud iam service-accounts add-iam-policy-binding secret-gsa@${PROJECT_ID}.iam.gserviceaccount.com \
    --role roles/iam.workloadIdentityUser \
    --member "serviceAccount:${PROJECT_ID}.svc.id.goog[default/secret-ksa]"
```

## Deploy the ExternalSecret Object

The last step is to deploy an ExternalSecret Object. The yaml file contains comments on what each field means, but here is a more detailed explanation

- metadata.name is the name of the External Secret Object in Kubernetes and the name of the Kubernetes Secret that will be created. This tool syncs from the backend secrets store to Secrets in k8s, we will later update the Secrets in GSM and observe the sync happening.
- spec.projectId is the GCP project ID
- data.key is the name of the Secret in GSM
- data.version is the secre version in GSM
- data.property is the key in the JSON payload in GSM
- data.name is the key in the Generate Kubernetes Secret

Deploy the object

```
kubectl apply -f k8s-manifest.yaml

```

## Check the secret in Kubernetes

WARNING: The steps below will print the content of your secret to the console. Use at your own risk.

```
kubectl get secret my-gcp-secret -o yaml
```

The output should look like

```
apiVersion: v1
data:
  secret-key: BASE64_ENCODED_SECRET_VALUE
kind: Secret
metadata:
  creationTimestamp: "2021-11-19T09:09:10Z"
  name: my-gcp-secret
  namespace: default
  ownerReferences:
  - apiVersion: kubernetes-client.io/v1
    controller: true
    kind: ExternalSecret
    name: my-gcp-secret
    uid: f50ca1d8-d7de-45f7-aa54-89a55f045874
  resourceVersion: "1534999"
  uid: e9fc5bfd-b1c4-4eb6-ad50-7db887c77a44
type: Opaque
```

Copy the value of BASE64_ENCODED_SECRET_VALUE and base64 decode it to read the secret

```
echo BASE64_ENCODED_SECRET_VALUE | base64 -d
```

## Update the Secret in GSM

Let's update the Secret in GSM and check that External Secret have synced it.

```
echo -n '{"value": "new-secret-value"}' | gcloud secrets versions add  my-secret --data-file=- 
```

This should create a new version of the secret in GSM. Because in our ExternalSecrets object with are always syncing the ```latest```. Wait few seconds and repeat the steps in [Check the secret in Kubernetes](). You should see the new value reflected in Kubernetes secrets.

## (Optional) Check the Data read Logs

If you have enabled the DATA Read logs on the GSM Service in the environment setup [steps](https://github.com/boredabdel/gke-secret-manager#optional-enable-data-access-logs-on-gsm). You can check the logs using the logging page from the console. 

Use the following filters

```
logName="projects/db-pso-project/logs/cloudaudit.googleapis.com%2Fdata_access"
protoPayload.serviceName="secretmanager.googleapis.com"
```

The logs should have a ```principalEmail``` field. NB: Regardless of the workload trying to access the secret, the PrincipalEmail will always be the GSA used to configure External Secrets to use Workload Identity.
  
## Cleanup

```
kubectl delete -f k8s-manifest.yaml
```