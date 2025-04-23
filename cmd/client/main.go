package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"net/http"

	"google.golang.org/protobuf/types/known/emptypb"

	"github.com/kellegous/sonar"
)

func main() {
	var baseURL string
	flag.StringVar(
		&baseURL,
		"base-url",
		"http://localhost:7699",
		"Base URL for the server")
	flag.Parse()

	client := sonar.NewSonarProtobufClient(baseURL, http.DefaultClient)

	res, err := client.GetStoreStats(
		context.Background(),
		&emptypb.Empty{})
	if err != nil {
		log.Panic(err)
	}

	fmt.Printf("%#v\n", res)
}
