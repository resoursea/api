# Resoursea
A high productivity web framework for quickly writing resource based services.

The easiest way to write REST services highly scalable.  As Fielding wrote in his doctoral dissertation: REST is how the web should work.

With this framework you can really focus on the resources, the base of the REST architecture, and how they are served by your service and leaves the tool to manage them for you, injecting it when needed.

This framework is written in [Golang](http://golang.org/) and uses its powerful decentralized package manager.

## Getting Started

First [install Go](https://golang.org/doc/install) and setting up your [GOPATH](http://golang.org/doc/code.html#GOPATH).

Then install the Resoursea package:

~~~
go get github.com/resoursea/api
~~~

Now you can create your first resource file description. We'll call it `resource.go`.

~~~ go
package main

type Resource struct {
	Message string
}

func (r *Resource) GET() *Resource {
	return r
}
~~~

So you just need to create the service that will provide your resource on the network. Name it `main.go`.

~~~ go
package main

import (
	"log"
	"net/http"

	"github.com/resoursea/api"
)

var route *api.Route

func init() {
	resource, err := api.NewResource(Resource{
		Message: "Hello World!",
	})
	if err != nil {
		log.Fatalln(err)
	}

	route, err = api.NewRoute(resource)
	if err != nil {
		log.Fatalln(err)
	}
}

func main() {
	// Starting de HTTP server
	log.Println("Starting the service on http://localhost:8080/")
	if err := http.ListenAndServe(":8080", route); err != nil {
		log.Fatalln(err)
	}
}
~~~


Then run your server:
~~~
go run server.go
~~~

You will now have a new REST service listening on `http://localhost:8080/`.

To GET your new Resource, open any browser and type `http://localhost:8080/resource`.

## Larn More

[Manual, Samples, Godocs, etc](http://resoursea.com)

## Join The Community

* [Google Groups](https://groups.google.com/d/forum/resoursea) via [resoursea@googlegroups.com](mailto:resoursea@googlegroups.com)
* [GitHub Issues](https://github.com/resoursea/api/issues)
* [Leave us a comment](https://docs.google.com/forms/d/1GCKn7yN4UYsS4Pv7p2cwHPRfdrURbvB0ajQbaTJrtig/viewform)
* [Twitter](https://twitter.com/resoursea)

 © 2014 Resoursea under the Apache License, Version 2.0