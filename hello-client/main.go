// Copyright 2016 Google, Inc. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package main

import (
	"crypto/tls"
	"crypto/x509"
	"flag"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"

	"github.com/kelseyhightower/grpc-hello-service/credentials/jwt"
	pb "github.com/kelseyhightower/grpc-hello-service/hello"

	"golang.org/x/net/context"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
)

func withConfigDir(path string) string {
	return filepath.Join(os.Getenv("HOME"), ".hello", "client", path)
}

func main() {
	var (
		caCert     = flag.String("ca-cert", withConfigDir("ca.pem"), "Trusted CA certificate.")
		serverAddr = flag.String("server-addr", "127.0.0.1:7900", "Hello service address.")
		tlsCert    = flag.String("tls-cert", withConfigDir("cert.pem"), "TLS server certificate.")
		tlsKey     = flag.String("tls-key", withConfigDir("key.pem"), "TLS server key.")
		token      = flag.String("token", withConfigDir(".token"), "Path to JWT auth token.")
	)
	flag.Parse()

	cert, err := tls.LoadX509KeyPair(*tlsCert, *tlsKey)
	if err != nil {
		log.Fatal(err)
	}

	rawCACert, err := ioutil.ReadFile(*caCert)
	if err != nil {
		log.Fatal(err)
	}
	caCertPool := x509.NewCertPool()
	caCertPool.AppendCertsFromPEM(rawCACert)

	creds := credentials.NewTLS(&tls.Config{
		Certificates: []tls.Certificate{cert},
		RootCAs:      caCertPool,
	})

	jwtCreds, err := jwt.NewFromTokenFile(*token)
	if err != nil {
		log.Fatal(err)
	}

	conn, err := grpc.Dial(*serverAddr,
		grpc.WithTransportCredentials(creds),
		grpc.WithPerRPCCredentials(jwtCreds))
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
}
