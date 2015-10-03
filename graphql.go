package main

import (
	"github.com/sogko/graphql-go-handler"
	"github.com/sogko/todomvc-relay-go/data"
	"log"
	"net/http"
)

func main() {

	// simplest relay-compliant graphql server HTTP handler
	h := gqlhandler.New(&gqlhandler.Config{
		Schema: &data.Schema,
		Pretty: true,
	})

	// create graphql endpoint
	http.Handle("/graphql", h)

	// serve!
	port := ":8080"
	log.Printf(`GraphQL server starting up on http://localhost%v`, port)
	err := http.ListenAndServe(port, nil)
	if err != nil {
		log.Fatalf("ListenAndServe failed, %v", err)
	}
}
