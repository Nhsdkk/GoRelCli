package generate

import "fmt"

func Generate(args ...string) error {
	fmt.Println(fmt.Sprintf("generate function called with args %s", args))
	return nil
}
