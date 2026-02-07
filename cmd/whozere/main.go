package main

import (
	"fmt"
	"runtime"
)

const version = "0.1.0"

func main() {
	fmt.Printf("whozere v%s\n", version)
	fmt.Printf("Who's here? - Login detection & notification tool\n")
	fmt.Printf("OS: %s, Arch: %s\n", runtime.GOOS, runtime.GOARCH)
}
