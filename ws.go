package gongular2

import "github.com/gorilla/websocket"

type WebsocketHandler interface {
	// Before is a filter applied just before upgrading the request to websocket
	// It can be useful for filtering the request and returning an error would not open a websocket but close it with an error
	Before(c *Context) error
	// Handle is regular handling of the web socket, user is fully responsible for the request
	Handle(conn *websocket.Conn)
}
