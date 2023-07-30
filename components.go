package di

import (
	"fmt"
	"reflect"
)

type Initializer interface {
	Init() error
}

type Finalizer interface {
	Done()
}

type Option[T any] func(options *Options[T])

type Options[T any] struct {
	init func(instance T) error
	done func(instance T)
}

func WithInit[T any](initializer func(instance T) error) Option[T] {
	return func(options *Options[T]) {
		options.init = initializer
	}
}

func WithDone[T any](finalizer func(instance T)) Option[T] {
	return func(options *Options[T]) {
		options.done = finalizer
	}
}

func WithMethodInit[T Initializer]() Option[T] {
	return func(options *Options[T]) {
		options.init = func(instance T) error {
			return instance.Init()
		}
	}
}

func WithMethodDone[T Finalizer]() Option[T] {
	return func(options *Options[T]) {
		options.done = func(instance T) {
			instance.Done()
		}
	}
}

func NewComponent[T any](
	builder func() (T, error),
	options ...Option[T],
) func() T {
	c := controller[T]{
		builder: builder,
	}

	for _, o := range options {
		o(&c.options)
	}

	return c.get
}

type controller[T any] struct {
	builder  func() (T, error)
	options  Options[T]
	active   bool
	instance T
}

func (c *controller[T]) enter() {
	if c.active {
		panic(fmt.Errorf("circular dependency detected"))
	}
	c.active = true
}

func (c *controller[T]) leave() {
	c.active = false
}

func (c *controller[T]) get() T {
	c.enter()
	defer c.leave()

	if isZeroVal(c.instance) {
		c.instance = c.newInstance()
		application.addComponent(c.newComponent())
	}

	return c.instance
}

func (c *controller[T]) newInstance() T {
	instance, err := c.builder()
	if err != nil {
		panic(err)
	}
	return instance
}

func (c *controller[T]) newComponent() *component {
	return &component{
		init: func() error {
			if c.options.init != nil {
				return c.options.init(c.instance)
			}
			return nil
		},
		done: func() {
			if c.options.done != nil {
				c.options.done(c.instance)
			}
		},
	}
}

// IsZeroVal check if any type is its zero value
func isZeroVal(x interface{}) bool {
	return x == nil || reflect.DeepEqual(x, reflect.Zero(reflect.TypeOf(x)).Interface())
}
