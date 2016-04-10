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
	"net"
	"net/http"
	"os"
	"path/filepath"

	pb "github.com/kelseyhightower/grpc-hello-service/hello"
	"golang.org/x/net/trace"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/health"
	healthpb "google.golang.org/grpc/health/grpc_health_v1"
)

func withConfigDir(path string) string {
	return filepath.Join(os.Getenv("HOME"), ".hello", "server", path)
}

func main() {
	var (
		caCert          = flag.String("ca-cert", withConfigDir("ca.pem"), "Trusted CA certificate.")
		debugListenAddr = flag.String("debug-listen-addr", "127.0.0.1:7901", "HTTP listen address.")
		listenAddr      = flag.String("listen-addr", "0.0.0.0:7900", "HTTP listen address.")
		tlsCert         = flag.String("tls-cert", withConfigDir("cert.pem"), "TLS server certificate.")
		tlsKey          = flag.String("tls-key", withConfigDir("key.pem"), "TLS server key.")
		jwtPublicKey    = flag.String("jwt-public-key", withConfigDir("jwt.pem"), "The RSA public key to use for validating JWTs")
	)
	flag.Parse()

	log.Println("Hello service starting...")

	cert, err := tls.LoadX509KeyPair(*tlsCert, *tlsKey)
	if err != nil {
		log.Fatal(err)
	}

	rawCaCert, err := ioutil.ReadFile(*caCert)
	if err != nil {
		log.Fatal(err)
	}
	caCertPool := x509.NewCertPool()
	caCertPool.AppendCertsFromPEM(rawCaCert)

	creds := credentials.NewTLS(&tls.Config{
		Certificates: []tls.Certificate{cert},
		ClientCAs:    caCertPool,
		ClientAuth:   tls.RequireAndVerifyClientCert,
	})

	gs := grpc.NewServer(grpc.Creds(creds))

	hs, err := NewHelloServer(*jwtPublicKey)
	if err != nil {
		log.Fatal(err)
	}

	pb.RegisterHelloServer(gs, hs)

	healthServer := health.NewHealthServer()
	healthServer.SetServingStatus("grpc.health.v1.helloservice", 1)
	healthpb.RegisterHealthServer(gs, healthServer)

	ln, err := net.Listen("tcp", *listenAddr)
	if err != nil {
		log.Fatal(err)
	}
	go gs.Serve(ln)

	trace.AuthRequest = func(req *http.Request) (any, sensitive bool) { return true, true }

	log.Println("Hello service started successfully.")
	log.Fatal(http.ListenAndServe(*debugListenAddr, nil))
}
