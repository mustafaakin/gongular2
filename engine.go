package gongular2

import (
	"net/http"

	"github.com/julienschmidt/httprouter"
)

type Engine struct {
	// The underlying router
	actualRouter *httprouter.Router

	// Injector
	injector *injector

	// HTTP Router
	httpRouter *Router
	// WS Router
	wsRouter *WSRouter
}

func NewEngine() *Engine {
	e := &Engine{
		actualRouter: httprouter.New(),
		injector:     newInjector(),
	}

	e.httpRouter = newRouter(e)
	e.wsRouter = newWSRouter(e)
	return e
}

func (e *Engine) GetRouter() *Router {
	return e.httpRouter
}

func (e *Engine) GetWSRouter() *WSRouter {
	return e.wsRouter
}

// ServeFiles serves the static files
func (e *Engine) ServeFiles(path string, root http.FileSystem) {
	e.actualRouter.ServeFiles(path, root)
}

// ServeHTTP serves from http
func (e *Engine) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	e.actualRouter.ServeHTTP(w, req)
}

// ListenAndServe
func (e *Engine) ListenAndServe(addr string) error {
	return http.ListenAndServe(addr, e.actualRouter)
}

// Provide provides with "default" key
func (e *Engine) Provide(value interface{}) {
	e.injector.Provide(value, "default")
}

// ProvideWithKey provides an interface with a key
func (e *Engine) ProvideWithKey(key string, value interface{}) {
	e.injector.Provide(value, key)
}
