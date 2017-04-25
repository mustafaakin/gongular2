package gongular2

import (
	"bytes"
	"log"
	"net/http"

	"path"

	"github.com/julienschmidt/httprouter"
)

const methodWS = "WEBSOCKET"

// Router holds the required states and does the mapping of requests
type Router struct {
	actualRouter *httprouter.Router
	errorHandler ErrorHandler
	injector     *injector
	prefix       string
	handlers     []interface{}
}

// NewRouter creates a new gongular2 Router
func NewRouter() *Router {
	r := Router{
		actualRouter: httprouter.New(),
		errorHandler: defaultErrorHandler,
		injector:     newInjector(),
		prefix:       "",
		handlers:     make([]interface{}, 0),
	}

	return &r
}

// WS is websocket handler
func (r *Router) WS(path string, handlers ...RequestHandler) {
	r.combineAndWrapHandlers(path, methodWS, handlers...)
}

// GET registers the given handlers at the path
func (r *Router) GET(path string, handlers ...RequestHandler) {
	r.combineAndWrapHandlers(path, http.MethodGet, handlers...)
}

// POST registers the given handlers at the path
func (r *Router) POST(path string, handlers ...RequestHandler) {
	r.combineAndWrapHandlers(path, http.MethodPost, handlers...)
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

// Group groups a given path with additional interfaces. It is useful to avoid
// repetitions while defining many paths
func (r *Router) Group(_path string, handlers ...RequestHandler) *Router {
	newRouter := &Router{
		actualRouter: r.actualRouter,
		injector:     r.injector,
		prefix:       path.Join(r.prefix, _path),
		errorHandler: r.errorHandler,
	}

	// Copy previous handlers references
	newRouter.handlers = make([]interface{}, len(r.handlers))
	copy(newRouter.handlers, r.handlers)

	// Append new handlers
	newRouter.handlers = append(newRouter.handlers, handlers...)

	return newRouter
}

// subpath initiates a new route with path and handlers, useful for grouping
func (r *Router) subpath(_path string, handlers []interface{}) (string, []interface{}) {
	combinedHandlers := r.handlers
	combinedHandlers = append(combinedHandlers, handlers...)

	resultingPath := path.Join(r.prefix, _path)
	return resultingPath, combinedHandlers
}

func (r *Router) combineAndWrapHandlers(path, method string, handlers ...interface{}) {
	resultingPath, combinedHandlers := r.subpath(path, handlers)

	fn := r.transformHandlers(resultingPath, method, combinedHandlers)

	switch method {
	case methodWS:
		r.actualRouter.GET(resultingPath, fn)
	case http.MethodGet:
		r.actualRouter.GET(resultingPath, fn)
	case http.MethodPost:
		r.actualRouter.POST(resultingPath, fn)
	case http.MethodPut:
		r.actualRouter.PUT(resultingPath, fn)
	case http.MethodHead:
		r.actualRouter.HEAD(resultingPath, fn)
	}
}

func (r *Router) transformHandlers(path string, method string, handlers []interface{}) httprouter.Handle {
	middleHandlers := make([]*handlerContext, len(handlers))

	var isWebsocket bool = method == methodWS

	for i, handler := range handlers {
		mh, err := transformHandler(path, method, isWebsocket, r.injector, handler)
		if err != nil {
			log.Fatal(err)
		}
		middleHandlers[i] = mh
	}

	var fn httprouter.Handle
	fn = func(wr http.ResponseWriter, req *http.Request, ps httprouter.Params) {
		// Create a logger for each request so that we can group the output
		buf := new(bytes.Buffer)
		logger := log.New(buf, "", log.LstdFlags)

		// Create a context that wraps the request, writer and logger
		ctx := contextFromRequest(path, wr, req, ps, logger)

		// For each of the handler this route has, try to execute it
		var requestErr error
		for _, handler := range middleHandlers {
			// Parse the parameters to the handler object
			fn := handler.requestHandler
			err := fn(ctx)

			// If an error occurs, stop the chain
			if err != nil {
				ctx.StopChain()
				r.errorHandler(err, ctx)
				requestErr = err
				break
			}

			if ctx.stopChain {
				break
			}
		}

		if !isWebsocket {
			// Finalize the request in the end
			ctx.Finalize()
		} else {
			// If an error has occurred, we should also try to finalize
			if requestErr != nil {
				ctx.Finalize()
			}
		}
	}

	return fn
}

func (r *Router) Provide(value interface{}) {
	r.injector.Provide(value, "default")
}

func (r *Router) ProvideWithKey(key string, value interface{}) {
	r.injector.Provide(value, key)
}
