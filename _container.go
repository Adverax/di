package di

import (
	"fmt"
	"reflect"
)

// Check checks all fields of structure c for nil.
// It used reflection for this.
// If field is nil, then panic.
func Check(s interface{}) {
	v := reflect.ValueOf(s)
	if v.Kind() == reflect.Ptr {
		v = v.Elem()
	}
	for i := 0; i < v.NumField(); i++ {
		f := v.Field(i)
		if f.Type().Kind() == reflect.Pointer && f.IsNil() {
			panic(fmt.Sprintf("field %s is nil", v.Type().Field(i).Name))
		}
	}
}

// Resolve iterate by fields of structure environment and for each field search in the container constructor of component.
// If constructor found, then it calls and return value of constructor writes to the field of structure environment.
func Resolve(container, environment interface{}) {
	e := reflect.ValueOf(environment)
	if e.Kind() != reflect.Ptr {
		panic("environment must be pointer")
	}
	e = e.Elem()
	if e.Kind() != reflect.Struct {
		panic("environment must be struct")
	}
	for i := 0; i < e.NumField(); i++ {
		f := e.Field(i)
		if f.CanSet() {
			v := findValue(container, f.Type())
			if v.IsValid() {
				f.Set(v)
			}
		}
	}
	Check(environment)
}

func findValue(c interface{}, typ reflect.Type) reflect.Value {
	v := reflect.ValueOf(c)
	if v.Kind() == reflect.Ptr {
		v = v.Elem()
	}
	if v.Kind() != reflect.Struct {
		panic("container must be struct")
	}
	for i := 0; i < v.NumField(); i++ {
		f := v.Field(i)
		t := f.Type()
		if t.Kind() == reflect.Func {
			if t.NumIn() == 0 && t.NumOut() == 1 && t.Out(0) == typ {
				return f.Call(nil)[0]
			}
		}
	}
	return reflect.Value{}

}

// IsZeroVal check if any type is its zero value
func isZeroVal(x interface{}) bool {
	return x == nil || reflect.DeepEqual(x, reflect.Zero(reflect.TypeOf(x)).Interface())
}
