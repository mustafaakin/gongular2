package gongular2

import (
	"bytes"
	"log"
	"net/http"

	"github.com/julienschmidt/httprouter"
)

// Router holds the required states and does the mapping of requests
type Router struct {
	actualRouter *httprouter.Router
	errorHandler ErrorHandler
}

// NewRouter creates a new gongular2 Router
func NewRouter() *Router {
	r := Router{
		actualRouter: httprouter.New(),
		errorHandler: defaultErrorHandler,
	}

	return &r
}

// GET registers the given handlers at the path
func (r *Router) GET(path string, handlers ...RequestHandler) {
	// TODO: Add recover here
	r.actualRouter.GET(path, r.transformHandlers(path, handlers))
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

func (r *Router) transformHandlers(path string, handlers []RequestHandler) httprouter.Handle {
	middleHandlers := make([]*handlerContext, len(handlers))

	for i, handler := range handlers {
		mh, _ := transformHandler(path, handler)
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

		// Finalize the reuqest in the end
		ctx.Finalize()
	}

	return fn
}
