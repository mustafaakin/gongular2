package gongular2

import (
	"fmt"
	"net/http"
	"reflect"
	"strings"
)

var defaultErrorHandler = func(err error, c *Context) {
	c.logger.Println("An error has occurred:", err)

	switch err := err.(type) {
	case InjectionError:
		c.MustStatus(http.StatusInternalServerError)
		c.logger.Println("Could not inject the requested data")
	case ValidationError:
		c.MustStatus(http.StatusBadRequest)
		c.SetBody(map[string]interface{}{"ValidationError": err})
	case ParseError:
		c.MustStatus(http.StatusBadRequest)
		c.SetBody(map[string]interface{}{"ParseError": err})
	default:
		c.SetBody(err.Error())
		c.MustStatus(http.StatusInternalServerError)
	}

	c.StopChain()
}

// ErrorHandler is generic interface for error handling
type ErrorHandler func(err error, c *Context)

type InjectionError struct {
	Tip             reflect.Type
	Key             string
	UnderlyingError error
}

func (i InjectionError) Error() string {
	return fmt.Sprintf("Could not inject type %s with key %s because %s", i.Key, i.Tip, i.UnderlyingError.Error())
}

type ValidationError struct {
	Fields map[string]string
	Place  string
}

func (v ValidationError) Error() string {
	s := []string{}
	for k, v := range v.Fields {
		s = append(s, fmt.Sprintf("%s: %s", k, v))
	}
	return fmt.Sprintf("Validation error in %s, %s", v.Place, strings.Join(s, ","))
}

type ParseError struct {
	Place     string
	FieldName string
	Reason    string
}

func (p ParseError) Error() string {
	return fmt.Sprintf("Parse error: %s %s %s", p.Place, p.FieldName, p.Reason)
}
