package gongular2

// Handler is a generic handler for gongular2
type Handler interface {
	Handle(c *Context) error
}
