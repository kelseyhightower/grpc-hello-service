package main

import (
	"crypto/rsa"
	"fmt"
	"io/ioutil"
	"log"

	"github.com/dgrijalva/jwt-go"
	pb "github.com/kelseyhightower/grpc-hello-service/hello"

	"golang.org/x/net/context"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
)

func validateToken(token string, publicKey *rsa.PublicKey) (*jwt.Token, error) {
	jwtToken, err := jwt.Parse(token, func(t *jwt.Token) (interface{}, error) {
		if _, ok := t.Method.(*jwt.SigningMethodRSA); !ok {
			log.Printf("Unexpected signing method: %v", t.Header["alg"])
			return nil, fmt.Errorf("invalid token")
		}
		return publicKey, nil
	})
	if err == nil && jwtToken.Valid {
		return jwtToken, nil
	}
	return nil, err
}

// helloServer is used to implement hello.HelloServer.
type helloServer struct {
	jwtPublicKey *rsa.PublicKey
}

func NewHelloServer(rsaPublicKey string) (*helloServer, error) {
	data, err := ioutil.ReadFile(rsaPublicKey)
	if err != nil {
		return nil, fmt.Errorf("Error reading the jwt public key: %v", err)
	}

	publickey, err := jwt.ParseRSAPublicKeyFromPEM(data)
	if err != nil {
		return nil, fmt.Errorf("Error parsing the jwt public key: %s", err)
	}

	return &helloServer{publickey}, nil
}

func (hs *helloServer) Say(ctx context.Context, request *pb.Request) (*pb.Response, error) {
	var (
		token *jwt.Token
		err   error
	)

	md, ok := metadata.FromContext(ctx)
	if !ok {
		return nil, grpc.Errorf(codes.Unauthenticated, "valid token required.")
	}

	jwtToken, ok := md["authorization"]
	if !ok {
		return nil, grpc.Errorf(codes.Unauthenticated, "valid token required.")
	}

	token, err = validateToken(jwtToken[0], hs.jwtPublicKey)
	if err != nil {
		return nil, grpc.Errorf(codes.Unauthenticated, "valid token required.")
	}

	response := &pb.Response{
		Message: fmt.Sprintf("Hello %s (%s)", request.Name, token.Claims["email"]),
	}

	return response, nil
}
