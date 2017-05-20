package gongular2

import (
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
)

type simpleHandler struct{}

func (s *simpleHandler) Handle(c *Context) error {
	c.SetBody("selam")
	return nil
}

func TestSimpleGetHandler(t *testing.T) {
	e := newEngineTest()
	e.GetRouter().GET("/", &simpleHandler{})

	resp, content := get(t, e, "/")

	assert.Equal(t, http.StatusOK, resp.Code)
	assert.Equal(t, "\"selam\"", content)
}

func TestSimplePostHandler(t *testing.T) {
	e := newEngineTest()
	e.GetRouter().POST("/", &simpleHandler{})

	resp, content := post(t, e, "/", nil)
	assert.Equal(t, http.StatusOK, resp.Code)
	assert.Equal(t, "\"selam\"", content)
}

type statusSetHandler struct{}

func (s *statusSetHandler) Handle(c *Context) error {
	c.Status(http.StatusExpectationFailed)
	return nil
}

func TestSetStatus(t *testing.T) {
	e := newEngineTest()
	e.GetRouter().GET("/", &statusSetHandler{})

	resp, _ := get(t, e, "/")
	assert.Equal(t, http.StatusExpectationFailed, resp.Code)
}
