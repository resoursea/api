package resource

import (
	"reflect"
	"strconv"
)

// IDs to be passed to resources when resceived in the URL
// Ex: resource/123/child/321
// Resource will receive the ID 123 in its arguments,
// ans its child will receive the ID 321 when asked for it
type ID string

type IDMap map[reflect.Type]reflect.Value

var IDType = reflect.TypeOf(ID(""))

var IDPtrType = reflect.TypeOf(reflect.New(IDType))

var EmptyIDValue = reflect.ValueOf(ID(""))

func (ids IDMap) extend(idMap IDMap) {
	for t, v := range idMap {
		ids[t] = v
	}
}

func (id ID) String() string {
	return string(id)
}

func (id ID) Int() (int, error) {
	return strconv.Atoi(id.String())
}
