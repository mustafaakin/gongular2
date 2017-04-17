package gongular2

import "reflect"

// injector remembers the provided values so that you can inject whenever
// you need them
type injector struct {
	values          map[reflect.Type]interface{}
	customProviders map[reflect.Type]CustomProvideFunction
}

// newInjector creates an Injector with its initial structures initialized
func newInjector() *injector {
	return &injector{
		values:          make(map[reflect.Type]interface{}),
		customProviders: make(map[reflect.Type]CustomProvideFunction),
	}
}ß

// Provide registers given value depending on its name
func (inj *injector) Provide(value interface{}) {
	name := reflect.TypeOf(value)
	inj.values[name] = value
}

// ProvideCustom gets the type information from value, however calls CustomProvideFunction
// each time to provide when needed
func (inj *injector) ProvideCustom(value interface{}, fn CustomProvideFunction) {
	name := reflect.TypeOf(value)
	inj.customProviders[name] = fn
}

// CustomProvideFunction is called whenever a value is needed to be provided
// with custom logic
type CustomProvideFunction func(c *Context) (error, interface{})