package di

import (
	"context"
	"fmt"
	"sync"
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
	instance interface{}
	priority int
	name     string
	id       string
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
	init     []func(ctx context.Context, instance T) error
	done     []func(ctx context.Context, instance T)
}

func WithPriority[T any](priority int) Option[T] {
	return func(options *Options[T]) {
		options.priority = priority
	}
}

func WithInit[T any](initializer ...func(ctx context.Context, instance T) error) Option[T] {
	return func(options *Options[T]) {
		options.init = append(options.init, initializer...)
	}
}

func WithDone[T any](finalizer ...func(ctx context.Context, instance T)) Option[T] {
	return func(options *Options[T]) {
		options.done = append(options.done, finalizer...)
	}
}

func UseInit[T Initializer]() Option[T] {
	return func(options *Options[T]) {
		fn := func(ctx context.Context, instance T) error {
			return instance.Init()
		}
		options.init = append(options.init, fn)
	}
}

func UseDone[T Finalizer]() Option[T] {
	return func(options *Options[T]) {
		fn := func(ctx context.Context, instance T) {
			instance.Done()
		}
		options.done = append(options.done, fn)
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
		id:      enumerator(),
	}

	for _, o := range options {
		o(&c.options)
	}

	return c.get
}

type controller[T any] struct {
	builder Builder[T]
	options Options[T]
	active  bool
	id      string
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

	app := GetAppFromContext(ctx)
	cc := app.get(ctx, c.id, c.newComponent)
	return cc.instance.(T)
}

func (c *controller[T]) newComponent(ctx context.Context) *component {
	instance := c.newInstance(ctx)
	return &component{
		id:       c.id,
		name:     c.options.name,
		priority: c.options.priority,
		instance: instance,
		init: func(ctx context.Context) error {
			for _, init := range c.options.init {
				err := init(ctx, instance)
				if err != nil {
					return fmt.Errorf("initialization error %w", err)
				}
			}
			return nil
		},
		done: func(ctx context.Context) {
			for _, done := range c.options.done {
				done(ctx, instance)
			}
		},
	}
}

func (c *controller[T]) newInstance(ctx context.Context) T {
	Logger.Log(c.options.name, "building")
	instance, err := c.builder(ctx)
	if err != nil {
		panic(fmt.Errorf("%s: %w", c.options.name, err))
	}
	return instance
}

var enumerator = func() func() string {
	var mx sync.Mutex
	var counter int

	return func() string {
		mx.Lock()
		defer mx.Unlock()

		counter++
		return fmt.Sprintf("%d", counter)
	}
}()
