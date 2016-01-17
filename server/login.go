package main

import (
	"log"

	pb "github.com/kelseyhightower/grpc-hello-service/hello"

	"golang.org/x/crypto/bcrypt"
	"golang.org/x/net/context"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"

	"github.com/boltdb/bolt"
	"github.com/golang/protobuf/proto"
)

type loginServer struct{}

func (ls *loginServer) Login(ctx context.Context, request *pb.LoginRequest) (*pb.LoginResponse, error) {
	log.Println(request.Username)
	user, err := getUser(boltdb, request.Username)
	if err != nil {
		return nil, err
	}
	err = bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(request.Password))
	if err != nil {
		return nil, grpc.Errorf(codes.PermissionDenied, "")
	}
	return &pb.LoginResponse{"token"}, nil
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
