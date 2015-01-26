package api

import (
	"reflect"
	"strconv"
)

// IDs to be passed to resources when resceived in the URL
// Ex: resource/123/child/321
// Resource will receive the ID 123 in its arguments,
// ans its child will receive the ID 321 when asked for it

type ID interface {
	String() string
	Int() (int, error)
}

type id struct {
	id string
}

func (i idMap) extend(ids idMap) {
	for t, v := range ids {
		i[t] = v
	}
}

func (i id) String() string {
	return i.id
}

func (i id) Int() (int, error) {
	return strconv.Atoi(i.String())
}

type idMap map[reflect.Type]reflect.Value

var nilIDValue = reflect.ValueOf((*id)(nil))

// TODO, refactor this code
// Dunno another way to do it
var idInterfaceType = reflect.TypeOf(([]ID)(nil)).Elem()
