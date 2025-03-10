package di

import (
	"context"
	"fmt"
	"sync"
)

type componentError struct {
	component string
	message   string
}

func (e *componentError) Error() string {
	return fmt.Sprintf("%s: %s", e.component, e.message)
}

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

func (that *component) runInit(ctx context.Context, app *App) error {
	if that.state != StateBuild {
		return nil
	}
	that.state = StateInit
	if that.init == nil {
		return nil
	}
	err := that.init(ctx)
	if err != nil {
		return fmt.Errorf("initialization error %w", err)
	}
	app.logger.Debugf(ctx, "Component %s initialized", that.name)
	return nil
}

func (that *component) runDone(ctx context.Context, app *App) {
	if that.state != StateInit {
		return
	}
	that.state = StateDone
	if that.done == nil {
		return
	}
	that.done(ctx)
	app.logger.Debugf(ctx, "Component %s done", that.name)
}

type Initializer interface {
	Init() error
}

type Finalizer interface {
	Done()
}

type Option[T any] func(options *Options[T])

type Options[T any] struct {
	name string
	init []func(ctx context.Context, instance T) error
	done []func(ctx context.Context, instance T)
}

// WithComponentInit adds initializers to the component
func WithComponentInit[T any](initializer ...func(ctx context.Context, instance T) error) Option[T] {
	return func(options *Options[T]) {
		options.init = append(options.init, initializer...)
	}
}

// WithComponentDone adds finalizers to the component
func WithComponentDone[T any](finalizer ...func(ctx context.Context, instance T)) Option[T] {
	return func(options *Options[T]) {
		options.done = append(options.done, finalizer...)
	}
}

// WithComponentNativeInit adds initializer to the component
func WithComponentNativeInit[T Initializer]() Option[T] {
	return func(options *Options[T]) {
		fn := func(ctx context.Context, instance T) error {
			return instance.Init()
		}
		options.init = append(options.init, fn)
	}
}

// WithComponentNativeDone adds finalizer to the component
func WithComponentNativeDone[T Finalizer]() Option[T] {
	return func(options *Options[T]) {
		fn := func(ctx context.Context, instance T) {
			instance.Done()
		}
		options.done = append(options.done, fn)
	}
}

// NewComponent makes new component by wrapping it by builder func
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

func (that *controller[T]) enter() {
	if that.active {
		panic(&componentError{that.options.name, "circular dependency detected"})
	}
	that.active = true
}

func (that *controller[T]) leave() {
	that.active = false
}

func (that *controller[T]) get(ctx context.Context) T {
	that.enter()
	defer that.leave()

	app := GetAppFromContext(ctx)
	cc := app.get(ctx, that.id, that.newComponent)
	return cc.instance.(T)
}

func (that *controller[T]) newComponent(ctx context.Context, app *App) *component {
	app.logger.Debugf(ctx, "Component %s building", that.options.name)
	instance := that.newInstance(ctx)
	return &component{
		id:       that.id,
		name:     that.options.name,
		instance: instance,
		init: func(ctx context.Context) error {
			for _, init := range that.options.init {
				err := init(ctx, instance)
				if err != nil {
					return err
				}
			}
			return nil
		},
		done: func(ctx context.Context) {
			for _, done := range that.options.done {
				done(ctx, instance)
			}
		},
	}
}

func (that *controller[T]) newInstance(ctx context.Context) T {
	instance, err := that.builder(ctx)
	if err != nil {
		panic(&componentError{that.options.name, err.Error()})
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
