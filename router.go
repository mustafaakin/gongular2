package gongular2

import (
	"encoding/json"
	"log"
	"net/http"
	"reflect"

	"github.com/julienschmidt/httprouter"
)

// Router holds the required states and does the mapping of requests
type Router struct {
	actualRouter *httprouter.Router
}

// NewRouter creates a new gongular2 Router
func NewRouter() *Router {
	r := Router{
		actualRouter: httprouter.New(),
	}

	return &r
}

// GET registers the given handlers at the path
func (r *Router) GET(path string, handlers ...Handler) {
	// TODO: Add recover here
	r.actualRouter.GET(path, r.transformHandler(path, handlers[0]))
}

// POST registers the given handlers at the path
func (r *Router) POST(path string, handlers ...Handler) {
	// TODO: Add recover here
	r.actualRouter.POST(path, r.transformHandler(path, handlers[0]))
}

// ServeFiles serves the files
func (r *Router) ServeFiles(path string, root http.FileSystem) {
	r.actualRouter.ServeFiles(path, root)
}

// ServeHTTP serves from http
func (r *Router) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	r.actualRouter.ServeHTTP(w, req)
}

// ListenAndServe is
func (r *Router) ListenAndServe(addr string) error {
	return http.ListenAndServe(addr, r.actualRouter)
}

func (r *Router) transformHandler(path string, handler Handler) httprouter.Handle {
	// Handler parse parameters
	handlerType := reflect.TypeOf(handler)
	handlerElem := handlerType.Elem()

	// See if we have any params
	param, paramOk := handlerElem.FieldByName("Param")
	if paramOk {
		// If we have something param, it should be a struct only
		// TODO: Additional check, it should be flat struct
		// TODO: Additional check, it should be compatible with path
		if param.Type.Kind() != reflect.Struct {
			panic("Param field added but it is not a struct")
		}
	}

	query, queryOk := handlerElem.FieldByName("Query")
	if queryOk {
		// If we have something param, it should be a struct only
		// TODO: Additional check, it should be flat struct
		// TODO: Additional check, it should be compatible with path
		if query.Type.Kind() != reflect.Struct {
			panic("Query field added but it is not a struct")
		}
	}

	_, bodyOk := handlerElem.FieldByName("Body")

	// TODO: In the future, analyze all the handlers and decide if we have any shared stuff (params, dependencies etc.) so we do not re-init them
	return func(w http.ResponseWriter, req *http.Request, ps httprouter.Params) {
		// TODO: Modify it to handle more than one handler later
		ctx := contextFromRequest(w, req, nil)
		_ = ctx
		// Create a new handler here

		newHandler := reflect.New(handlerElem)
		newHandlerElem := newHandler.Elem()
		if paramOk {
			param := newHandlerElem.FieldByName("Param")
			paramType := param.Type()

			numFields := paramType.NumField()
			for i := 0; i < numFields; i++ {
				field := paramType.Field(i)
				s := ps.ByName(field.Name)
				param.Field(i).SetString(s)
			}
		}

		if queryOk {

		}

		if bodyOk {
			body := newHandlerElem.FieldByName("Body")
			b := body.Addr().Interface()

			err := json.NewDecoder(req.Body).Decode(b)
			if err != nil {
				log.Println("could not decode", err)
			}
		}

		// Convert it to an interface and call its handle method which may modify the context
		handlerInterface := newHandler.Interface().(Handler)
		handlerInterface.Handle(ctx)

		json.NewEncoder(w).Encode(ctx.body)
	}
}
