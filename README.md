![Image of Yaktocat](http://resoursea.com/img/resoursea-logo-header.png)

## What is Resoursea?

A high productivity web framework for quickly writing resource based services fully implementing the REST architectural style.

This framework allows you to really focus on the Resources and how it behaves, and let the tool for routing the requests and inject the required dependencies.

This framework is written in [Golang](http://golang.org/) and uses the power of its implicit Interface and decentralized package manager.

## Features

- Describes the service API as a Go *struct* structure.
- Method dependencies are constructed and injected when requested.
- Resources becomes accessible simply defining the HTTP methods it is listening to.

## Getting Started

First [install Go](https://golang.org/doc/install) and setting up your [GOPATH](http://golang.org/doc/code.html#GOPATH).

Install the Resoursea package:

~~~
go get github.com/resoursea/api
~~~

To create your service all you have to do is create ordinary Go *structs* and call the `api.newRouter` to route them for you. Then, just call the standard Go server to provide the resources on the network.

## By Example

Save the code below in a file named `main.go`.

~~~ go
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
~~~

Then run your new service:

~~~
go run main.go
~~~

Now you have a new REST service runnig, to **GET** your new `Gopher` Resource, open any browser and type `http://localhost:8080/gopher`.

Another more complete example shows how to build and testing a [simple library service](https://github.com/resoursea/example) with database access, dependency injection and the use of `api.ID`.

## Basis

- Create a hierarchy of ordinary Go *structs* and it will be mapped and routed, each *struct* will turn into a new Resource.

- Define HTTP methods for Resources these methods will be cached and routed.

- Define the dependencies of each method and these dependencies will be constructed and injected whenever necessary.

- You can define the initial state of some Resource and it will be injected in the initializer and constructor methods.

- Resources can define an initializer `Init` method and it will be used to change the initial state of this Resource. It runs just one time when resources are being mapped.

- Resources can define a constructor `New` method, it will be used to construct the Resource every time it needs to be injected. It runs every time one method depends on this resource.

- The URI address of the Resource will be the identifier of the field that receives this Resource.

- The root of the Resource tree isn't attached to any field, so you can pass 2 optional parameters when creating the router: the field identifier and the field tag.

### More Info

* Initial state of Resources are optional, if not defined a new empty instance will be used.

* The initialzer method `Init` is optional, if not declared the initial state will remains the same.

* The constructor method `New` is optional, if not declared the initial state will be injected.

* The first argument of a Go *struct* method is the *struct* itself, it means that for mapped methods the instance of the Resource will be always injected as the first argument.

* One of the constraints for a REST services is to don't keep states in the server component, it means that the Resources shouldn't keep states over the connection. For this rason, every request will receive a new constructed Resource of each dependency.

* Initializers cant have dependencies, and you can return just the resource itself and/or an error.

* Constructors can have dependencies, but **you can't design a circular dependency**, and you can return just the resource itself and/or an error.

* Obs: If you change the state of some dependency somewhere that isn't it's method constructor, when it receives pointer Dependency value for instance, it can cause unexpected behavior.

## The Resource Tree

Resources is declared using ordinary Go *structs* and *slices* of *struts*.

When declaring the service you create a tree of *structs* that will be mapped in routes.

If you declare a list of Resources `type Gophers []Gopher` its behavior will be::

- Requests for the route `/gophers` will be answered by the `Gophers` type.

- Requests for the route `/gophers/:ID` will be answered by the `Gopher` type.

- Methods in `Gopher` could request for `*api.ID`. This dependency keeps the requested ID for this Resource present in the URI.

## The Initializer `Init` Method

This method is used to insert/modify the initial value of some method. If you defined the initial state of this resource on the API creation, this state will always be injected as the first argument of this method. This method just can return the resource itself and/or an error. If this method returns an error, this value will be returned by the `api.NewRouter` method.

## The Constructor `New` Method

This method is used to construct the value of the Resource before it is injected. The initial value of this method will always be injected as the first argument of this method. This method just can return the resource itself and/or an error. If this method returns an error, this value can be caught by any subsequent method.


### The ID Dependency

This dependency is used to identify one Resource in a list. The `api.ID` dependency will be injected in the Resource's methods that it's parent is a slice of the Resource itself.

## A More Complete Example

~~~ go
package main

import (
	"log"
	"net/http"

	"github.com/resoursea/api"
)

type Gopher struct {
	ID          int
	Message     string
	Initialized bool
}

func (r *Gopher) Init() *Gopher {
	r.Initialized = true
	return r
}

func (r *Gopher) New(id api.ID) (*Gopher, error) {
	idInt, err := id.Int()
	if err != nil {
		return nil, err
	}
	r.ID = idInt
	return r, nil
}

func (r *Gopher) GET(err error) (*Gopher, error) {
	return r, err
}

type Gophers []Gopher

type API struct {
	Gophers Gophers
}

func main() {
	router, err := api.NewRouter(API{
		Gophers: Gophers{
			Gopher{
				Message:     "Hello Gophers!",
				Initialized: false,
			},
		},
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

~~~

When you run de service above and try to **GET** one specific `Gopher`, accessing `http://localhost:8080/api/gophers/123` in a browser, the server will return:

~~~ javascript
{
	"Gopher": {
		"ID": 123,
		"Message": "Hello Gophers!",
		"Initialized": true,
	}
}
~~~

Here we can see that the declared initial state was injected in the `Gopher` initializer, and this method updates the initial state. The initial state of `Gopher` and the `api.ID`, sent by the URI, was injected in the constructor method. The `GET` method of `Gopher` just listen for an **HTTP GET** action and return the injected values to the client.


## The Mapped Methods

In the REST arquitecture HTTP methods should be used explicitly in a way that's consistent with the protocol definition. This basic REST design principle establishes a one-to-one mapping between create, read, update, and delete (CRUD) operations and HTTP methods. According to this mapping:


- GET = Retrieve a representation of a Resource.
- POST = Create a new Resource subordinate of the specified resource collection.
- PUT = Update the specified Resource.
- DELETE = Delete the specified Resource.
- HEAD = Get metadata about the specified Resource.

This thing scans and route all Resource's methods that has some of those prefix. Methods also can be used to create the Actions some Resource can perform, you can declare it this way: `POSTLike()`. It will be mapped to the route `[POST] /resource/like`. If you declare just `POST()`, it will be mapped to the route `[POST] /resource`.


## The Dependency Injection

When this framework is creating the Routes for mapped methods, it creates a tree with the dependencies of each method and ensures that there is no circular dependency. This tree is used to answer the request using a depth-first pos-order scanning to construct the dependencies, which ensures that every dependency will be present in the context before it is was requested.

When injecting the required dependency, first the framework search for the initial value of the Resource in the Resource tree, if it wasn't declared, it creates a new empty value for the `struct`. If this dependency has a creator method (New), it is called using this value, and its returned values is injected on the subsequent dependencies until arrive to the root of the dependency tree, the mapped HTTP method itself.

If the method is requesting for an Interface, the framework need find in the Resource tree which one implements it, the framework will search in the siblings, parents or uncles. The same search is done when requiring Structs too, but it is not necessary to be in the Resource tree, if it is not present just a new empty value is used. All this process is done in the route creation time, it guarantee that everything is cached before start to receive the client requests.


## The Resoursea Ecosystem

You also has a high software reuse through the sharing of Resources already created by the community. It’s the resource sea!

Think about a Resource used by virtually all web services, like an instance of the database. Most web services require this Resource to process the request. This Resource and its behavior not need to be implemented by all the developers. It is much better if everyone uses and contribute with just one package that contains this Resource. Thus we’ll have stable and secure packages with Resources to suit most of the needs. Allowing the exclusive focus on particular business rule, of your service.

In Go the explicit declaration of implementation of an Interface is not required, which provide a decoupling of the Interface with the struct which satisfies this interface. This added to the fact tat Go provides a decentralized package manager provides the ideal environment for the sustainable growth of an ecosystem with interfaces and features that can be reused.

Think of a scenario with a list of interfaces, each with a list of Resources that implements it. His work as a developer of services is choosing the interfaces and Resources to attend the requirements of the service and implements only the specific features of your nincho.

### Router Printer

A [printer package](https://github.com/resoursea/printer) was created just for debug reasons. If you want to see the tree of mapped routes and methods, you can import the `https://github.com/resoursea/printer` package and use the `printer.Router` method, passing the *Router* interface returned by a `api.newRouter` call.

## Larn More

[The concept, Samples, Documentation, Interfaces and Resources to use...](http://resoursea.com)

## Join The Community

* [Google Groups](https://groups.google.com/d/forum/resoursea) via [resoursea@googlegroups.com](mailto:resoursea@googlegroups.com)
* [GitHub Issues](https://github.com/resoursea/api/issues)
* [Leave us a comment](https://docs.google.com/forms/d/1GCKn7yN4UYsS4Pv7p2cwHPRfdrURbvB0ajQbaTJrtig/viewform)
* [Twitter](https://twitter.com/resoursea)
