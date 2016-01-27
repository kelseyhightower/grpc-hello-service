# Building Docker Images

## Build Statically Linked Linux Binaries

### Auth Service

The Auth service container requires the `auth-admin` and `auth-server`
binaries. Build the binaries and save them to the auth-server directory: 

```
$ GOOS=linux go build -a -o auth-server/auth-admin \
  --ldflags '-extldflags "-static"' \
  -tags netgo -installsuffix netgo \
  ./auth-admin
```

```
$ GOOS=linux go build -a -o auth-server/auth-server \
  --ldflags '-extldflags "-static"' \
  -tags netgo -installsuffix netgo \
  ./auth-server
```

### Hello Service

The Hello service container requires the `hello-server` binary. Build the
binary and save it to the hello-server directory:

```
$ GOOS=linux go build -a -o hello-server/hello-server \
  --ldflags '-extldflags "-static"' \
  -tags netgo -installsuffix netgo \
  ./hello-server
```

## Building Docker Images

### Auth Service

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
