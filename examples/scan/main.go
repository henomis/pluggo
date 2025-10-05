package main

import (
	"context"
	"fmt"
	"path/filepath"

	"github.com/henomis/pluggo"
	"github.com/henomis/pluggo/examples/scan/plugins/shared"
)

type Data struct {
	Path     string
	Plugin   *pluggo.Client
	Function *pluggo.Function[shared.Input, shared.Output]
}

func main() {
	pluginsFiles, err := filepath.Glob("plugins/**/*.so")
	if err != nil {
		fmt.Println("Error listing files:", err)
		return
	}

	var data map[string]Data = make(map[string]Data)

	for _, plugin := range pluginsFiles {
		fmt.Println("Found plugin:", plugin)
		p := pluggo.New(plugin)

		err := p.Open(context.Background())
		if err != nil {
			fmt.Printf("error opening plugin %s: %v\n", plugin, err)
			return
		}

		fn := pluggo.NewFunction[shared.Input, shared.Output]("exec", p.Connection())

		data[plugin] = Data{
			Path:     plugin,
			Plugin:   p,
			Function: fn,
		}
	}

	for _, d := range data {
		in := shared.Input{Text: "Hello, World!"}
		out, err := d.Function.Call(&in)
		if err != nil {
			fmt.Printf("error calling function: %v\n", err)
			continue
		}
		fmt.Printf("Called plugin %s\n", d.Path)
		fmt.Println("Plugin output:", out.Text)
		d.Plugin.Close()
	}
}
