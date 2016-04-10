// Copyright 2016 Google, Inc. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package main

import (
	"crypto/tls"
	"crypto/x509"
	"database/sql"
	"flag"
	"io/ioutil"
	"log"
	"net"
	"net/http"
	"time"

	pb "github.com/kelseyhightower/grpc-hello-service/auth"
	healthpb "google.golang.org/grpc/health/grpc_health_v1"

	"github.com/go-sql-driver/mysql"
	"golang.org/x/net/trace"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/health"
)

var db *sql.DB

func main() {
	var (
		dbUser          = flag.String("db-user", "auth", "Auth database username.")
		dbPasswd        = flag.String("db-pass", "", "Auth database password.")
		dbHost          = flag.String("db-host", "", "Auth database host.")
		dbPort          = flag.String("db-port", "3306", "Auth database port.")
		dbServerName    = flag.String("db-server-name", "", "Auth database server name.")
		dbServerCACert  = flag.String("db-server-ca-cert", "/etc/auth/db-server-ca.pem", "Database server ca certificate")
		dbClientCert    = flag.String("db-client-cert", "/etc/auth/db-client-cert.pem", "Database client certificate.")
		dbClientKey     = flag.String("db-client-key", "/etc/auth/db-client-key.pem", "Database client key.")
		debugListenAddr = flag.String("debug-listen-addr", "127.0.0.1:7801", "HTTP listen address.")
		listenAddr      = flag.String("listen-addr", "127.0.0.1:7800", "HTTP listen address.")
		tlsCert         = flag.String("tls-cert", "/etc/auth/cert.pem", "TLS server certificate.")
		tlsKey          = flag.String("tls-key", "/etc/auth/key.pem", "TLS server key.")
		jwtPrivateKey   = flag.String("jwt-private-key", "/etc/auth/jwt-key.pem", "The RSA private key to use for signing JWTs")
	)
	flag.Parse()

	var err error
	log.Println("Auth service starting...")

	certPool := x509.NewCertPool()
	pem, err := ioutil.ReadFile(*dbServerCACert)
	if err != nil {
		log.Fatal(err)
	}
	if ok := certPool.AppendCertsFromPEM(pem); !ok {
		log.Fatal("Failed to append PEM.")
	}
	clientCert := make([]tls.Certificate, 0, 1)
	certs, err := tls.LoadX509KeyPair(*dbClientCert, *dbClientKey)
	if err != nil {
		log.Fatal(err)
	}
	clientCert = append(clientCert, certs)
	mysql.RegisterTLSConfig("custom", &tls.Config{
		ServerName:   *dbServerName,
		RootCAs:      certPool,
		Certificates: clientCert,
	})

	dbAddr := net.JoinHostPort(*dbHost, *dbPort)
	dbConfig := mysql.Config{
		User:      *dbUser,
		Passwd:    *dbPasswd,
		Net:       "tcp",
		Addr:      dbAddr,
		DBName:    "auth",
		TLSConfig: "custom",
	}

	for {
		db, err = sql.Open("mysql", dbConfig.FormatDSN())
		if err != nil {
			goto dberror
		}
		err = db.Ping()
		if err != nil {
			goto dberror
		}
		break

	dberror:
		log.Println(err)
		log.Println("error connecting to the auth database, retrying in 5 secs.")
		time.Sleep(5 * time.Second)
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
