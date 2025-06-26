package utils

import (
	"errors"
	"reflect"
)

// Filter apply a function(the first parameter) to each element of a slice(second parameter), and filter out ones which make the function return true.
func Filter(function, slice interface{}) (ret interface{}, err error) {
	defer getErr(&err)
	ret = filter(function, slice)
	return
}

func filter(function, slice interface{}) interface{} {
	in := reflect.ValueOf(slice)
	if in.Kind() != reflect.Slice {
		newErr(errors.New("The first param is not a slice"), "Filter")
	}
	fn := reflect.ValueOf(function)
	inType := in.Type().Elem()
	if !verifyFilterFuncType(fn, inType) {
		newErr(errors.New("Function must be of type func("+inType.String()+") bool"), "Filter")
	}
	var param [1]reflect.Value
	out := reflect.MakeSlice(in.Type(), 0, in.Len())
	for i := 0; i < in.Len(); i++ {
		param[0] = in.Index(i)
		if fn.Call(param[:])[0].Bool() {
			out = reflect.Append(out, in.Index(i))
		}
	}
	return out.Interface()
}

func verifyFilterFuncType(fn reflect.Value, elemType reflect.Type) bool {
	if fn.Kind() != reflect.Func {
		return false
	}
	if fn.Type().NumIn() != 1 || fn.Type().NumOut() != 1 {
		return false
	}
	if fn.Type().In(0) != elemType || fn.Type().Out(0).Kind() != reflect.Bool {
		return false
	}
	return true
}

func CaseElse[T interface{}](when bool, a, b T) T {
	if when {
		return a
	}
	return b
}

func CaseWhen3[K comparable, T interface{}](target K, k1 K, v1 T, k2 K, v2 T, k3 K, v3 T) (r T, ok bool) {
	switch target {
	case k1:
		return v1, true
	case k2:
		return v2, true
	case k3:
		return v3, true
	}
	return
}
