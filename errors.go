package gongular2

import "net/http"

var defaultErrorHandler = func(err error, c *Context) {
	c.logger.Println("An error has occured:", err)
	c.SetBody(err.Error())
	c.MustStatus(http.StatusInternalServerError)
	c.StopChain()
}

// ErrorHandler is generic interface for error handling
type ErrorHandler func(err error, c *Context)
