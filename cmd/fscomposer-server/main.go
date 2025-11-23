package main

import (
	"flag"
	"log"

	"github.com/absfs/fscomposer/api"
)

func main() {
	addr := flag.String("addr", ":8080", "HTTP server address")
	flag.Parse()

	server := api.NewServer()
	log.Fatal(server.Start(*addr))
}
