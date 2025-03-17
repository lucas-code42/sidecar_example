package main

import (
	"encoding/base64"
	"fmt"
	"os"
)

func main() {
	if len(os.Args) < 2 {
		fmt.Println("Usage: ./sidecar <string>")
		os.Exit(1)
	}

	input := os.Args[1]
	encoded := base64.StdEncoding.EncodeToString([]byte(input))
	fmt.Println(encoded)
}
