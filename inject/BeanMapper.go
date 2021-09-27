package inject

import (
	"reflect"
)

type BeanMapper map[reflect.Type]reflect.Value

func (bm BeanMapper) get(t reflect.Type) reflect.Value {
	if v, ok := bm[t]; ok {
		return v
	}
	if t.Kind() == reflect.Ptr {
		t = t.Elem()
	}
	if t.Kind() == reflect.Interface {
		for k, v := range bm {
			if k.Implements(t) {
				return v
			}
		}
	}
	return reflect.Value{}
}

func (bm BeanMapper) set(t reflect.Type, v reflect.Value) {
	if v.IsValid() {
		bm[t] = v
	}
}
