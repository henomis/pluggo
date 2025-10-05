package main

import (
	"context"
	"fmt"
	"os"
	"strings"

	"github.com/henomis/pluggo"
	"github.com/henomis/pluggo/examples/scan/plugins/shared"
)

func exec(ctx context.Context, in *shared.Input) (*shared.Output, error) {
	return &shared.Output{Text: strings.ToUpper(in.Text)}, nil
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
