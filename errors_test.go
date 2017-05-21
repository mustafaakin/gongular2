package gongular2

import (
	"net/http"
	"testing"

	"database/sql"

	"github.com/stretchr/testify/assert"
)

type expectingIntParam struct {
	Param struct {
		UserID int
	}
}

func (e *expectingIntParam) Handle(c *Context) error {
	c.SetBody("WOW")
	return nil
}

func TestIncompatibleTypes(t *testing.T) {
	e := newEngineTest()

	e.GetRouter().GET("/hey/:UserID", &expectingIntParam{})

	resp1, content1 := get(t, e, "/hey/5")

	assert.Equal(t, http.StatusOK, resp1.Code)
	assert.Equal(t, `"WOW"`, content1)

	resp2, content2 := get(t, e, "/hey/notReallyAInteger")

	assert.Equal(t, http.StatusBadRequest, resp2.Code)
	assert.NotEqual(t, `"WOW"`, content2)
}

type expectingAlphaNumParam struct {
	Param struct {
		UserID string `valid:"alphanum"`
	}
}

func (e *expectingAlphaNumParam) Handle(c *Context) error {
	c.SetBody("WOW")
	return nil
}

func TestValidation(t *testing.T) {
	e := newEngineTest()

	e.GetRouter().GET("/hey/:UserID", &expectingAlphaNumParam{})

	resp1, content1 := get(t, e, "/hey/abc300")

	assert.Equal(t, http.StatusOK, resp1.Code)
	assert.Equal(t, `"WOW"`, content1)

	resp2, content2 := get(t, e, "/hey/abc$")

	assert.Equal(t, http.StatusBadRequest, resp2.Code)
	assert.NotEqual(t, `"WOW"`, content2)
}

type failingInjection struct {
	DB *sql.DB `inject:"primary"`
}

func (e *failingInjection) Handle(c *Context) error {
	c.SetBody("WOW")
	return nil
}

func TestFailingInjection(t *testing.T) {
	e := newEngineTest()

	e.CustomProvideWithKey("primary", &sql.DB{}, func(c *Context) (interface{}, error) {
		return nil, sql.ErrTxDone
	})

	e.GetRouter().GET("/hey", &failingInjection{})

	resp, content := get(t, e, "/hey")

	assert.Equal(t, http.StatusInternalServerError, resp.Code)
	assert.NotEqual(t, `"WOW"`, content)
}

type justFail struct{}

func (e *justFail) Handle(c *Context) error {
	c.SetBody("WOW")
	return http.ErrLineTooLong
}

func TestFailing(t *testing.T) {
	e := newEngineTest()

	e.GetRouter().GET("/uuu", &justFail{})

	resp, content := get(t, e, "/uuu")

	assert.Equal(t, http.StatusInternalServerError, resp.Code)
	assert.NotEqual(t, `"WOW"`, content)
}
