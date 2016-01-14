// Copyright 2016 Google, Inc. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package main

import (
	"crypto/tls"
	"crypto/x509"
	"io/ioutil"
	"log"

	healthpb "google.golang.org/grpc/health/grpc_health_v1alpha"
	pb "github.com/kelseyhightower/grpc-hello-service/hello"

	"golang.org/x/net/context"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/credentials/oauth"
)

func main() {
	cert, err := tls.LoadX509KeyPair("client-cert.pem", "client-key.pem")
	if err != nil {
		log.Fatal("load client cert/key error:%v", err)
	}

	caCert, err := ioutil.ReadFile("ca-cert.pem")
	if err != nil {
		log.Fatal("read ca cert file error:%v", err)
	}

	caCertPool := x509.NewCertPool()
	caCertPool.AppendCertsFromPEM(caCert)

	creds := credentials.NewTLS(&tls.Config{
		Certificates: []tls.Certificate{cert},
		RootCAs:      caCertPool,
	})

	conn, err := grpc.Dial("127.0.0.1:8080",
		grpc.WithTransportCredentials(creds),
		grpc.WithPerRPCCredentials(oauth.NewComputeEngine()))
	if err != nil {
		log.Fatal(err)
	}
	defer conn.Close()

	c := pb.NewHelloClient(conn)
	message, err := c.Say(context.Background(), &pb.Request{"Kelsey"})
	if err != nil {
		log.Fatal(err)
	}

	log.Println(message.Message)

	log.Println("Starting health check..")
	hc := healthpb.NewHealthClient(conn)
	status, err := hc.Check(context.Background(),
		&healthpb.HealthCheckRequest{Service: "grpc.health.v1.hello"})
	if err != nil {
		log.Fatal(err)
	}
	log.Println("status:", status)
}
