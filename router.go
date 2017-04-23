package gongular2

import (
	"bytes"
	"log"
	"net/http"

	"path"

	"github.com/julienschmidt/httprouter"
)

// Router holds the required states and does the mapping of requests
type Router struct {
	actualRouter *httprouter.Router
	errorHandler ErrorHandler
	injector     *injector
	prefix       string
	handlers     []RequestHandler
}

// NewRouter creates a new gongular2 Router
func NewRouter() *Router {
	r := Router{
		actualRouter: httprouter.New(),
		errorHandler: defaultErrorHandler,
		injector:     newInjector(),
		prefix:       "",
		handlers:     make([]RequestHandler, 0),
	}

	return &r
}

// GET registers the given handlers at the path
func (r *Router) GET(path string, handlers ...RequestHandler) {
	// TODO: Add recover here
	r.combineAndWrapHandlers(path, http.MethodGet, handlers...)
}

// POST registers the given handlers at the path
func (r *Router) POST(path string, handlers ...RequestHandler) {
	// TODO: Add recover here
	r.actualRouter.POST(path, r.transformHandlers(path, handlers))
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
	newRouter.handlers = make([]RequestHandler, len(r.handlers))
	copy(newRouter.handlers, r.handlers)

	// Append new handlers
	newRouter.handlers = append(newRouter.handlers, handlers...)

	return newRouter
}

// subpath initiates a new route with path and handlers, useful for grouping
func (r *Router) subpath(_path string, handlers []RequestHandler) (string, []RequestHandler) {
	combinedHandlers := r.handlers
	combinedHandlers = append(combinedHandlers, handlers...)

	resultingPath := path.Join(r.prefix, _path)
	return resultingPath, combinedHandlers
}

func (r *Router) combineAndWrapHandlers(path, method string, handlers ...RequestHandler) {
	resultingPath, combinedHandlers := r.subpath(path, handlers)
	fn := r.transformHandlers(resultingPath, combinedHandlers)

	r.actualRouter.GET(resultingPath, fn)
}

func (r *Router) transformHandlers(path string, handlers []RequestHandler) httprouter.Handle {
	middleHandlers := make([]*handlerContext, len(handlers))

	for i, handler := range handlers {
		mh, _ := transformHandler(path, r.injector, handler)
		// TODO: Error handle
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
		for _, handler := range middleHandlers {
			// Parse the parameters to the handler object
			fn := handler.requestHandler
			err := fn(ctx)

			// If an error occurs, stop the chain
			if err != nil {
				ctx.StopChain()
				r.errorHandler(err, ctx)
				break
			}

			if ctx.stopChain {
				break
			}
		}

		// Finalize the request in the end
		ctx.Finalize()
	}

	return fn
}

func (r *Router) Provide(value interface{}) {
	r.injector.Provide(value, "default")
}

func (r *Router) ProvideWithKey(key string, value interface{}) {
	r.injector.Provide(value, key)
}
