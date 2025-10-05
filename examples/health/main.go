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
	client := pluggo.New("./plugin/plugin", pluggo.WithHeartbeatInterval(200*time.Millisecond))
	err := client.Open(context.Background())
	if err != nil {
		fmt.Printf("error opening plugin: %v\n", err)
		return
	}
	defer client.Close()

	check := client.HealthCheck()
	go func() {
		<-check
		fmt.Println("plugin killed, exiting")
	}()

	time.Sleep(3 * time.Second)

	hello := pluggo.NewFunction[In, Out]("hello", client.Connection())

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
