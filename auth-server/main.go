// Copyright 2016 Google, Inc. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package main

import (
	"flag"
	"log"
	"net"
	"net/http"
	"os"
	"path/filepath"
	"time"

	pb "github.com/kelseyhightower/grpc-hello-service/auth"
	healthpb "google.golang.org/grpc/health/grpc_health_v1alpha"

	"github.com/boltdb/bolt"
	"golang.org/x/net/trace"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/health"
)

func withConfigDir(path string) string {
	return filepath.Join(os.Getenv("HOME"), ".hello", "server", path)
}

var boltdb *bolt.DB

func main() {
	var (
		debugListenAddr = flag.String("debug-listen-addr", "127.0.0.1:7801", "HTTP listen address.")
		listenAddr      = flag.String("listen-addr", "127.0.0.1:7800", "HTTP listen address.")
		tlsCert         = flag.String("tls-cert", withConfigDir("cert.pem"), "TLS server certificate.")
		tlsKey          = flag.String("tls-key", withConfigDir("key.pem"), "TLS server key.")
		jwtPrivateKey   = flag.String("jwt-private-key", withConfigDir("jwt-key.pem"), "The RSA private key to use for signing JWTs")
	)
	flag.Parse()

	var err error
	log.Println("Auth service starting...")
	for {
		_, err := os.Open("/var/lib/auth.db")
		if !os.IsNotExist(err) {
			break
		}
		log.Println("missing auth database, retrying in 5 secs.")
		time.Sleep(5 * time.Second)
	}

	boltdb, err = bolt.Open("/var/lib/auth.db", 0600, nil)
	if err != nil {
		log.Fatal(err)
	}

	ta, err := credentials.NewServerTLSFromFile(*tlsCert, *tlsKey)
	if err != nil {
		log.Fatal(err)
	}

	gs := grpc.NewServer(grpc.Creds(ta))

	as, err := NewAuthServer(*jwtPrivateKey)
	if err != nil {
		log.Fatal(err)
	}
	pb.RegisterAuthServer(gs, as)

	hs := health.NewHealthServer()
	hs.SetServingStatus("grpc.health.v1.authservice", 1)
	healthpb.RegisterHealthServer(gs, hs)

	ln, err := net.Listen("tcp", *listenAddr)
	if err != nil {
		log.Fatal(err)
	}
	go gs.Serve(ln)

	trace.AuthRequest = func(req *http.Request) (any, sensitive bool) { return true, true }
	log.Println("Auth service started successfully.")
	log.Fatal(http.ListenAndServe(*debugListenAddr, nil))
}
