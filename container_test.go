package di

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

type MyComponent1 struct {
}

type MyComponent2 struct {
}

type Container struct {
	Component1 func() *MyComponent1
	Component2 func() *MyComponent2
}

func newContainer() *Container {
	c := new(Container)
	c.Component1 = NewComponent(func() *MyComponent1 {
		return new(MyComponent1)
	})
	c.Component2 = NewComponent(func() *MyComponent2 {
		return new(MyComponent2)
	})
	return c
}

type MyEngine struct {
	Component1 *MyComponent1
	Component2 *MyComponent2
}

func (e *MyEngine) Run() {
	// Place your code here
}

func TestResolve(t *testing.T) {
	engine := new(MyEngine)
	container := newContainer()
	Resolve(container, engine)
	assert.NotNil(t, engine.Component1)
	assert.NotNil(t, engine.Component2)
	engine.Run()
}
