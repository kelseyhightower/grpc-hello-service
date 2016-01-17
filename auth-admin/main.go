package main

import (
	"flag"
	"fmt"
	"log"

	"github.com/boltdb/bolt"
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
		err := b.Put([]byte(user.Username), data)
		return err
	})
	if err != nil {
		log.Fatal(err)
	}
}
