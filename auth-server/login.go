// Copyright 2016 Google, Inc. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package main

import (
	"crypto/rsa"
	"fmt"
	"io/ioutil"
	"time"

	"database/sql"

	pb "github.com/kelseyhightower/grpc-hello-service/auth"

	jwt "github.com/dgrijalva/jwt-go"
	"golang.org/x/crypto/bcrypt"
	"golang.org/x/net/context"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"

	"github.com/golang/protobuf/proto"
)

type authServer struct {
	jwtPrivatekey *rsa.PrivateKey
}

func NewAuthServer(rsaPrivateKey string) (*authServer, error) {
	key, err := ioutil.ReadFile(rsaPrivateKey)
	if err != nil {
		return nil, fmt.Errorf("Error reading the jwt private key: %s", err)
	}
	parsedKey, err := jwt.ParseRSAPrivateKeyFromPEM(key)
	if err != nil {
		return nil, fmt.Errorf("Error parsing the jwt private key: %s", err)
	}
	return &authServer{parsedKey}, nil
}

func (as *authServer) Login(ctx context.Context, request *pb.LoginRequest) (*pb.LoginResponse, error) {
	user, err := getUser(db, request.Username)
	if err != nil {
		return nil, err
	}
	err = bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(request.Password))
	if err != nil {
		return nil, grpc.Errorf(codes.PermissionDenied, "")
	}

	token := jwt.New(jwt.SigningMethodRS256)
	token.Claims["exp"] = time.Now().Add(time.Hour * 72).Unix()
	token.Claims["admin"] = user.IsAdmin
	token.Claims["iss"] = "auth.service"
	token.Claims["iat"] = time.Now().Unix()
	token.Claims["email"] = user.Email
	token.Claims["sub"] = user.Username

	tokenString, err := token.SignedString(as.jwtPrivatekey)
	if err != nil {
		return nil, grpc.Errorf(codes.Internal, err.Error())
	}

	return &pb.LoginResponse{tokenString}, nil
}

func getUser(db *sql.DB, username string) (*pb.User, error) {
	user := &pb.User{}
	var rawUser []byte

	rows, err := db.Query("SELECT proto FROM users WHERE username=?", username)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		if err := rows.Scan(&rawUser); err != nil {
			return nil, err
		}
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	err = proto.Unmarshal(rawUser, user)
	if err != nil {
		return nil, err
	}
	return user, nil
}
