package di

import (
	"context"
	"fmt"
	"reflect"
)

type Builder[T any] func(ctx context.Context) (T, error)
type Constructor[T any] func(ctx context.Context) T

type State int

const (
	StateBuild State = iota
	StateInit
	StateDone
)

type component struct {
	state    State
	init     func(ctx context.Context) error
	done     func(ctx context.Context)
	priority int
	name     string
}

func (c *component) runInit(ctx context.Context) error {
	if c.state != StateBuild {
		return nil
	}
	c.state = StateInit
	if c.init == nil {
		return nil
	}
	err := c.init(ctx)
	if err != nil {
		return fmt.Errorf("initialization error %w", err)
	}
	Logger.Log(c.name, "initialized")
	return nil
}

func (c *component) runDone(ctx context.Context) {
	if c.state != StateInit {
		return
	}
	c.state = StateDone
	if c.done == nil {
		return
	}
	c.done(ctx)
	Logger.Log(c.name, "done")
}

type logger struct{}

func (l *logger) Log(component, msg string) {
	// empty
}

var Logger interface {
	Log(component, msg string)
} = new(logger)

type Initializer interface {
	Init() error
}

type Finalizer interface {
	Done()
}

type Option[T any] func(options *Options[T])

type Options[T any] struct {
	name     string
	priority int
	init     func(ctx context.Context, instance T) error
	done     func(ctx context.Context, instance T)
}

func WithPriority[T any](priority int) Option[T] {
	return func(options *Options[T]) {
		options.priority = priority
	}
}

func WithInit[T any](initializer func(ctx context.Context, instance T) error) Option[T] {
	return func(options *Options[T]) {
		options.init = initializer
	}
}

func WithDone[T any](finalizer func(ctx context.Context, instance T)) Option[T] {
	return func(options *Options[T]) {
		options.done = finalizer
	}
}

func UseInit[T Initializer]() Option[T] {
	return func(options *Options[T]) {
		options.init = func(ctx context.Context, instance T) error {
			return instance.Init()
		}
	}
}

func UseDone[T Finalizer]() Option[T] {
	return func(options *Options[T]) {
		options.done = func(ctx context.Context, instance T) {
			instance.Done()
		}
	}
}

func NewComponent[T any](
	name string,
	builder Builder[T],
	options ...Option[T],
) Constructor[T] {
	c := controller[T]{
		options: Options[T]{
			name: name,
		},
		builder: builder,
	}

	for _, o := range options {
		o(&c.options)
	}

	return c.get
}

type controller[T any] struct {
	builder  Builder[T]
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

func (c *controller[T]) get(ctx context.Context) T {
	c.enter()
	defer c.leave()

	if isZeroVal(c.instance) {
		c.instance = c.newInstance(ctx)
		app := GetAppFromContext(ctx)
		app.addComponent(c.newComponent())
	}

	return c.instance
}

func (c *controller[T]) newInstance(ctx context.Context) T {
	Logger.Log(c.options.name, "building")
	instance, err := c.builder(ctx)
	if err != nil {
		panic(fmt.Errorf("%s: %w", c.options.name, err))
	}
	return instance
}

func (c *controller[T]) newComponent() *component {
	return &component{
		name:     c.options.name,
		priority: c.options.priority,
		init: func(ctx context.Context) error {
			if c.options.init != nil {
				return c.options.init(ctx, c.instance)
			}
			return nil
		},
		done: func(ctx context.Context) {
			if c.options.done != nil {
				c.options.done(ctx, c.instance)
			}
		},
	}
}

// IsZeroVal check if any type is its zero value
func isZeroVal(x interface{}) bool {
	return x == nil || reflect.DeepEqual(x, reflect.Zero(reflect.TypeOf(x)).Interface())
}
