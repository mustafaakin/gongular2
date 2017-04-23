package main

import (
	"fmt"
	"log"

	"github.com/mustafaakin/gongular2"
)

type MyInt int

type someRequest struct {
	Param struct {
		UserID string
	}
	Body struct {
		Name string
		Age  int
	}
	M1 MyInt `inject:"m1"`
	M2 MyInt `inject:"m2"`
}

func (s *someRequest) Handle(c *gongular2.Context) error {
	response := fmt.Sprintf("Hello user with ID %s, %s, you are %d years old, and M1: %d  M2: %d",
		s.Param.UserID, s.Body.Name, s.Body.Age, s.M1, s.M2)

	c.SetBody(response)
	return nil
}

type someOtherRequest struct {
}

func (s *someOtherRequest) Handle(c *gongular2.Context) error {
	c.SetBody("hi")
	return nil
}

func main() {
	r := gongular2.NewRouter()

	r.ProvideWithKey("m1", MyInt(45))
	r.ProvideWithKey("m2", MyInt(30))

	r.POST("/test/:UserID/sayHi", &someRequest{})
	r.GET("/aa", &someOtherRequest{})
	g1 := r.Group("/deneme")
	g1.GET("/bb", &someOtherRequest{})

	log.Println("starting")
	log.Fatal(r.ListenAndServe(":8080"))
}
