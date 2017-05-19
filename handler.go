package gongular2

import (
	"errors"
	"reflect"

	"net/http"

	"fmt"

	"github.com/gorilla/websocket"
)

// RequestHandler is a generic handler for gongular2
type RequestHandler interface {
	Handle(c *Context) error
}

type middleRequestHandler func(c *Context) error

type handlerContext struct {
	method    string
	name      string
	websocket bool
	// The analyzed reflection data so that we can cache it
	param     bool
	query     bool
	body      bool
	form      bool
	injection bool

	// HandlerType
	tip reflect.Type

	// The actual function
	RequestHandler middleRequestHandler
}

func transformRequestHandler(path string, method string, injector *injector, handler RequestHandler) (*handlerContext, error) {
	rhc := handlerContext{}
	// Handler parse parameters
	handlerElem := reflect.TypeOf(handler).Elem()
	rhc.name = fmt.Sprintf("%s.%s", handlerElem.PkgPath(), handlerElem.Name())

	rhc.tip = handlerElem
	rhc.method = method

	// See if we have any params
	param, paramOk := handlerElem.FieldByName("Param")
	if paramOk {
		// If we have something param, it should be a struct only
		// TODO: Additional check, it should be flat struct
		// TODO: Additional check, it should be compatible with path
		if param.Type.Kind() != reflect.Struct {
			return nil, errors.New("Param field added but it is not a struct")
		}
	}
	rhc.param = paramOk

	query, queryOk := handlerElem.FieldByName("Query")
	if queryOk {
		// If we have something param, it should be a struct only
		// TODO: Additional check, it should be flat struct
		if query.Type.Kind() != reflect.Struct {
			return nil, errors.New("Query field added but it is not a struct")
		}
	}
	rhc.query = queryOk

	_, bodyOk := handlerElem.FieldByName("Body")
	_, formOk := handlerElem.FieldByName("Form")

	if method == http.MethodGet {
		if bodyOk || formOk {
			return nil, errors.New("A GET request handler cannot have body or form")
		}
	}

	rhc.form = formOk
	rhc.body = bodyOk

	for i := 0; i < handlerElem.NumField(); i++ {
		name := handlerElem.Field(i).Name
		if name == "Body" || name == "Form" || name == "Query" || name == "Param" {
			continue
		} else {
			// TODO: Check if we can set it!, is the field exported?
			rhc.injection = true
			break
		}
	}

	rhc.RequestHandler = rhc.getMiddleRequestHandler(injector)
	return &rhc, nil
}

func transformWebsocketHandler(path string, injector *injector, handler WebsocketHandler) (*handlerContext, error) {
	hc := &handlerContext{
		websocket: true,
	}

	// Handler parse parameters
	handlerElem := reflect.TypeOf(handler).Elem()

	hc.tip = handlerElem

	// See if we have any params
	param, paramOk := handlerElem.FieldByName("Param")
	if paramOk {
		// If we have something param, it should be a struct only
		// TODO: Additional check, it should be flat struct
		// TODO: Additional check, it should be compatible with path
		if param.Type.Kind() != reflect.Struct {
			return nil, errors.New("Param field added but it is not a struct")
		}
	}
	hc.param = paramOk

	query, queryOk := handlerElem.FieldByName("Query")
	if queryOk {
		// If we have something param, it should be a struct only
		// TODO: Additional check, it should be flat struct
		if query.Type.Kind() != reflect.Struct {
			return nil, errors.New("Query field added but it is not a struct")
		}
	}

	hc.query = queryOk
	hc.RequestHandler = hc.getMiddleRequestHandler(injector)
	return hc, nil
}

func (hc *handlerContext) getMiddleRequestHandler(injector *injector) middleRequestHandler {
	// Create a new handler here
	fn := func(c *Context) error {
		obj := reflect.New(hc.tip)
		objElem := obj.Elem()

		if hc.param {
			err := c.parseParams(objElem)
			if err != nil {
				return err
			}
		}

		if hc.query {
			err := c.parseQuery(objElem)
			if err != nil {
				return err
			}
		}

		// ws is a get request, so it cannot have neither body or form
		if !hc.websocket {
			if hc.body {
				err := c.parseBody(objElem)
				if err != nil {
					return err
				}
			}

			if hc.form {
				err := c.parseForm(objElem)
				if err != nil {
					return err
				}
			}
		}

		if hc.injection {
			err := c.parseInjections(objElem, injector)
			if err != nil {
				return err
			}
		}

		if hc.websocket {
			wsHandler, ok := obj.Interface().(WebsocketHandler)
			if !ok {
				// It should, it cannot be here
				return errors.New("The interface does not implement WebsocketHandler: " + hc.tip.Name())
			}

			responseHeader, err := wsHandler.Before(c)
			if err != nil {
				return err
			}

			var upgrader = websocket.Upgrader{}

			conn, err := upgrader.Upgrade(c.w, c.r, responseHeader)
			if err != nil {
				return err
			}

			wsHandler.Handle(conn)
			return nil
		}

		// If it is not websocket, it is a HTTP Request
		reqHandler, ok := obj.Interface().(RequestHandler)
		if !ok {
			// It should, it cannot be here
			return errors.New("The interface does not implement RequestHandler: " + hc.tip.Name())
		}
		return reqHandler.Handle(c)
	}
	return fn
}
