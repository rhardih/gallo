package main

import (
	"fmt"

	"github.com/gorilla/securecookie"
)

func main() {
	fmt.Printf("SESSION_AUTH_KEY=%x\n", securecookie.GenerateRandomKey(16))
	fmt.Printf("SESSION_ENC_KEY=%x\n", securecookie.GenerateRandomKey(16))
}
