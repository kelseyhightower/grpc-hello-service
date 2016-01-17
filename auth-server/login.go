// Copyright 2016 Google, Inc. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package main

import (
	"io/ioutil"
	"time"

	pb "github.com/kelseyhightower/grpc-hello-service/auth"

	jwt "github.com/dgrijalva/jwt-go"
	"golang.org/x/crypto/bcrypt"
	"golang.org/x/net/context"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"

	"github.com/boltdb/bolt"
	"github.com/golang/protobuf/proto"
)

type authServer struct{}

func (as *authServer) Login(ctx context.Context, request *pb.LoginRequest) (*pb.LoginResponse, error) {
	key, err := ioutil.ReadFile(withConfigDir("key.pem"))
	if err != nil {
		return nil, grpc.Errorf(codes.Internal, err.Error())
	}
	parsedKey, err := jwt.ParseRSAPrivateKeyFromPEM(key)
	if err != nil {
		return nil, grpc.Errorf(codes.Internal, err.Error())
	}

	user, err := getUser(boltdb, request.Username)
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

	tokenString, err := token.SignedString(parsedKey)
	if err != nil {
		return nil, grpc.Errorf(codes.Internal, err.Error())
	}

	return &pb.LoginResponse{tokenString}, nil
}

func getUser(db *bolt.DB, username string) (*pb.User, error) {
	user := &pb.User{}
	var rawUser []byte
	err := db.View(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte("users"))
		rawUser = b.Get([]byte(username))
		return nil
	})
	if err != nil {
		return nil, err
	}

	err = proto.Unmarshal(rawUser, user)
	if err != nil {
		return nil, err
	}
	return user, nil
}
