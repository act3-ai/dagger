// An example pkg that can be compiled to test goreleaser module.
package main

import (
	"fmt"

	"example.com/hello-world/internal"
)

func main() {
	fmt.Println("Hello, World!", internal.Foo(10))
}
