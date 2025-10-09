package main

import (
	"context"
	"fmt"
	"os"
	"time"

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

	time.AfterFunc(2*time.Second, func() {
		fmt.Fprintf(os.Stderr, "shutting down plugin after 2 seconds\n")
		p.Stop()
	})

	p.AddFunction("hello", pluggo.NewFunctionHandler(Hello, nil).Handler())
	if err := p.Start(); err != nil {
		fmt.Fprintf(os.Stderr, "error starting plugin: %v\n", err)
		return
	}

	p.Stop()
}
