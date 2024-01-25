package main

import (
	"fmt"

	"github.com/sdinsure/agent/pkg/version"
)

func main() {
	fmt.Printf("v%s", version.GetVersion())
}
