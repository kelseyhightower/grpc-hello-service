// Copyright 2016 Google, Inc. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package main

import (
	"crypto/tls"
	"crypto/x509"
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"net"
	"net/http"
	"os"
	"path/filepath"

	pb "github.com/kelseyhightower/grpc-hello-service/hello"
	"google.golang.org/grpc/codes"
	healthpb "google.golang.org/grpc/health/grpc_health_v1alpha"

	"github.com/dgrijalva/jwt-go"
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
	md, ok := metadata.FromContext(ctx)
	if ok {
		jwtToken, ok := md["authorization"]
		if ok {
			token, err := jwt.Parse(jwtToken[0], func(token *jwt.Token) (interface{}, error) {
				if _, ok := token.Method.(*jwt.SigningMethodRSA); !ok {
					return nil, grpc.Errorf(codes.Unauthenticated, "Unexpected signing method: %v", token.Header["alg"])
				}

				data, err := ioutil.ReadFile(withConfigDir("cert.pem"))
				if err != nil {
					return nil, grpc.Errorf(codes.Internal, err.Error())
				}
				publickey, err := jwt.ParseRSAPublicKeyFromPEM(data)
				if err != nil {
					return nil, grpc.Errorf(codes.Internal, err.Error())
				}
				return publickey, nil
			})
			if err == nil && token.Valid {
				log.Println("valid token")
			} else {
				return nil, grpc.Errorf(codes.PermissionDenied, "permission denied %v", err)
			}
		}
	} else {
		return nil, grpc.Errorf(codes.Unauthenticated, "valid token required.")
	}

	response := &pb.Response{
		Message: fmt.Sprintf("Hello %s", request.Name),
	}

	return response, nil
}

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
	)
	flag.Parse()

	cert, err := tls.LoadX509KeyPair(*tlsCert, *tlsKey)
	if err != nil {
		log.Fatal(err)
		return
	}

	rawCaCert, err := ioutil.ReadFile(*caCert)
	if err != nil {
		log.Fatal(err)
		return
	}
	caCertPool := x509.NewCertPool()
	caCertPool.AppendCertsFromPEM(rawCaCert)

	creds := credentials.NewTLS(&tls.Config{
		Certificates: []tls.Certificate{cert},
		ClientCAs:    caCertPool,
		ClientAuth:   tls.RequireAndVerifyClientCert,
	})

	gs := grpc.NewServer(grpc.Creds(creds))
	pb.RegisterHelloServer(gs, &helloServer{})

	hs := health.NewHealthServer()
	hs.SetServingStatus("grpc.health.v1.helloservice", 1)
	healthpb.RegisterHealthServer(gs, hs)

	ln, err := net.Listen("tcp", *listenAddr)
	if err != nil {
		log.Fatal(err)
	}
	go gs.Serve(ln)

	trace.AuthRequest = func(req *http.Request) (any, sensitive bool) { return true, true }
	log.Fatal(http.ListenAndServe(*debugListenAddr, nil))
}
