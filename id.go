package api

import (
	"reflect"
	"strconv"
)

// IDs to be passed to resources when resceived in the URL
// Ex: resource/123/child/321
// Resource will receive the ID 123 in its arguments,
// ans its child will receive the ID 321 when asked for it
type ID string

type idMap map[reflect.Type]reflect.Value

var idType = reflect.TypeOf(ID(""))

var idPtrType = reflect.TypeOf(reflect.New(idType))

var emptyIDValue = reflect.ValueOf(ID(""))

func (i idMap) extend(ids idMap) {
	for t, v := range ids {
		i[t] = v
	}
}

func (id ID) String() string {
	return string(id)
}

func (id ID) Int() (int, error) {
	return strconv.Atoi(id.String())
}
