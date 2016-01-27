# grpc-hello-service

The repo holds an example gRPC set of applications used to help people learn gRPC.

## Tutorials

* [Kubernetes Deployment](tutorials/kubernetes)
* [Build Docker Images](tutorials/docker)

## Build

Checkout the source code:

```
$ mkdir -p $GOPATH/src/github.com/kelseyhightower/
$ cd $GOPATH/src/github.com/kelseyhightower/
$ git clone https://github.com/kelseyhightower/grpc-hello-service.git
```

Run the build script from the source directory:

```
$ cd grpc-hello-service
$ ./build
```

You should have the following binaries under the bin directory.

```
bin/
├── auth-admin
├── auth-client
├── auth-server
├── hello-client
└── hello-server
``` 
