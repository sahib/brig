package main

import (
	"crypto/rand"
	"fmt"

	"github.com/sahib/brig/cmd/pwd"
	"github.com/sahib/brig/util"
)

func main() {
	pwd, err := pwd.PromptNewPassword(40.0)
	if err != nil {
		fmt.Println("Failed: ", err)
		return
	}

	salt := make([]byte, 32)
	if _, err := rand.Reader.Read(salt); err != nil {
		fmt.Println("Reading salt failed, you're likely doomed.")
		return
	}

	key := util.DeriveKey([]byte(pwd), salt, 32)
	fmt.Printf("Key:  %x\nSalt: %x\n", key, salt)
}
