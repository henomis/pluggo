# 🔌 Pluggo

A simple HTTP-based plugin system for Go.

[![Build Status](https://github.com/henomis/pluggo/actions/workflows/checks.yml/badge.svg)](https://github.com/henomis/pluggo/actions/workflows/checks.yml) [![GoDoc](https://godoc.org/github.com/henomis/pluggo?status.svg)](https://godoc.org/github.com/henomis/pluggo) [![Go Report Card](https://goreportcard.com/badge/github.com/henomis/pluggo)](https://goreportcard.com/report/github.com/henomis/pluggo) [![GitHub release](https://img.shields.io/github/release/henomis/pluggo.svg)](https://github.com/henomis/pluggo/releases)

## 🚀 Features

- 🚀 **Simple Plugin System**: Easy-to-use API for creating and managing plugins
- 🔒 **Type Safety**: Generic-based functions with compile-time type checking
- 🌐 **HTTP Communication**: Reliable plugin communication over HTTP
- 📋 **Schema Validation**: Automatic JSON schema generation and validation
- 🔄 **Dynamic Loading**: Load and execute plugins at runtime
- ⚡ **Performance**: Minimal overhead with efficient execution
- 🏥 **Health Checks**: Built-in health monitoring for plugin processes


## 📦 Installation

```bash
go get github.com/henomis/pluggo
```

## 🎯 Quick Start

### Creating a Plugin

Create a plugin that implements a simple greeting function:

**plugin/plugin.go**
```go
package main

import (
    "context"
    "github.com/henomis/pluggo"
)

type Input struct {
    Name string `json:"name" jsonschema:"minLength=3"`
}

type Output struct {
    Greeting string `json:"greeting"`
}

func Hello(ctx context.Context, in *Input) (*Output, error) {
    return &Output{Greeting: "Hello, " + in.Name + "!"}, nil
}

func main() {
    p := pluggo.NewPlugin()
    
    // Create validator for input validation
    v, err := pluggo.NewValidator(&Input{})
    if err != nil {
        panic(err)
    }
    
    // Register the function
    p.AddFunction("hello", pluggo.NewFunctionHandler(Hello, v).Handler())
    
    // Start the plugin server
    err = p.Start()
    if err != nil {
        panic(err)
    }
}
```

**Build the plugin:**
```bash
go build -o plugin plugin.go
```

### Using the Plugin

**main.go**
```go
package main

import (
    "context"
    "fmt"
    "github.com/henomis/pluggo"
)

type Input struct {
    Name string `json:"name"`
}

type Output struct {
    Greeting string `json:"greeting"`
}

func main() {
    // Create client and load plugin
    client := pluggo.New("./plugin/plugin")
    err := client.Open(context.Background())
    if err != nil {
        panic(err)
    }
    defer client.Close()
    
    // Create type-safe function
    hello := pluggo.NewFunction[Input, Output]("hello", client.Connection())
    
    // Call the function
    result, err := hello.Call(&Input{Name: "World"})
    if err != nil {
        panic(err)
    }
    
    fmt.Println(result.Greeting) // Output: Hello, World!
}
```


## 📚 Examples

The repository includes several examples in the [`examples/`](examples/) directory

## 🏗️ Architecture

Pluggo uses an HTTP-based architecture where:

1. **🚀 Plugin Launch**: Client launches plugin as separate process
2. **📡 HTTP Communication**: Plugin starts HTTP server and communicates port via stdout
3. **🔍 Discovery**: Client discovers available functions via `/_schemas` endpoint
4. **🏥 Health Monitoring**: Built-in health checks via `/_healthz` endpoint
5. **⚡ Function Execution**: Type-safe function calls via HTTP POST requests

```
┌─────────────┐    HTTP     ┌─────────────┐
│   Client    │◄───────────►│   Plugin    │
│             │             │             │
│ ┌─────────┐ │             │ ┌─────────┐ │
│ │Function │ │             │ │Handler  │ │
│ │         │ │────────────►│ │         │ │
│ └─────────┘ │             │ └─────────┘ │
└─────────────┘             └─────────────┘
```

## 🛠️ API Reference

### Plugin Server

#### Creating a Plugin
```go
p := pluggo.NewPlugin()
```

#### Adding Functions
```go
p.AddFunction(name string, handler http.HandlerFunc)
```

#### Starting the Server
```go
err := p.Start()
```

### Client

#### Creating a Client
```go
client := pluggo.New(pluginPath string)
```

#### Opening Connection
```go
err := client.Open(ctx context.Context)
```

#### Getting Available Functions
```go
schemas, err := client.Schemas()
```

### Type-Safe Functions

#### Creating a Function
```go
fn := pluggo.NewFunction[InputType, OutputType](name string, connection *pluggo.Connection)
```

#### Calling a Function
```go
result, err := fn.Call(input *InputType)
```

#### Getting Function Schema
```go
schema, err := fn.Schema()
```

## 🛡️ Input Validation

Pluggo supports automatic input validation using JSON Schema tags:

```go
type Input struct {
    Name  string `json:"name" jsonschema:"minLength=3,maxLength=50"`
    Age   int    `json:"age" jsonschema:"minimum=0,maximum=120"`
    Email string `json:"email" jsonschema:"format=email"`
}
```


## 📄 License

This project is licensed under the MIT License - see the [LICENSE](LICENSE) file for details.

## 🙏 Acknowledgments

- Built with Go's native HTTP server for maximum performance
- Uses JSON Schema for robust input validation
- Inspired by modern microservices architecture patterns

---

**Made with ❤️ by [henomis](https://github.com/henomis)**