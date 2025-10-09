package main

import (
	"context"
	"fmt"

	"github.com/henomis/pluggo"
)

type In struct {
	Name string `json:"name"`
}
type Out struct {
	Greeting string `json:"greeting"`
}

func main() {
	client := pluggo.New("./plugin/plugin")
	err := client.Open(context.Background())
	if err != nil {
		fmt.Printf("error opening plugin: %v\n", err)
		return
	}
	defer func() {
		_ = client.Close()
	}()

	schemas, err := client.Schemas()
	if err != nil {
		fmt.Printf("error getting schemas: %v\n", err)
		return
	}
	fmt.Println("Available functions:")
	for functionName := range schemas {
		fmt.Printf("- %s\n", functionName)
	}

	hello, err := pluggo.NewFunction[In, Out]("hello", client.Connection())
	if err != nil {
		fmt.Printf("error creating function: %v\n", err)
		return
	}

	helloSchema, err := hello.Schema()
	if err != nil {
		fmt.Printf("error pinging function: %v\n", err)
		return
	}
	fmt.Printf("Function 'hello' schema: %+v\n", helloSchema)

	fmt.Println("Calling function 'hello'")
	var in In
	in.Name = "world"
	out, err := hello.Call(&in)
	if err != nil {
		fmt.Printf("error calling function: %v\n", err)
		return
	}
	fmt.Println(out.Greeting)

	fmt.Println("Calling function 'error'")
	in.Name = ""
	outErr, err := hello.Call(&in)
	if err != nil {
		fmt.Printf("error calling function: %v\n", err)
		return
	}
	fmt.Println(outErr.Greeting)

}
