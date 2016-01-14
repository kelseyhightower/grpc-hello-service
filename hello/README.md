# Hello Protobuf Messages and Services

## Generating the code

```
$ go get -a github.com/golang/protobuf/protoc-gen-go
```

From this directory generate the go source using the protoc tool with the Go plugin:

```
$ protoc --go_out=plugins=grpc:. hello.proto 
```
