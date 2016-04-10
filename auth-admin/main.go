// Copyright 2016 Google, Inc. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package main

import (
	"crypto/tls"
	"crypto/x509"
	"database/sql"
	"flag"
	"fmt"
	"io/ioutil"
	"log"

	"github.com/go-sql-driver/mysql"

	"github.com/golang/protobuf/proto"
	"golang.org/x/crypto/bcrypt"
	"golang.org/x/crypto/ssh/terminal"

	pb "github.com/kelseyhightower/grpc-hello-service/auth"
)

func main() {
	var (
		email    = flag.String("e", "", "The user email address.")
		username = flag.String("u", "", "The username.")
		isAdmin  = flag.Bool("a", false, "Enable the admin flag.")
	)
	flag.Parse()

	fmt.Println("enter password:")
	password, err := terminal.ReadPassword(0)
	if err != nil {
		log.Fatal(err)
	}

	passwordHash, err := bcrypt.GenerateFromPassword(password, 12)
	if err != nil {
		log.Fatal(err)
	}

	user := pb.User{
		Email:        *email,
		Username:     *username,
		PasswordHash: string(passwordHash),
		IsAdmin:      *isAdmin,
	}

	data, err := proto.Marshal(&user)
	if err != nil {
		log.Fatal("marshaling error: ", err)
	}

	certPool := x509.NewCertPool()
	pem, err := ioutil.ReadFile("/etc/auth/server-ca.pem")
	if err != nil {
		log.Fatal(err)
	}
	if ok := certPool.AppendCertsFromPEM(pem); !ok {
		log.Fatal("Failed to append PEM.")
	}
	clientCert := make([]tls.Certificate, 0, 1)
	certs, err := tls.LoadX509KeyPair("/etc/auth/client-cert.pem", "/etc/auth/client-key.pem")
	if err != nil {
		log.Fatal(err)
	}
	clientCert = append(clientCert, certs)
	mysql.RegisterTLSConfig("custom", &tls.Config{
		ServerName:   "hightowerlabs:database",
		RootCAs:      certPool,
		Certificates: clientCert,
	})

	db, err := sql.Open("mysql", "auth:grpcauth@tcp(104.196.125.211:3306)/auth?tls=custom")
	if err != nil {
		log.Fatal(err)
	}

	_, err = db.Exec("INSERT INTO users (username, proto) VALUES (?, ?)", user.Username, data)
	if err != nil {
		log.Fatal(err)
	}
}
