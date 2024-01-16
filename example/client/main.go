package main

import (
	"context"
	"flag"
	"fmt"
	"time"

	helloservicehttpclient "github.com/sdinsure/agent/example/api/go-http-client"
	helloserviceclientstub "github.com/sdinsure/agent/example/api/go-openapiv2/client/hello_service"
	"github.com/sdinsure/agent/pkg/version"
)

var serverAddr = flag.String("server_addr", "http://localhost:50091", "server http addr")

func main() {
	flag.Parse()

	version.Print()

	clientService := helloservicehttpclient.MustNewClient(*serverAddr)

	ticker := time.NewTicker(10 * time.Millisecond)
	defer ticker.Stop()

	for _ = range ticker.C {
		greetingMessage := "greeting"

		params := &helloserviceclientstub.HelloServiceSayHelloParams{
			Greeting: &greetingMessage,
			Context:  context.Background(),
		}

		helloServiceSayHelloOk, err := clientService.HelloServiceSayHello(params)
		if err != nil {
			panic(err)
		}
		fmt.Printf("helloServiceSayHelloOk: +%v\n", helloServiceSayHelloOk)
	}
}
