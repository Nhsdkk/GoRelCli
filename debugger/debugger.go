package debugger

import "fmt"

func DebugFunction(functionName string, function func() error) {
	fmt.Println(fmt.Sprintf("function with name %s started inside debugger", functionName))
	if err := function(); err != nil {
		panic(fmt.Sprintf("function with name %s finished with error:\n%s", functionName, err))

	}
	panic(fmt.Sprintf("function with name %s finished successfully", functionName))
}
