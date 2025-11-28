// An example pkg that can be compiled to test goreleaser module.
package main

import "fmt"

func foo(x int) int {
	x *= 2
	x = x * x
	x += 10
	return x
}

func main() {
	fmt.Println("Hello, World!")
}
