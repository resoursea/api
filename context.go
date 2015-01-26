package api

import (
	"log"
	"net/http"
	"reflect"
)

type context struct {
	method *method
	Values []reflect.Value
	IDMap  idMap
	Errors []reflect.Value // To append the errors outputed
}

// Creates a new context
// It creates the initial state used to answer the request
// Since states are not allowed to be stored on te server,
// this initial state is all the service has to answer a request
func newContext(m *method, w http.ResponseWriter, req *http.Request, ids idMap) *context {
	return &context{
		method: m,
		Values: []reflect.Value{
			reflect.ValueOf(w),
			reflect.ValueOf(req),
		},
		IDMap:  ids,
		Errors: []reflect.Value{},
	}
}

func (c *context) run() []reflect.Value {

	//log.Println("Running Context method Method:", c.method.Method.Method.Type)

	// Then run the main method
	inputs := c.getInputs(&c.method.Method)

	return c.method.Method.Func.Call(inputs)
}

// Return the inputs Values from a Method
// For the especial case of the ID input, we should know the requester Type
func (c *context) getInputs(m *reflect.Method) []reflect.Value {

	requester := m.Type.In(0) // Get the requester Type

	values := make([]reflect.Value, m.Type.NumIn())

	//log.Println("Getting inputs:", inputs)
	for i := 0; i < m.Type.NumIn(); i++ {
		t := m.Type.In(i)

		//log.Println("Getting input", t)
		values[i] = c.valueOf(t, requester)
		//log.Println("Getted", values[i], "for", t)

	}

	//log.Println("Returning values:", values, "for", inputs)

	return values
}

// Get the reflect.Value for the required type
func (c *context) valueOf(t reflect.Type, requester reflect.Type) reflect.Value {

	//log.Println("Searching for", t)

	// If it is requesting the first error in the list
	if t == errorType {
		return c.errorValue()
	}

	// If it is requesting the whole error list
	if t == errorSliceType {
		return c.errorSliceValue()
	}

	// If it is requesting the *ID type
	if t == idPtrType {
		return c.idValue(requester)
	}

	// So it can only be a Resource Value
	// Or Request or Writer
	v := c.resourceValue(t)

	// If it is requiring the Elem itself and it returned a Ptr to Elem
	if t.Kind() == reflect.Struct && v.Kind() == reflect.Ptr {
		// It is requiring the Elem of a nil Ptr?
		// Ok, give it an empty Elem of that Type
		if v.IsNil() {
			return reflect.New(t).Elem()
		}

		return v.Elem()
		//log.Println("Transformed", v, "for", t)
	}

	return v
}

// Get the Resource Value of the required Resource Type
// It could be http.ResponseWriter or *http.Request too
func (c *context) resourceValue(t reflect.Type) reflect.Value {
	for _, v := range c.Values {
		switch t.Kind() {
		case reflect.Interface:
			if v.Type().Implements(t) {
				return v
			}
		case reflect.Struct, reflect.Slice: // non-pointer
			if v.Type().Elem() == t {
				return v
			}
		case reflect.Ptr:
			if v.Type() == t {
				return v
			}
		}

	}
	// It is not present yet, so we need to construct it
	return c.initDependencie(t)
}

// Return the first error of the list, or an nil error
func (c *context) errorValue() reflect.Value {
	if len(c.Errors) > 0 {
		return c.Errors[0]
	}
	return errorNilValue
}

// Return a whole error list
func (c *context) errorSliceValue() reflect.Value {
	errs := make([]error, len(c.Errors))
	for i, err := range c.Errors {
		errs[i] = err.Interface().(error)
	}
	return reflect.ValueOf(errs)
}

// Get the reflect.Value for the ID list caught in the URI
// It returns an nil *ID if ID were not passed in the URI
func (c *context) idValue(t reflect.Type) reflect.Value {

	id, exist := c.IDMap[t]
	if exist {
		return id // its an reflect.Value from the type of ID
	}

	// Doesn't exist, returning an empty default ID
	return nilIDValue
}

// Construct all the dependencies level by level
// Garants that every dependencie exists before be requisited
func (c *context) initDependencie(t reflect.Type) reflect.Value {

	dependencie, exist := c.method.Dependencies[t]
	if !exist { // It should never occours
		log.Panicf("Dependencie %s not mapped!!!", t)
	}

	//log.Println("Constructing dependency", dependencie.Value.Type())

	// This Value will be mapped in the index index
	index := len(c.Values)

	// Instanciate a new dependency and add it to the list
	c.Values = append(c.Values, dependencie.init())

	if dependencie.Init != nil {

		inputs := c.getInputs(dependencie.Init) //dependencie.Input, dependencie.Value.Type())

		out := make([]reflect.Value, dependencie.Init.Type.NumOut())

		//log.Printf("Calling %s with %q \n", dependencie.Method.Method.Type, inputs)

		out = dependencie.Init.Func.Call(inputs)

		// If the Init method return something,
		// it will be the resource itself with
		// its values updated
		if dependencie.Init.Type.NumOut() > 0 {

			for i := 0; i < dependencie.Init.Type.NumOut(); i++ {

				out[i].Type()

				//log.Println("### Threating output:", dependencie.Method.Outputs[i])

				if out[i].Type() == errorType {
					//log.Println("### Fucking shit error!!!!", out[i].IsNil(), out[i].IsValid(), out[i].CanSet(), out[i].CanInterface())
					if !out[i].IsNil() {
						c.Errors = append(c.Errors, out[i])
						//log.Println("### Appending the error!!!!")
					}
					continue
				}
				// Check if this output is the dependency itself
				if dependencie.isType(out[i].Type()) {
					//log.Println("### Its just me...", out[i].Type(), out[i].IsValid(), out[i].CanSet(), out[i].CanInterface(), out[i].Type())

					// If this method outputs an Elem insted an Ptr to the Elem
					if out[i].Type().Kind() != reflect.Ptr {
						value := reflect.New(out[i].Type())
						value.Elem().Set(out[i])
						c.Values[index] = value
					} else {
						c.Values[index] = out[i]
					}
				}

			}
		}
	}

	//log.Println("Constructed", c.Values[index], "for", t, "value", c.Values[index].Interface())

	return c.Values[index]
}
