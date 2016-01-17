package main

import (
	"fmt"
	"log"

	"github.com/boltdb/bolt"
	"github.com/golang/protobuf/proto"
	"golang.org/x/crypto/bcrypt"

	pb "github.com/kelseyhightower/grpc-hello-service/hello"
)

func main() {
	passwordHash, err := bcrypt.GenerateFromPassword([]byte("password"), 12)
	if err != nil {
		log.Fatal(err)
	}

	user := pb.User{
		Username:     "kelsey",
		PasswordHash: string(passwordHash),
		IsAdmin:      true,
	}

	data, err := proto.Marshal(&user)
	if err != nil {
		log.Fatal("marshaling error: ", err)
	}

	db, err := bolt.Open("hello.db", 0600, nil)
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	err = db.Update(func(tx *bolt.Tx) error {
		_, err := tx.CreateBucketIfNotExists([]byte("users"))
		if err != nil {
			return fmt.Errorf("create bucket: %s", err)
		}
		return nil
	})
	if err != nil {
		log.Fatal(err)
	}

	err = db.Update(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte("users"))
		err := b.Put([]byte("kelseyhightower"), data)
		return err
	})
	if err != nil {
		log.Fatal(err)
	}
}
