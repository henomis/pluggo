package main

import (
	"context"
	"fmt"
	"os"

	"github.com/henomis/pluggo"
	"github.com/henomis/pluggo/examples/scan/plugins/shared"
)

func exec(ctx context.Context, in *shared.Input) (*shared.Output, error) {
	// Reverse the input string
	runes := []rune(in.Text)
	for i, j := 0, len(runes)-1; i < j; i, j = i+1, j-1 {
		runes[i], runes[j] = runes[j], runes[i]
	}
	return &shared.Output{Text: string(runes)}, nil
}

func main() {
	p := pluggo.NewPlugin()

	p.AddFunction("exec", pluggo.NewFunctionHandler(exec, nil).Handler())
	if err := p.Start(); err != nil {
		fmt.Fprintf(os.Stderr, "error starting plugin: %v\n", err)
		return
	}

	p.Stop()
}
