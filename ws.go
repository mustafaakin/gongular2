package gongular2

import (
	"bytes"
	"log"
	"net/http"

	"github.com/gorilla/websocket"
	"github.com/julienschmidt/httprouter"
)

type WebsocketHandler interface {
	// Before is a filter applied just before upgrading the request to websocket
	// It can be useful for filtering the request and returning an error would not open a websocket but close it with an error
	Before(c *Context) error
	// Handle is regular handling of the web socket, user is fully responsible for the request
	Handle(conn *websocket.Conn)
}

type WSRouter struct {
	engine *Engine
}

func newWSRouter(e *Engine) *WSRouter {
	return &WSRouter{
		engine: e,
	}
}

func (r *WSRouter) Handle(path string, handler WebsocketHandler) {
	mh, err := transformWebsocketHandler(path, r.engine.injector, handler)
	if err != nil {
		log.Fatal(err)
	}

	var fn httprouter.Handle
	fn = func(wr http.ResponseWriter, req *http.Request, ps httprouter.Params) {
		// Create a logger for each request so that we can group the output
		buf := new(bytes.Buffer)
		logger := log.New(buf, "", log.LstdFlags)

		// Create a context that wraps the request, writer and logger
		ctx := contextFromRequest(path, wr, req, ps, logger)

		// Parse the parameters to the handler object
		fn := mh.RequestHandler
		fn(ctx)
	}

	r.engine.actualRouter.GET(path, fn)
}
