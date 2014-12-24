package resource

import (
	"reflect"
)

// IDs to be passed to resources when resceived in the URL
// Ex: resource/123/child/321
// Resource will receive the ID 123 in its arguments,
// ans its child will receive the ID 321 when asked for it
type ID string

type IDMap map[reflect.Type]reflect.Value

var IdType = reflect.TypeOf(ID(""))

var IdPtrType = reflect.TypeOf(reflect.New(IdType))

var EmptyIDValue = reflect.ValueOf(ID(""))

func (ids IDMap) add(child IDMap) {
	for t, v := range child {
		ids[t] = v
	}
}
