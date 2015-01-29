/*
A high productivity web framework for quickly writing resource based services fully implementing the REST architectural style.

This framework allows you to really focus on the Resources and how it behaves, and let the tool for routing the requests and inject the required dependencies.

For a full guide visit https://github.com/resoursea/api

Example of usage:

package main

import (
	"log"
	"net/http"

	"github.com/resoursea/api"
)

type Gopher struct {
	Message string
}

func (r *Gopher) GET() *Gopher {
	return r
}

func main() {
	router, err := api.NewRouter(Gopher{
		Message: "Hello Gophers!",
	})
	if err != nil {
		log.Fatalln(err)
	}

	// Starting de HTTP server
	log.Println("Starting the service on http://localhost:8080/")
	if err := http.ListenAndServe(":8080", router); err != nil {
		log.Fatalln(err)
	}
}
*/
package main
