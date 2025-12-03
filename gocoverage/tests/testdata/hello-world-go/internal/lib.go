package internal

func Foo(x int) int {
	x = DoubleIt(x)
	x = x * x
	x += 10
	return x
}
