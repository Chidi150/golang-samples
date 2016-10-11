
Go
Running the Go Shopplace on Container Engine

Create the cluster:

$gcloud container clusters create shopplace --scopes "cloud-platform" --num-nodes 2 --zone us-central1-b --project udumotalag

Get the credentials for the cluster:

$gcloud container clusters get-credentials shopplace --zone us-central1-b --project udumotalag

Verify that you have access to the cluster:

$kubectl cluster-info

This command should indicate that the Kubernetes master is running.

You'll use the kubectl command to create resources in a Container Engine Cluster. 

Cloning the sample application

The sample application is available on GitHub at Chidi150/golang-samples.

Clone the repository: TTTTTTTTTTTTTTTTTTTTT:

$go get -u -d github.com/Chidi150/golang-samples/getting-started/shopplace

Navigate to the sample directory:

$cd $GOPATH/src/github.com/Chidi150/golang-samples/getting-started/shopplace

Creating a Cloud Storage bucket

The Shopplace application uses Google Cloud Storage to store image files.

Enter these commands to create a Cloud Storage bucket. Replace udumotalag with your project ID:

$gsutil mb gs://udumotalag
$gsutil defacl set public-read gs://udumotalag

Note: You can choose any name for your Cloud Storage bucket. To keep the name easy to 
remember, the preceding commands use your project ID as the bucket name. Bucket names 
must be unique across all of Google Cloud Platform, so there is some chance that you 
won't be able to use your project ID as the bucket name. In that case, create a different 
bucket name.

Configuring the application

Copy the Dockerfile to the appropriate directories:

$cp gke_deployment/Dockerfile app/
$cp gke_deployment/Dockerfile pubsub_worker/

Open config.go for editing. This file contains the configuration settings for the sample app.
Uncomment this line:

// DB, err = configureDatastoreDB("<your-project-id>")
Replace <your-project-id> with your project ID:
Uncomment these lines:

// StorageBucketName = "<your-storage-bucket>"
// StorageBucket, err = configureStorage(StorageBucketName)
Replace <your-storage-bucket> with the name of the bucket you created in the previous step.
Uncomment this line:

// PubsubClient, err = configurePubsub("<your-project-id>")
Replace <your-project-id> with your project ID.
Go dependencies

To containerize the Go application, this tutorial uses the aedeploy tool, which assembles 
your app's dependencies. To install aedeploy, run this command:

$go get -u google.golang.org/appengine/cmd/aedeploy

Containerizing the application

The sample application includes a Dockerfile, which is used create the application's Docker 
image. This Docker image used to run the application on Container Engine.

getting-started/shopplace/gke_deployment/Dockerfile VIEW ON GITHUB

Build the application's Docker image:

$cd app/
$aedeploy docker build -t gcr.io/udumotalag/shopplace .
$cd ..

Push the image to Google Container Registry so that your cluster can access the image:

$gcloud docker push gcr.io/udumotalag/shopplace

Containerizing the worker

Containerize the backend worker using the same commands. Build the worker's Docker image:

$cd pubsub_worker/
$aedeploy docker build -t gcr.io/udumotalag/shopplace-worker .
$cd ../gke_deployment/

Push the image to Google Container Registry so that your cluster can access the image:

$gcloud docker push gcr.io/udumotalag/shopplace-worker

Deploying the Shopplace front end

The Shopplace application has a front end that handles web requests and a backend worker 
that processes shops and adds additional information. The cluster resources needed to run 
the front end are defined in shopplace-frontend.yaml. These resources are described as a 
Kubernetes Deployment. Deployments make it easy to create and update a replica set and its 
associated pods.

getting-started/shopplace/gke_deployment/shopplace-frontend.yaml VIEW ON GITHUB

In shopplace-frontend.yaml, replace udumotalag with your project ID.
Use kubectl to deploy the resources to the cluster:

$kubectl create -f shopplace-frontend.yaml

Track the status of the deployment:

$kubectl get deployments

Once the deployment has the same number of available pods as desired pods,
the deployment is complete. If you run into issues with the deployment, 
you can delete it and start over:

$kubectl delete deployments frontend

Once the deployment is complete you can see the pods that the deployment 
created:

$kubectl get pods

Deploying the Shopplace back end

The Shopplace back end is deployed the same way as the front end.

In shopplace-worker.yaml, replace udumotalag with your project ID.
Use kubectl to deploy the resources to the cluster:

$kubectl create -f shopplace-worker.yaml

Verify that the pods are running:

$kubectl get pods

Creating the Shopplace service

Kubernetes Services are used to provide a single point of access to a set of pods. While it's possible to access a single pod, pods are ephemeral and it's usually more convenient to address a set of pods with a single endpoint. In the Shopplace application, The Shopplace service allows you to access the Shopplace frontend pods from a single IP address. This service is defined in shopplace-service.yaml

getting-started/shopplace/gke_deployment/shopplace-service.yaml VIEW ON GITHUB

Notice that the pods and the service that uses the pods are separate. Kubernetes uses labels to select the pods that a service addresses. With labels, you can have a service that addresses pods from different replica sets and have multiple services that point to an individual pod.

Create the Shopplace service:

$kubectl create -f shopplace-service.yaml

Get the service's external IP address with the following:

$kubectl describe service shopplace

Note that it may take up to 60 seconds for the IP address to be allocated.
 The external IP address will be listed under LoadBalancer Ingress.