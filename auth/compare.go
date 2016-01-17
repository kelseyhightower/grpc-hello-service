package main

import (
	"fmt"
	"log"
	"os"

	"golang.org/x/crypto/bcrypt"
)

func main() {
	hashed := []byte("$2a$10$dylCRoEHhvq7Aoci7ZmIGedkz3XgNyA7liZ5yC7TJUGl/zaZHIr1i")
	if err := bcrypt.CompareHashAndPassword(hashed, []byte(os.Args[1])); err != nil {
		log.Fatal(err)
	}
	fmt.Println("valid password")
}
