package util

import "fmt"

func ExamplePath_IsLocalhost() {
	fmt.Printf("%v\n", IsLocalhost("example.com"))
	fmt.Printf("%v\n", IsLocalhost("localhost"))
	fmt.Printf("%v\n", IsLocalhost("127.0.0.1"))
	fmt.Printf("%v\n", IsLocalhost("255.255.255.0"))
	fmt.Printf("%v\n", IsLocalhost("::1"))
	fmt.Printf("%v\n", IsLocalhost("::2"))

	// Output:
	// false
	// true
	// true
	// false
	// true
	// false
}
