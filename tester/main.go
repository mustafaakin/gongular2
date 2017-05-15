package main

import (
	"fmt"
	"log"

	"time"

	"github.com/gorilla/websocket"
	"github.com/mustafaakin/gongular2"
	"github.com/mustafaakin/gongular2/deneme"
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
	Param struct {
		UserID int8
	}
}

func (s *someOtherRequest) Handle(c *gongular2.Context) error {
	c.SetBody(fmt.Sprintf("hi %d", s.Param.UserID))
	return nil
}

type wsTest struct {
	Param struct {
		UserID int
	}
}

func (w *wsTest) Before(c *gongular2.Context) error {
	return nil
}

func (w *wsTest) Handle(conn *websocket.Conn) {
	conn.WriteMessage(websocket.TextMessage, []byte(fmt.Sprintf("Selam kullanıcı %d", w.Param.UserID)))
	go func() {
		for i := 0; i < 10; i++ {
			err := conn.WriteMessage(websocket.TextMessage, []byte(fmt.Sprintf("Selam kullanıcı %d", w.Param.UserID)))
			if err != nil {
				log.Println(err)
			}
			time.Sleep(time.Millisecond * 1000)
		}
	}()

	for {
		t, p, err := conn.ReadMessage()
		log.Println(t, string(p))
		if err != nil {
			break
		}
	}
}

func main() {
	// Create a new router
	e := gongular2.NewEngine()

	// Interfaces
	e.ProvideWithKey("m1", MyInt(45))
	e.ProvideWithKey("m2", MyInt(30))

	// HTTP Handlers
	r := e.GetRouter()
	r.POST("/test/:UserID/sayHi", &someRequest{})
	r.GET("/user/:UserID", &someOtherRequest{})
	r.GET("/deneme", &deneme.SelamMid{}, &deneme.SelamFn{})

	// WS Handlers
	w := e.GetWSRouter()
	w.Handle("/ws/:UserID/test", &wsTest{})

	log.Fatal(e.ListenAndServe(":8000"))
}
