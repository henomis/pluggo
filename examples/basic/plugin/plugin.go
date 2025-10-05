package main

import (
	"context"
	"fmt"
	"os"

	"github.com/henomis/pluggo"
)

type In struct {
	Name string `json:"name" jsonschema:"minLength=3"`
}
type Out struct {
	Greeting string `json:"greeting"`
}

func Hello(ctx context.Context, in *In) (*Out, error) {
	return &Out{Greeting: "hello, " + in.Name + "!"}, nil
}

func main() {
	p := pluggo.NewPlugin()
	v, err := pluggo.NewValidator(&In{})
	if err != nil {
		fmt.Fprintf(os.Stderr, "error creating validator: %v\n", err)
		return
	}

	p.AddFunction("hello", pluggo.NewFunctionHandler(Hello, v).Handler())
	if err := p.Start(); err != nil {
		fmt.Fprintf(os.Stderr, "error starting plugin: %v\n", err)
		return
	}

	p.Stop()
}
