package deneme

import (
	"fmt"

	"github.com/mustafaakin/gongular2"
)

type SelamFn struct {
}

func (a *SelamFn) Handle(c *gongular2.Context) error {
	c.SetBody("selam")
	return nil
}

type SelamMid struct {
}

func (a *SelamMid) Handle(c *gongular2.Context) error {
	fmt.Println("im mid")
	return nil
}
