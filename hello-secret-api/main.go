
// Sample program that uses Secret Manager.
package main

import (
        "context"
        "fmt"
        "net/http"
	"os"
        "log"

        secretmanager "cloud.google.com/go/secretmanager/apiv1"
        secretmanagerpb "google.golang.org/genproto/googleapis/cloud/secretmanager/v1"
)

func main() {

        //Start an HTTP Server
	http.HandleFunc("/", GetSecret)
        log.Printf("Starting Server listening on port 8080")
	http.ListenAndServe(":8080", nil) // set listen port

}

func GetSecret(rw http.ResponseWriter, req *http.Request) {

	// Fetch environment variables
        projectID, ok := os.LookupEnv("PROJECT_ID")
        if !ok {
                log.Fatalf("Environment variable PROJECT_ID is required")
        }
	secretName, ok := os.LookupEnv("SECRET_NAME")
        if !ok {
                log.Fatalf("Environment variable SECRET_NAME is required")
        }
	secretVersion, ok := os.LookupEnv("SECRET_VERSION")
        if !ok {
                log.Fatalf("Environment variable SECRET_VERSION is required")
        }

        // Create the client.
        ctx := context.Background()
        client, err := secretmanager.NewClient(ctx)
        if err != nil {
                log.Fatalf("failed to setup client: %v", err)
        }
        defer client.Close()

        // Build the request.
        accessRequest := &secretmanagerpb.AccessSecretVersionRequest{
                Name: fmt.Sprintf("projects/%s/secrets/%s/versions/%s", projectID,secretName, secretVersion),
        }

        // Call the API.
        result, err := client.AccessSecretVersion(ctx, accessRequest)
        if err != nil {
                log.Fatalf("failed to access secret version: %v", err)
        }


	// Return the secret payload.
	// WARNING: Do not print the secret in a production environment - this
	// snippet is showing how to access the secret material ONLY.
        fmt.Fprint(rw, string(result.Payload.Data))
}