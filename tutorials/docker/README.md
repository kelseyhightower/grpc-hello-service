# Building Docker Images

## Build Statically Linked Linux Binaries

```
$ GOOS=linux ./build
```

```
$ ls -1 bin/
```

```
auth-admin
auth-client
auth-server
hello-client
hello-server
```

## Building Docker Images

### Auth Service

The Auth service container requires the `auth-admin` and `auth-server`
binaries. Copy the binaries to the auth-server directory: 

```
$ cp bin/auth-admin bin/auth-server auth-server
```

Build the `auth-server` Docker image:

```
$ docker build -f auth-server/Dockerfile \
  -t kelseyhightower/auth-server:1.0.0 \
  auth-server/
```

Upload the `auth-server` image to a Docker registry:

```
$ docker push kelseyhightower/auth-server:1.0.0
```

### Hello Service

The Hello service container requires the `hello-server` binary. Copy the
binary to the hello-server directory:

```
$ cp bin/hello-server hello-server/
```

Build the `hello-server` Docker image:

```
$ docker build -f hello-server/Dockerfile \
  -t kelseyhightower/hello-server:1.0.0 \
  hello-server/
```

Upload the `hello-server` image to a Docker registry:

```
$ docker push kelseyhightower/hello-server:1.0.0
```
