package container_test

import (
	"fmt"
	"testing"

	"github.com/studiolambda/cosmos/container"
)

type Name string

type Stuff struct {
	n    int
	p    string
	name Name
}

func NewStuff(n int, p string, name Name) Stuff {
	return Stuff{n, p, name}
}

func (s Stuff) String() string {
	return fmt.Sprintf("%s %s %d", s.p, s.name, s.n)
}

func TestExample(t *testing.T) {
	c := container.NewContainer()

	c.Register(func(c *container.Container) (int, error) {
		return 10, nil
	})

	c.Register(func(c *container.Container) (string, error) {
		return "hello", nil
	})

	c.Register(func(c *container.Container) (Name, error) {
		return "world", nil
	})

	r, err := c.Make[Stuff](NewStuff)

	if err != nil {
		t.FailNow()
	}

	t.Log(r.String())
}
