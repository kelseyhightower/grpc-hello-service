# Hello Service Deployment Tutorial

## This is currently a work in progress and incomplete. ETA 1/28/2016

The following tutorial walks you through deploying the hello service gRPC collection of micro-services.

## Prerequisites

* A working Kubernetes cluster running 1.1.x or greater
* A active Google Cloud Platform account

### Creating a Kubernetes Cluster

The easiest way to get a Kubernetes cluster is to use GKE:

```
$ gcloud container clusters create hello-tutorial
```

At this point you should have a 3 node kubernetes cluster. Run the following command
to configure the kubectl command line tool to use it:

```
$ gcloud container clusters get-credentials hello-tutorial
```

Verify the cluster is healthy:

```
$ kubectl get cs
```

```
NAME                 STATUS    MESSAGE              ERROR
etcd-1               Healthy   {"health": "true"}   nil
controller-manager   Healthy   ok                   nil
scheduler            Healthy   ok                   nil
etcd-0               Healthy   {"health": "true"}   nil
```

## Generating TLS Certs

The microservices in this tutorial are secured by TLS which requires TLS certificates.
In addition to securing our gRPC services a TLS key pair will be used to sign and
validate JWT tokens.

Generate the required TLS certs by running the `generate-certs` script from this directory:

```
$ ./generate-certs
```

## Deploying the Auth Service

The auth service is responsible for authenticating users and issuing JWT tokens that can be used to access other gRPC services.
This section will walk you through deploying the auth service using Kubernetes and GCE.

### Create the Auth Data Volume

The auth service requires a presistent disk to store the user database backed by [boltDB](https://github.com/boltdb/bolt).
Create the GCE disk using the gcloud command line tool:

```
$ gcloud compute disks create auth-data
```

### Create the Auth Service Secrets

The auth service requires a set of TLS certificates to serve secure connections between gRPC clients.

#### Create the Auth Service TLS secrets

Create the `auth-tls` Kubernetes secret and store the auth service TLS private key
as `key.pem` using conf2kube:

```
$ conf2kube -n auth-tls -f auth-key.pem -k key.pem | \
  kubectl create -f -
```

Next, append the auth service TLS certificate and CA certificate to the `auth-tls` secret:

```
$ kubectl patch secret auth-tls \
  -p `conf2kube -n auth-tls -f auth.pem -k cert.pem`
```

```
$ kubectl patch secret auth-tls \
  -p `conf2kube -n auth-tls -f ca.pem -k ca.pem`
```

Run the `kubectl describe` command to display the details of the `auth-server-tls` secret:

```
$ kubectl describe secrets auth-server-tls
```

#### Create the JWT secrets

The auth service uses a RSA private key for signing JWT tokens.

Create the `jwt-private-key` and `jwt-public-key` secrets using conf2kube:

```
$ conf2kube -n jwt-private-key -f jwt-key.pem -k key.pem | \
  kubectl create -f -
```

```
$ conf2kube -n jwt-public-key -f jwt.pem -k jwt.pem | \
  kubectl create -f -
```

### Create the Auth Service Replication Controllers

Replication controllers are used to define the auth service in Kubernetes
and ensure it's running at all times.

Create the auth service replication controller using kubectl:

```
$ kubectl create -f auth-controller.yaml
```

Run the `kubectl get pods` command to monitor the auth service pod:

```
$ kubectl get pods --watch
```

Once the auth server pod is up and running view the logs using the `kubectl logs` command:

```
$ kubectl logs auth-server-xxxx
```

Notice the auth service is waiting on the auth.db user database file. This file
does not currently exist so we have to create it.

Create the `auth.db` user database. First jump into the container using the
`kubectl exec` command:

```
$ kubectl exec -i -t -p auth-server-xxxxx -c auth-server /bin/ash
```

Next, create a new user using the `auth-admin` command:

```
/auth-admin -a -e kelsey.hightower@gmail.com -u kelseyhightower
```

Remember the password you type at the prompt. You'll need it later in the
tutorial.

Exit the container:

```
exit
```

At this point the `auth.db` user database is in place. Run the `kubectl logs`
command again to verify the auth service as started successfully:


```
$ kubectl logs auth-server-xxxx
```

##  Deploying the Hello Server

The hello service is responsible for returning a hello message to gRPC clients
after validating the JWT token supplied by the client.

Deployment Requirements:

* TLS server certs
* RSA public key for validating JWT tokens
* The `kelseyhightower/hello-server:1.0.0` docker image

### Create Hello Server Secrets

```
$ conf2kube -n hello-server-tls -f hello-server-key.pem -k key.pem | \
  kubectl create -f -
```

```
$ kubectl patch secret hello-server-tls \
  -p `conf2kube -n hello-server-tls -f hello-server.pem -k cert.pem`
```

```
$ kubectl patch secret hello-server-tls \
  -p `conf2kube -n hello-server-tls -f ca.pem -k ca.pem`
```

```
$ kubectl patch secret hello-server-tls \
  -p `conf2kube -n hello-server-tls -f auth-server.pem -k jwt.pem`
```

```
$ kubectl describe secrets hello-server-tls
```

### Create Hello Server Replication Controller

```
$ kubectl create -f hello-controller.yaml
```

## Get auth token

```
$ kubectl port-forward auth-server-xxxxx 7801:7801 7800:7800
```

```
auth-client -username kelseyhightower
```

## Say Hello

```
hello-client
```

```
kubectl port-forward hello-server-xxxxx 7901:7901 7900:7900
```

## Create Services

```
$ kubectl create -f hello-service.yaml
```

```
$ kubectl create -f auth-service.yaml
```


## Cleanup

Once you have complete the tutorial run the following commands to clean up the
hello service Kubernetes objects:

Delete the replication controllers:

```
$ kubectl delete rc hello-server auth-server
```

Delete the services:

```
$ kubectl delete svc auth-server hello-server
```

Delete the secrets:

```
$ kubectl delete secrets auth-server-tls hello-server-tls
```

Delete the auth service data volume:

```
$ gcloud compute disks delete auth-data
```
