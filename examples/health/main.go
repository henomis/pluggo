package main

import (
	"context"
	"fmt"
	"time"

	"github.com/henomis/pluggo"
)

type In struct {
	Name string `json:"name"`
}
type Out struct {
	Greeting string `json:"greeting"`
}

func main() {
	client := pluggo.New("./plugin/plugin", pluggo.WithHeartbeatInterval(200*time.Millisecond), pluggo.WithHealthCheckTimeout(100*time.Millisecond))
	err := client.Open(context.Background())
	if err != nil {
		fmt.Printf("error opening plugin: %v\n", err)
		return
	}
	defer func() {
		_ = client.Close()
	}()

	check := client.Done()
	go func() {
		<-check
		fmt.Println("plugin killed, exiting")
	}()

	time.Sleep(3 * time.Second)

	hello, err := pluggo.NewFunction[In, Out]("hello", client.Connection())
	if err != nil {
		fmt.Printf("error creating function: %v\n", err)
		return
	}

	fmt.Println("Calling function 'hello'")
	var in In
	in.Name = "world"
	out, err := hello.Call(&in)
	if err != nil {
		fmt.Printf("error calling function: %v\n", err)
		return
	}
	fmt.Println(out.Greeting)
}
