package main

import (
	"crypto/rand"
	"fmt"

	"github.com/disorganizer/brig/repo"
	"github.com/disorganizer/brig/util/security"
)

func main() {
	pwd, err := repo.PromptNewPassword(40.0)
	if err != nil {
		fmt.Println("Failed: ", err)
		return
	}

	salt := make([]byte, 32)
	rand.Reader.Read(salt)

	key := security.Scrypt([]byte(pwd), salt, 32)
	fmt.Printf("Key:  %x\nSalt: %x\n", key, salt)
}
