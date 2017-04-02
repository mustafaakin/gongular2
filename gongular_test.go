package gongular2

import (
	"fmt"
	"testing"
)

type someRequest struct {
	param struct {
		UserID string
	}
	body struct {
		Name string
		Age  int
	}
}

func (s *someRequest) Handle(c *Context) error {
	response := fmt.Sprintf("Hello user with ID %s, %s, you are %d years old", s.param.UserID, s.body.Name, s.body.Age)
	c.SetBody(response)
	return nil
}

func TestBody(t *testing.T) {
	r := NewRouter()
	r.GET("/body", &someRequest{})
}
