package resource

import (
	"reflect"
)

// IDs to be passed to resources when resceived in the URL
// Ex: resource/123/child/321
// Resource will receive the ID 123 in its arguments,
// ans its child will receive the ID 321 when asked for it
type ID string

type IDValueMap map[reflect.Type]reflect.Value

func NewIDValueMap() IDValueMap {
	return IDValueMap{}
}

var idType = reflect.TypeOf(ID(""))

var idPtrType = reflect.TypeOf(reflect.New(idType))

var emptyIDValue = reflect.ValueOf(ID(""))
