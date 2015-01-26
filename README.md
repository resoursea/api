# Resoursea
A high productivity web framework for quickly writing resource based services fully implementing the REST architectural style.

The base for the REST are the Resources. This framework allows you to really focus on the Resources and how it behaves. And you can let the tool for routing the requests and inject the dependencies when needed.

This framework is written in [Golang](http://golang.org/) and uses the power of its implicit Interface and decentralized package manager.

## Getting Started

First [install Go](https://golang.org/doc/install) and setting up your [GOPATH](http://golang.org/doc/code.html#GOPATH).

Then install the Resoursea package:

~~~
go get github.com/resoursea/api
~~~

Create a new Resource and call the `api` to route the Resource for you. Now you can call the standard Go server to provide the resource on the network. Create a file as shown below and save as `main.go`.

~~~ go
package main

import (
	"log"
	"net/http"

	"github.com/resoursea/api"
)

type Resource struct {
	Message string
}

func (r *Resource) GET() *Resource {
	return r
}

func main() {
	route, err := api.NewRoute(Resource{
		Message: "Hello world!",
	})
	if err != nil {
		log.Fatalln(err)
	}

	// Starting de HTTP server
	log.Println("Starting the service on http://localhost:8080/")
	if err := http.ListenAndServe(":8080", route); err != nil {
		log.Fatalln(err)
	}
}
~~~

Then run your server:

~~~
go run main.go
~~~

You will have a new REST service listening on `http://localhost:8080/`.

To GET your new Resource, open any browser and type `http://localhost:8080/resource`.

## REST and Resources

REST is a set of architectural principles for design web services with a focus on Resources, including how they are addressed and transferred through the HTTP protocol for a wide range of clients written in different languages​​.

However, there were no current tool on the market that allows a level of abstraction that allows you to focus only on the resources, until now. With this tool you focus only on resources and how they are served and the tool takes care of routes your resources and inject the required resources to process the request.

## The Resource Tree

Resources is declared using Go structs and slices of struts.

When declaring the service you create a tree of Resources that will be mapped in routes. The Resource name will be used as its URI. If you declare a list of Resources `type Resources []Resource` and put it in the tree, the service will behave like the imagined: Requests for the route `/resources` will be answer by the `Resources` type, and requests for the route `/resources/:ID` will be answer by the `Resource` struct, and the ID will be cautch and injected as the resource `*api.ID` whenever `Resource` requests for it.

## The Mapped Methods

In the REST arquitecture HTTP methods should be used explicitly in a way that's consistent with the protocol definition. This basic REST design principle establishes a one-to-one mapping between create, read, update, and delete (CRUD) operations and HTTP methods. According to this mapping:


- GET = Retrieve a representation of a Resource.
- POST = Create a new Resource subordinate of the specified resource collection.
- PUT = Update the specified Resource.
- DELETE = Delete the specified Resource.
- HEAD = Get meta-information about the especified Resource.

This API will scan and Route all methods declared that has some of those prefix. Methods also can be used to create the Actions some Resource can perform, you can declare it this way: `POSTLike()`. It will be mapped to the route `[POST] /resource/like`. If you declare just `POST()`, it will be mapped to the route `[POST] /resource`.


## The Dependency Injection

When this framework is creating the Routes for mapped methods, it creates a tree with the dependencies of each method and garants that there is no Circular dependency. This tree is used to answare the request using a depth-first pos-order scanning, witch garants every depenency will be present in the context before it is requisited.

When injecting the required dependency, first the framework search for the initial value of the Resource in the Resource tree, if it wasn't declared, it creates a new empty value for the struct. If this dependency has a creator method (Init), it is called using this value, and its returned values is injected on the subsequent dependencies until arrive to the root of the dependency tree, the mapped method itself.

If the method is requiring for an Interface, the framework need find in the Resource tree witch one implements it, the framework will search in the siblings, parents or uncles. This search is done when requiring Structs, but it is not necessary to be in the resource tree. All this process is done in the route creation time, it guarantee that everything is cached before start to receive the client requests.


## The Resoursea Ecosystem

You also has a high software reuse through the sharing of resources already created by the community. It’s the resource sea!

Think about a resource used by virtually all web services, like an instance of the database. Most web services require this resource to process the request. This resource and its behavior not need to be implemented by all the developers. It is much better if everyone uses and contribute with just one package that contains this resource. Thus we’ll have stable and secure packages with resources to suit most of the needs. Allowing the exclusive focus on particular business rule, of your service.

In Go the explicit declaration of implementation of an Interface is not required, which provide a decoupling of the Interface with the struct which satisfies this interface. This added to the fact tat Go provides a decentralized package manager provides the ideal environment for the sustainable growth of an ecosystem with interfaces and features that can be reused.

Think of a scenario with a list of interfaces, each with a list of resources that implements it. His work as a developer of services is choosing the interfaces and resources to attend the requirements of the service and implements only the specific features of your nincho.

## Larn More

[The concept, Samples, Documentation, interfaces and resources to use...](http://resoursea.com)

## Join The Community

* [Google Groups](https://groups.google.com/d/forum/resoursea) via [resoursea@googlegroups.com](mailto:resoursea@googlegroups.com)
* [GitHub Issues](https://github.com/resoursea/api/issues)
* [Leave us a comment](https://docs.google.com/forms/d/1GCKn7yN4UYsS4Pv7p2cwHPRfdrURbvB0ajQbaTJrtig/viewform)
* [Twitter](https://twitter.com/resoursea)
