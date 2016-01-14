// Copyright 2016 Google, Inc. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package main

import (
	"crypto/tls"
	"crypto/x509"
	"fmt"
	"io/ioutil"
	"log"
	"net"
	"net/http"

	pb "github.com/kelseyhightower/grpc-hello-service/hello"
	healthpb "google.golang.org/grpc/health/grpc_health_v1alpha"

	"golang.org/x/net/context"
	"golang.org/x/net/trace"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/health"
	"google.golang.org/grpc/metadata"
)

// helloServer is used to implement hello.HelloServer.
type helloServer struct{}

func (hs *helloServer) Say(ctx context.Context, request *pb.Request) (*pb.Response, error) {
	m, ok := metadata.FromContext(ctx)
	if ok {
		fmt.Println(m)
	}

	response := &pb.Response{
		Message: fmt.Sprintf("Hello %s", request.Name),
	}

	return response, nil
}

func main() {
	// Load server certs.
	cert, err := tls.LoadX509KeyPair("server-cert.pem", "server-key.pem")
	if err != nil {
		log.Fatal("load peer cert/key error:%v", err)
		return
	}

	// Load the CA certs for client auth.
	caCert, err := ioutil.ReadFile("ca-cert.pem")
	if err != nil {
		log.Fatal("read ca cert file error:%v", err)
		return
	}
	caCertPool := x509.NewCertPool()
	caCertPool.AppendCertsFromPEM(caCert)

	// Create the credentials for the grpc server.
	creds := credentials.NewTLS(&tls.Config{
		Certificates: []tls.Certificate{cert},
		ClientCAs:    caCertPool,
		ClientAuth:   tls.RequireAndVerifyClientCert,
	})

	gs := grpc.NewServer(grpc.Creds(creds))

	pb.RegisterHelloServer(gs, &helloServer{})

	// Register the hello service health check.
	hs := health.NewHealthServer()
	hs.SetServingStatus("grpc.health.v1.hello", 1)
	healthpb.RegisterHealthServer(gs, hs)

	lis, err := net.Listen("tcp", ":8080")
	if err != nil {
		log.Fatal(err)
	}
	trace.AuthRequest = func(req *http.Request) (any, sensitive bool) { return true, true }
	go gs.Serve(lis)

	log.Fatal(http.ListenAndServe(":8888", nil))
}
