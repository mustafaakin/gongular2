package gongular2

import (
	"reflect"
)

// injector remembers the provided values so that you can inject whenever
// you need them
type injector struct {
	values          map[reflect.Type]map[string]interface{}
	customProviders map[reflect.Type]map[string]CustomProvideFunction
}

// newInjector creates an Injector with its initial structures initialized
func newInjector() *injector {
	return &injector{
		values:          make(map[reflect.Type]map[string]interface{}),
		customProviders: make(map[reflect.Type]map[string]CustomProvideFunction),
	}
}

// Provide registers given value depending on its name
func (inj *injector) Provide(value interface{}, key string) {
	tip := reflect.TypeOf(value)
	if inj.values[tip] == nil {
		inj.values[tip] = make(map[string]interface{})
	}
	inj.values[tip][key] = value
}

// ProvideCustom gets the type information from value, however calls CustomProvideFunction
// each time to provide when needed
func (inj *injector) ProvideCustom(value interface{}, fn CustomProvideFunction, key string) {
	tip := reflect.TypeOf(value)
	if inj.values[tip] == nil {
		inj.values[tip] = make(map[string]interface{})
	}

	inj.customProviders[tip][key] = fn
}

func (inj *injector) GetDirectValue(tip reflect.Type, key string) (interface{}, bool) {
	val, ok := inj.values[tip][key]
	return val, ok
}

func (inj *injector) GetCustomValue(tip reflect.Type, key string) (CustomProvideFunction, bool) {
	val, ok := inj.customProviders[tip][key]
	return val, ok
}

// CustomProvideFunction is called whenever a value is needed to be provided
// with custom logic
type CustomProvideFunction func(c *Context) (interface{}, error)
