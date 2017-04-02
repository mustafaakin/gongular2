package main

import (
	"fmt"
	"log"

	"github.com/jmoiron/sqlx"
	"github.com/mustafaakin/gongular2"
)

type someRequest struct {
	Param struct {
		UserID string
	}
	Body struct {
		Name string
		Age  int
	}
	db *sqlx.DB
}

func (s *someRequest) Handle(c *gongular2.Context) error {
	response := fmt.Sprintf("Hello user with ID %s, %s, you are %d years old", s.Param.UserID, s.Body.Name, s.Body.Age)
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
	r.POST("/test/:UserID/sayHi", &someRequest{})
	r.GET("/aa", &someOtherRequest{})

	log.Println("starting")
	log.Fatal(r.ListenAndServe(":8080"))
}
