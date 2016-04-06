# Deploying gRPC Applications with Kubernetes

The following tutorial walks you through deploying the hello service gRPC collection of micro-services.

## Prerequisites

* [Google Cloud Platform Account](https://cloud.google.com)
* [Google Cloud SDK](https://cloud.google.com/sdk)
* [Go 1.6](https://golang.org/dl)

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

Fist install the `cfssl` and `cfssljson` command line tools following the
[cfssl installation guide](https://github.com/cloudflare/cfssl#installation).

Generate the required TLS certs by running the `generate-certs` script from this directory:

```
$ ./generate-certs
```

You should now have the following keys and certs:

```
auth-key.pem
auth.pem
ca-key.pem
ca.pem
client-key.pem
client.pem
hello-key.pem
hello.pem
jwt-key.pem
jwt.pem
```

## Deploying the Auth Service

The auth service is responsible for authenticating users and issuing JWT tokens that can be used to access other gRPC services.
This section will walk you through deploying the auth service using Kubernetes and GCE.

### Create the Auth Data Volume

The auth service requires a persistent disk to store the user database backed by [boltDB](https://github.com/boltdb/bolt).
Create the GCE disk using the gcloud command line tool:

```
$ gcloud compute disks create auth-data
```

### Create the Auth Service Secrets

The auth service requires a set of TLS certificates to serve secure connections between gRPC clients.
In this section we will use the conf2kube utility to convert configuration files to Kubernetes
secrets.

Install the conf2kube utility:

```
$ go get github.com/kelseyhightower/conf2kube
```

#### Create the Auth Service TLS secrets

Create the `auth-tls` Kubernetes secret and store the auth service TLS private key
as `key.pem` using conf2kube:

```
$ conf2kube -n auth-tls -f auth-key.pem -k key.pem | \
  kubectl create -f -
```

Append the Auth service TLS certificate and CA certificate to the `auth-tls` secret:

```
$ kubectl patch secret auth-tls \
  -p `conf2kube -n auth-tls -f auth.pem -k cert.pem`
```

```
$ kubectl patch secret auth-tls \
  -p `conf2kube -n auth-tls -f ca.pem -k ca.pem`
```

Run the `kubectl describe` command to display the details of the `auth-tls` secret:

```
$ kubectl describe secrets auth-tls
```

```
...
Data
====
key.pem:	1679 bytes
ca.pem:		1342 bytes
cert.pem:	1440 bytes
```

#### Create the JWT secrets

The auth service uses a RSA private key for signing JWT tokens.

Create the `jwt-private-key` and `jwt-public-key` secrets using conf2kube:

```
$ conf2kube -n jwt-private-key -f jwt-key.pem -k key.pem | \
  kubectl create -f -
```

```
$ conf2kube -n jwt-public-key -f jwt.pem -k cert.pem | \
  kubectl create -f -
```

#### Auth Secrets Verification

```
$ kubectl get secrets
```
```
NAME                  TYPE                                  DATA      AGE
auth-tls              Opaque                                3         4m
default-token-pg0g8   kubernetes.io/service-account-token   2         3h
jwt-private-key       Opaque                                1         13s
jwt-public-key        Opaque                                1         7s
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
$ kubectl logs auth-xxxx
```

```
2016/01/27 02:18:45 Auth service starting...
2016/01/27 02:18:45 missing auth database, retrying in 5 secs.
2016/01/27 02:18:50 missing auth database, retrying in 5 secs.
2016/01/27 02:18:55 missing auth database, retrying in 5 secs.
```

Notice the auth service is waiting on the auth.db user database file. This file
does not currently exist so we have to create it.

Create the `auth.db` user database. First jump into the container using the
`kubectl exec` command:

```
$ kubectl exec -i -t auth-xxxxx -c auth /bin/ash
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
$ kubectl logs auth-xxxx
```

```
2016/01/27 02:24:30 missing auth database, retrying in 5 secs.
2016/01/27 02:24:35 missing auth database, retrying in 5 secs.
2016/01/27 02:24:40 missing auth database, retrying in 5 secs.
2016/01/27 02:24:45 Auth service started successfully.
```

##  Deploying the Hello Server

The hello service is responsible for returning a hello message to gRPC clients
after validating the JWT token supplied by the client.

Deployment Requirements:

* TLS server certs
* RSA public key for validating JWT tokens
* The `kelseyhightower/hello-server:1.0.0` docker image

### Create Hello Service Secrets

Create the `hello-tls` Kubernetes secret and store the Hello service TLS
private key as key.pem using conf2kube:

```
$ conf2kube -n hello-tls -f hello-key.pem -k key.pem | \
  kubectl create -f -
```

Append the Hello service TLS certificate and CA certificate to the
`hello-tls` secret:

```
$ kubectl patch secret hello-tls \
  -p `conf2kube -n hello-tls -f hello.pem -k cert.pem`
```

```
$ kubectl patch secret hello-tls \
  -p `conf2kube -n hello-tls -f ca.pem -k ca.pem`
```

Run the `kubectl describe` command to display the details of the `hello-tls` secret:

```
$ kubectl describe secrets hello-tls
```

```
...
Data
====
ca.pem:		1342 bytes
cert.pem:	1444 bytes
key.pem:	1679 bytes
```

### Create Hello Server Replication Controller

```
$ kubectl create -f hello-controller.yaml
```

## Expose the Auth and Hello Services

With both the Auth and Hello services deployed we need to expose them
so external gRPC clients can communicate with them. In this tutorial
will will expose both services using a [GCE network loadbalancer](https://cloud.google.com/compute/docs/load-balancing/network).

Expose the auth service:

```
$ kubectl create -f auth-service.yaml
```

Expose the hello service:

```
$ kubectl create -f hello-service.yaml
```

### Edit local DNS

The TLS certificates we created ealier in the tutorial where generated to work
with the following domains:

```
auth.example.com
auth.default.svc.cluster.local
hello.example.com
hello.default.svc.cluster.local
```

Within the Kubernetes cluster the `*.default.svc.cluster.local` domains will resolve
to the service endpoints created in the previous step. The `*.example.com` domains
can be forced to resolve to the external IPs of the auth and hello services by updating
our local `/etc/hosts` file.

Display the external IPs for the auth and hello services:

```
$ kubectl get svc
```
```
NAME         CLUSTER_IP       EXTERNAL_IP       PORT(S)             SELECTOR    AGE
auth         10.159.246.203   130.211.xxx.xxx   7800/TCP,7801/TCP   app=auth    45m
hello        10.159.245.15    104.197.xxx.xxx   7900/TCP,7901/TCP   app=hello   45m
kubernetes   10.159.240.1     <none>            443/TCP             <none>      5h
```

Append the following lines to your local `/etc/hosts` file:

```
130.211.xxx.xxx auth.example.com
104.197.xxx.xxx hello.example.com
```

> Be sure to subsitute '130.211.xxx.xxx' with the acutal external IP from your output.

## Get an auth token

```
$ mkdir -p $HOME/.hello/client/token
```

Install the auth-client:

```
$ go get github.com/kelseyhightower/grpc-hello-service/auth-client
```

```
$ auth-client \
  -ca-cert ca.pem \
  -server-addr auth.example.com:7800 \
  -username kelseyhightower
```

```
wrote /Users/kelseyhightower/.hello/client/.token
```

## Say Hello

Install the hello-client:

```
$ go get github.com/kelseyhightower/grpc-hello-service/hello-client
```

```
$ hello-client \
  -ca-cert ca.pem \
  -server-addr hello.example.com:7900 \
  -tls-cert client.pem \
  -tls-key client-key.pem 
```

```
2016/01/26 19:49:38 Hello Kelsey (kelsey.hightower@gmail.com)
```

## Cleanup

Once you have complete the tutorial run the following commands to clean up the
hello service Kubernetes objects:

Delete the replication controllers:

```
$ kubectl delete rc hello auth
```

Delete the services:

```
$ kubectl delete svc auth hello
```

Delete the secrets:

```
$ kubectl delete secrets auth-tls hello-tls jwt-public-key jwt-private-key
```

Delete the auth service data volume:

After the auth replication controller is deleted you'll be able to remove the `auth-data` volume once the it's no longer in use -- you may have to wait up to 1 minute after the auth controller is deleted. 

```
$ gcloud compute disks delete auth-data
```

Delete the Kubernetes cluster:

If you wanted to avoid being charged for the Kubernetes cluster used in this tutorial run the following command to delete it:

```
$ gcloud container clusters delete hello-tutorial
```
