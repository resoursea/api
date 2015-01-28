# Resoursea

A high productivity web framework for quickly writing resource based services fully implementing the REST architectural style.

This framework allows you to really focus on the Resources and how it behaves, and let the tool for routing the requests and inject the required dependencies.

This framework is written in [Golang](http://golang.org/) and uses the power of its implicit Interface and decentralized package manager.

## Getting Started

First [install Go](https://golang.org/doc/install) and setting up your [GOPATH](http://golang.org/doc/code.html#GOPATH).

Install the Resoursea package:

~~~
go get github.com/resoursea/api
~~~

To create your service all you have to do is create ordinary Go *structs* and call the `api.newRouter` to route them for you. Then, just call the standard Go server to provide the resources on the network.

### By example

Save the code below in a file named `main.go`.

~~~ go
package main

import (
	"log"
	"net/http"

	"github.com/resoursea/api"
)

type HelloWorld struct {
	Message string
}

func (r *HelloWorld) GET() *HelloWorld {
	return r
}

func main() {
	router, err := api.NewRouter(HelloWorld{
		Message: "Hello world!",
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

Now you have a new REST service runnig, to *GET* your new `HelloWorld` Resource, open any browser and type `http://localhost:8080/helloworld`.

Another more complete example shows how to build and testing a [simple library service](https://github.com/resoursea/example) with database access.

## Basis

REST services should be designed as a resource hierarchy.

- Create a hierarchy of ordinary Go *structs* and it will be mapped and routed, each *struct* will turn into a Resource available in your serivce.

- Define HTTP methods for Resource you want to be accessible to the cliente, and these methods will be cached and routed efficiently.

- Define the dependencies of each method and these dependencies will be constructed and injected whenever necessary.

- Resources could define a constructor `Init` method, it will be used to construct the Resource every time it needs to be injected.

- If you define the initial state of some Resource, it will be injected in the constructor method every time it was requested.

- The URI address of the resource will be the identifier of the field that receives this Resource.

- The root of the Resource tree isn't attached to any field, so you can pass 2 optional parameters when creating the router: the field identifier and the field tag.

This way the ordinary Go *structs* are mapped as resources that together forms the service to be offered by the server.

* Initial state of Resources are optional, if not defined a new empty instance will be injected.

* The constructor method `Init` is optional, if not declared the initial state, or a new empty instance will be injected.

* Remember that the first argument of a Go *struct* method is the *struct* itself, it means that for mapped methods (like `GET`, `POST`...) the instance of the resource will be always injected as the first argument.

* One of the constraints for a REST services is to don't keep states in the server component, it means that the Resources shouldn't keep states over the connection. For this rason every request will receive a new initial state of every dependency will be create, constructed and injected.

* Constructors can have dependencies, but *you can't design a circular dependency*. The tool ensures that the dependency will be constructed before the injection itself occurs.


* Obs: Just the constructor should change the state Resource itself and other constructors should not change that state. If you change the state of some dependency that is not the method owner, when it receives pointer Dependency value, it could cause inconsistency.

## The Resource Tree

Resources is declared using ordinary Go *structs* and *slices* of *struts*.

When declaring the service you create a tree of *structs* that will be mapped in routes. The Resource name will be the name of the field that receives the Resource and it is used as its URI address. If you declare a list of Resources `type Resources []Resource` and put it in the Resource tree, the service will behave like the imagined: Requests for the route `/resources` will be answer by the `Resources` type, and requests for the route `/resources/:ID` will be answer by the `Resource` struct, and the ID will be cautch and injected as the dependency `*api.ID` whenever `Resource` requests for it.

### ID

The ID will be injected in every Resource if it's parent is a slice of the Resource itself. For instance:

~~~ go
package main

import (
	"log"
	"net/http"

	"github.com/resoursea/api"
)

type Resource struct {
	ID      int
	Message string
}

func (r *Resource) Init(id api.ID) (*Resource, error) {
	idInt, err := id.Int()
	if err != nil {
		return nil, err
	}
	r.ID = idInt
	return r, nil
}

func (r *Resource) GET(err error) (*Resource, error) {
	return r, err
}

type Resources []Resource

type API struct {
	Resources Resources
}

func main() {
	router, err := api.NewRouter(API{
		Resources: Resources{
			Resource{Message: "Hello world!"},
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

When you run de service above and try to **GET** one specific resource, accessing `http://localhost:8080/api/resources/123` in a browser, the server will return:

~~~ javascript
{
	"Resource": {
		"ID": 123,
		"Message": "Hello world!"
	}
}
~~~

Here we can see that the declared initial state was injected in the Resource constructor, and the `ID` sent by the URI was also injected in the constructor.


## The Mapped Methods

In the REST arquitecture HTTP methods should be used explicitly in a way that's consistent with the protocol definition. This basic REST design principle establishes a one-to-one mapping between create, read, update, and delete (CRUD) operations and HTTP methods. According to this mapping:


- GET = Retrieve a representation of a Resource.
- POST = Create a new Resource subordinate of the specified resource collection.
- PUT = Update the specified Resource.
- DELETE = Delete the specified Resource.
- HEAD = Get metadata about the specified Resource.

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

REST is a set of architectural principles for design web services with a focus on Resources, including how they are addressed and transferred through the HTTP protocol.

With this tool you can focus only on resources and how it behaves,  the tool takes care of routes your resources and inject the required dependencies to process the each request.

[The concept, Samples, Documentation, interfaces and resources to use...](http://resoursea.com)

## Join The Community

* [Google Groups](https://groups.google.com/d/forum/resoursea) via [resoursea@googlegroups.com](mailto:resoursea@googlegroups.com)
* [GitHub Issues](https://github.com/resoursea/api/issues)
* [Leave us a comment](https://docs.google.com/forms/d/1GCKn7yN4UYsS4Pv7p2cwHPRfdrURbvB0ajQbaTJrtig/viewform)
* [Twitter](https://twitter.com/resoursea)
